package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/wjbbig/fabric-distributed-tool/tools/configtxgen/genesisconfig"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wjbbig/fabric-distributed-tool/connectionprofile"
	docker_compose "github.com/wjbbig/fabric-distributed-tool/docker-compose"
	"github.com/wjbbig/fabric-distributed-tool/fabricconfig"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/network"
	"github.com/wjbbig/fabric-distributed-tool/sdkutil"
	"github.com/wjbbig/fabric-distributed-tool/sshutil"
	"github.com/wjbbig/fabric-distributed-tool/utils"
)

// utils for network cmd
var logger = mylogger.NewLogger()

func GenerateCryptoConfig(dataDir string, networkConfig *network.NetworkConfig) error {
	if err := fabricconfig.GenerateCryptoConfigFile(dataDir, networkConfig.GetPeerNodes(), networkConfig.GetOrdererNodes()); err != nil {
		return err
	}
	return nil
}

func GenerateNetwork(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc string,
	ccInitParams []string, ccPolicy string, ccInitRequired bool, sequence int64, couchdb bool, peerUrls,
	ordererUrls []string, networkVersion string) (*network.NetworkConfig, error) {
	return network.GenerateNetworkConfig(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParams, ccPolicy, ccInitRequired, sequence, couchdb, peerUrls, ordererUrls, networkVersion)
}

func GenerateConfigtx(dataDir, consensus, channelId, fversion string, networkConfig *network.NetworkConfig) error {
	var err error
	var entities []*fabricconfig.ConfigtxApplicationChannelEntity
	// use existing network config file to generate configtx.yaml
	if channelId == "" {
		for channelName := range networkConfig.Channels {
			peerNodes, _, err := networkConfig.GetNodesByChannel(channelName)
			if err != nil {
				return err
			}
			if peerNodes == nil {
				return errors.New("at least set one peer node")
			}
			if networkConfig.Consensus == "" {
				consensus = "solo"
			}

			fversion = networkConfig.Version
			if networkConfig.Version == "" {
				fversion = fabricconfig.FabricVersion_V14
			}
			var orgs []string
			for _, node := range peerNodes {
				orgs = append(orgs, node.OrgId)
			}
			entities = append(entities, &fabricconfig.ConfigtxApplicationChannelEntity{ChannelId: channelName, Peers: orgs})
		}

		if err := fabricconfig.GenerateConfigtxFile(dataDir, consensus, networkConfig.GetPeerNodes(),
			networkConfig.GetOrdererNodes(), entities, fversion); err != nil {
			return err
		}

		return nil
	}

	// use the "network generate" command to generate configtx.yaml
	peerNodes, ordererNodes, err := networkConfig.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}

	if consensus == "" {
		if len(ordererNodes) == 1 {
			consensus = "solo"
		} else {
			consensus = "etcdraft"
		}
	}
	var orgs []string
	for _, node := range peerNodes {
		orgs = append(orgs, node.OrgId)
	}
	entities = append(entities, &fabricconfig.ConfigtxApplicationChannelEntity{ChannelId: channelId, Peers: orgs})
	if err := fabricconfig.GenerateConfigtxFile(dataDir, consensus, peerNodes, ordererNodes, entities, fversion); err != nil {
		return err
	}
	return nil
}

func GenerateDockerCompose(dataDir, networkName string, peerNodes, ordererNodes []*network.Node, imageTags []string, couchdb bool, ca bool) error {
	peersByOrg := make(map[string][]*network.Node)
	for _, peerNode := range peerNodes {
		peersByOrg[peerNode.OrgId] = append(peersByOrg[peerNode.OrgId], peerNode)
	}
	for _, peer := range peerNodes {
		var gossipUrl string
		orgPeers := peersByOrg[peer.OrgId]
		// if this org has only one peer, set this peer = gossip peer
		if len(orgPeers) == 1 {
			gossipUrl = fmt.Sprintf("%s:%d", peer.GetHostname(), peer.NodePort)
		} else {
			// this org has many peers. we choose one peer randomly, but exclude this peer
			for _, orgPeer := range orgPeers {
				if peer.GetHostname() == orgPeer.GetHostname() {
					continue
				}
				gossipUrl = fmt.Sprintf("%s:%d", orgPeer.GetHostname(), orgPeer.NodePort)
				break
			}
		}
		var extraHosts []string
		extraHosts = append(extraHosts, spliceHostnameAndIP(peer, peerNodes)...)
		extraHosts = append(extraHosts, spliceHostnameAndIP(peer, ordererNodes)...)
		if couchdb != peer.Couch {
			couchdb = peer.Couch
		}
		if err := docker_compose.GeneratePeerDockerComposeFile(dataDir, peer, gossipUrl, extraHosts, imageTags, couchdb); err != nil {
			return err
		}
	}
	for _, orderer := range ordererNodes {
		var extraHosts []string
		extraHosts = append(extraHosts, spliceHostnameAndIP(orderer, peerNodes)...)
		extraHosts = append(extraHosts, spliceHostnameAndIP(orderer, ordererNodes)...)
		if err := docker_compose.GenerateOrdererDockerComposeFile(dataDir, orderer, imageTags[0], extraHosts); err != nil {
			return err
		}
	}
	return nil
}

func spliceHostnameAndIP(excludeNode *network.Node, nodes []*network.Node) (extraHosts []string) {
	for _, node := range nodes {
		if node.GetHostname() == excludeNode.GetHostname() {
			continue
		}
		// if the ip is localhost or 127.0.0.1, the node will be abandoned
		isLocal, err := utils.CheckLocalIp(fmt.Sprintf("%s:%d", node.Host, node.NodePort))
		if err != nil {
			panic(fmt.Sprintf("IP address is wrong, err=%s", err))
		}
		if isLocal {
			continue
		}

		extraHosts = append(extraHosts, fmt.Sprintf("%s:%s", node.GetHostname(), node.Host))
	}
	return nil
}

func GenerateConnectionProfile(dataDir, channelId string, networkConfig *network.NetworkConfig) error {
	var entities []connectionprofile.ChannelEntity
	var channels []string
	if channelId != "" {
		channels = append(channels, channelId)
	} else {
		for channelName := range networkConfig.Channels {
			channels = append(channels, channelName)
		}
	}
	for _, channel := range channels {
		var entity connectionprofile.ChannelEntity
		peerNodes, ordererNodes, err := networkConfig.GetNodesByChannel(channel)
		if err != nil {
			return err
		}
		entity.ChannelId = channel
		for _, node := range peerNodes {
			entity.Peers = append(entity.Peers, fmt.Sprintf("%s:%d:%s", node.GetHostname(), node.NodePort, node.Host))
		}
		for _, node := range ordererNodes {
			entity.Orderers = append(entity.Orderers, fmt.Sprintf("%s:%d:%s", node.GetHostname(), node.NodePort, node.Host))
		}
		entities = append(entities, entity)
	}

	return connectionprofile.GenerateNetworkConnProfile(dataDir, entities)
}

func ReadSSHConfigFromNetwork(dataDir string, networkConfig *network.NetworkConfig) (*sshutil.SSHUtil, error) {
	sshUtil := sshutil.NewSSHUtil()
	for _, node := range networkConfig.GetPeerNodes() {
		if err := sshUtil.Add(node.GetHostname(), node.Username, node.Password, fmt.Sprintf("%s:%d", node.Host, node.SSHPort),
			node.Type, dataDir, node.Dest, node.Couch); err != nil {
			return nil, err
		}
	}
	for _, node := range networkConfig.GetOrdererNodes() {
		if err := sshUtil.Add(node.GetHostname(), node.Username, node.Password, fmt.Sprintf("%s:%d", node.Host, node.SSHPort),
			node.Type, dataDir, node.Dest, false); err != nil {
			return nil, err
		}
	}

	for _, node := range networkConfig.GetCANodes() {
		if err := sshUtil.Add(node.GetHostname(), node.Username, node.Password, fmt.Sprintf("%s:%d", node.Host, node.SSHPort),
			node.Type, dataDir, node.Dest, false); err != nil {
			return nil, err
		}
	}
	return sshUtil, nil
}

func TransferFilesByNodeName(sshUtil *sshutil.SSHUtil, dataDir string, nodes []string) error {
	for _, node := range nodes {
		client := sshUtil.GetClientByName(node)
		if client == nil {
			return errors.Errorf("node %s does not exist", node)
		}
		if err := transferFiles(client, node, dataDir); err != nil {
			return err
		}
	}
	return nil
}

func TransferFilesAllNodes(sshUtil *sshutil.SSHUtil, dataDir string, nc *network.NetworkConfig) error {
	for name, client := range sshUtil.Clients() {
		// the file of node has been transferred when building the other channel
		if nc.Nodes[name].Transferred {
			continue
		}
		if err := transferFiles(client, name, dataDir); err != nil {
			return err
		}
		nc.Nodes[name].Transferred = true
	}
	return nil
}

func transferFiles(sshClient *sshutil.SSHClient, nodeName, dataDir string) error {
	_, _, domain := utils.SplitNameOrgDomain(nodeName)
	// send node self keypairs and certs
	var certDir string
	var certDestDir string
	if sshClient.NodeType == network.PeerNode {
		peerCryptoConfigPrefix := filepath.Join(dataDir, "crypto-config", "peerOrganizations")
		peerCryptoConfigDestPrefix := filepath.Join(sshClient.DestPath, "crypto-config", "peerOrganizations")
		certDir = filepath.Join(peerCryptoConfigPrefix, domain)
		certDestDir = filepath.Join(peerCryptoConfigDestPrefix, domain)
	} else {
		ordererCryptoConfigPrefix := filepath.Join(dataDir, "crypto-config", "ordererOrganizations")
		ordererCryptoConfigDestPrefix := filepath.Join(sshClient.DestPath, "crypto-config", "ordererOrganizations")
		certDir = filepath.Join(ordererCryptoConfigPrefix, domain)
		certDestDir = filepath.Join(ordererCryptoConfigDestPrefix, domain)
	}
	err := sshClient.Sftp(certDir, certDestDir)
	if err != nil {
		return err
	}
	// send genesis.block, channel.tx and anchor.tx
	channelArtifactsSourcePath := filepath.Join(dataDir, "channel-artifacts")
	channelArtifactsDestPath := filepath.Join(sshClient.DestPath, "channel-artifacts")
	if err = sshClient.Sftp(channelArtifactsSourcePath, channelArtifactsDestPath); err != nil {
		return err
	}

	dockerComposeFilePath := filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(nodeName, ".", "-")))
	if err = sshClient.Sftp(dockerComposeFilePath, sshClient.DestPath); err != nil {
		return err
	}
	if sshClient.NeedCouch {
		dockerComposeFilePath = filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s-couchdb.yaml", strings.ReplaceAll(nodeName, ".", "-")))
		if err = sshClient.Sftp(dockerComposeFilePath, sshClient.DestPath); err != nil {
			return err
		}
	}
	return nil
}

func TransferNewChannelFiles(dataDir string, channelId string, sshUtil *sshutil.SSHUtil, nc *network.NetworkConfig) error {
	channelTxSourcePath := filepath.Join(dataDir, "channel-artifacts", fmt.Sprintf("%s.tx", channelId))
	peerNodes, ordererNodes, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	sshClients := sshUtil.Clients()
	for _, node := range peerNodes {
		client := sshClients[node.GetHostname()]
		channelTxDestPath := filepath.Join(client.DestPath, "channel-artifacts", fmt.Sprintf("%s.tx", channelId))
		if err := client.Sftp(channelTxSourcePath, channelTxDestPath); err != nil {
			return err
		}
	}
	for _, node := range ordererNodes {
		client := sshClients[node.GetHostname()]
		channelTxDestPath := filepath.Join(client.DestPath, "channel-artifacts", fmt.Sprintf("%s.tx", channelId))
		if err := client.Sftp(channelTxSourcePath, channelTxDestPath); err != nil {
			return err
		}
	}
	return nil
}

func StartupNetwork(sshUtil *sshutil.SSHUtil, nc *network.NetworkConfig) error {
	for name, client := range sshUtil.Clients() {
		if nc.Nodes[name].Start {
			continue
		}
		if err := startupNode(name, client); err != nil {
			return err
		}
		nc.Nodes[name].Start = true
	}
	return nil
}

func startupNode(name string, client *sshutil.SSHClient) error {
	var dockerComposeFilePath string
	if client.NeedCouch && client.NodeType == network.PeerNode {
		// start couchdb first
		dockerComposeFilePath = filepath.Join(client.DestPath, fmt.Sprintf("docker-compose-%s-couchdb.yaml", strings.ReplaceAll(name, ".", "-")))
		if err := client.RunCmd(fmt.Sprintf("docker-compose -f %s up -d", dockerComposeFilePath)); err != nil {
			logger.Info(err.Error())
		}
	}
	dockerComposeFilePath = filepath.Join(client.DestPath, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(name, ".", "-")))
	// start node
	if err := client.RunCmd(fmt.Sprintf("docker-compose -f %s up -d", dockerComposeFilePath)); err != nil {
		logger.Info(err.Error())
	}
	return nil
}

func ShutdownNetwork(sshUtil *sshutil.SSHUtil) error {
	for name, client := range sshUtil.Clients() {
		if err := stopNodeByNodeName(client, name); err != nil {
			return err
		}
	}
	return nil
}

func stopNodeByNodeName(client *sshutil.SSHClient, name string) error {
	if client == nil {
		return errors.Errorf("node %s does not exist", name)
	}
	if strings.HasPrefix(name, "ca_") {
		name = strings.ReplaceAll(name, "_", "-")
	}
	dockerComposeFilePath := filepath.Join(client.DestPath, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(name, ".", "-")))
	// shutdown nodes
	if err := client.RunCmd(fmt.Sprintf("docker-compose -f %s down -v", dockerComposeFilePath)); err != nil {
		logger.Info(err.Error())
	}

	// delete couchdb container
	if client.NeedCouch {
		dockerComposeFilePath := filepath.Join(client.DestPath, fmt.Sprintf("docker-compose-%s-couchdb.yaml", strings.ReplaceAll(name, ".", "-")))
		if err := client.RunCmd(fmt.Sprintf("docker-compose -f %s down -v", dockerComposeFilePath)); err != nil {
			logger.Info(err.Error())
		}
	}

	return nil
}

func CreateChannel(nc *network.NetworkConfig, dataDir string, channelId string, sdk *sdkutil.FabricSDKDriver) error {
	var peerEndpoint, orgId, ordererEndpoint string
	// find a random orderer
	peerNodes, ordererNodes, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	ordererEndpoint = ordererNodes[0].GetHostname()
	peerEndpoint = peerNodes[0].GetHostname()
	orgId = peerNodes[0].OrgId

	if err := sdk.CreateChannel(channelId, orgId, dataDir, ordererEndpoint, peerEndpoint); err != nil {
		return err
	}

	return nil
}

func JoinChannel(nc *network.NetworkConfig, channelId string, sdk *sdkutil.FabricSDKDriver) error {
	var ordererEndpoint string
	peerNodes, ordererNodes, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	// find an orderer
	ordererEndpoint = ordererNodes[0].GetHostname()
	// every peer should join the channel
	for _, node := range peerNodes {
		if err := sdk.JoinChannel(channelId, node.OrgId, ordererEndpoint, node.GetHostname()); err != nil {
			return err
		}
	}
	return nil
}

func joinChannelWithNodeName(nc *network.NetworkConfig, channelId, nodeName string, sdk *sdkutil.FabricSDKDriver) error {
	var ordererEndpoint string
	_, ordererNodes, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	// find an orderer
	ordererEndpoint = ordererNodes[0].GetHostname()
	node := nc.Nodes[nodeName]
	if err := sdk.JoinChannel(channelId, node.OrgId, ordererEndpoint, nodeName); err != nil {
		return err
	}
	return nil
}

func DoInstallOnlyCmd(dataDir, channelId, ccId, peerName string) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	peerNode := nc.GetNode(peerName)
	if err != nil {
		return err
	}
	cc, ok := nc.Chaincodes[ccId]
	if !ok {
		return errors.Wrapf(err, "chaincode %s does not exist")
	}
	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	if err := sdk.InstallCC(ccId, cc.Path, cc.Version, channelId, peerNode.OrgId, peerNode.GetHostname()); err != nil {
		return err
	}

	return nil
}

func InstallCC(nc *network.NetworkConfig, ccId, ccPath, ccVersion, channelId string, sdk *sdkutil.FabricSDKDriver) error {
	peerNodes, _, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	for _, client := range peerNodes {
		if err := sdk.InstallCC(ccId, ccPath, ccVersion, channelId, client.OrgId, client.GetHostname()); err != nil {
			return err
		}
	}
	return nil
}

func deployCCV2(nc *network.NetworkConfig, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy string, initRequired bool, initFunc string, initParams []string, redeploy bool, sdk *sdkutil.FabricSDKDriver, ccaas bool) error {
	peerNodes, ordererNodes, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	var packageId string
	// get chaincode sequence, only used for fabric v2.x
	var sequence int64
	var peers []string
	var ccPkg []byte
	if redeploy {
		if ccaas {
			ccPkg, err = ioutil.ReadFile(ccPath)
			if err != nil {
				return err
			}
		} else {
			ccOutputPath := filepath.Join(dataDir, "chaincode_packages", fmt.Sprintf("%s.tar.gz", ccId))
			_, ccPkg, err = sdk.PackageCC(ccId, ccPath, ccOutputPath)
			if err != nil {
				return err
			}
		}

		for _, chaincode := range nc.Channels[channelId].Chaincodes {
			if chaincode.Name == ccId {
				sequence = chaincode.Sequence
				break
			}
		}

		for _, client := range peerNodes {
			if packageId, err = sdk.InstallCCV2(ccId, channelId, client.OrgId, client.GetHostname(), ccPkg); err != nil {
				return err
			}
			logger.Infof("install chaincode %s success for node %s, the packageid=%s", ccId, client.GetHostname(), packageId)
			peers = append(peers, client.GetHostname())
		}
	}

	for _, client := range peerNodes {
		if err = sdk.ApproveCC(ccId, ccVersion, ccPolicy, channelId, client.OrgId, client.GetHostname(),
			ordererNodes[0].GetHostname(), packageId, sequence, initRequired); err != nil {
			return err
		}
	}

	if err := sdk.CommitCC(ccId, ccVersion, ccPolicy, channelId, peerNodes[0].OrgId,
		ordererNodes[0].GetHostname(), peers, sequence, initRequired); err != nil {
		return err
	}

	if initRequired && !ccaas {
		var peers []string
		for _, node := range peerNodes {
			peers = append(peers, node.GetHostname())
		}
		if err := sdk.InitCC(ccId, channelId, peerNodes[0].OrgId, initFunc, initParams, peers); err != nil {
			return err
		}
	}

	cc, err := sdk.QueryCC(ccId, channelId, peerNodes[0].OrgId, "Query", []string{"abc"}, peers)
	if err != nil {
		return nil
	}
	fmt.Println(cc)
	return nil
}

func InstantiateCC(nc *network.NetworkConfig, ccId, ccPath, ccVersion, channelId,
	policy string, initArgs []string, sdk *sdkutil.FabricSDKDriver) error {
	peerNodes, _, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	// pick a random peer to instantiate chaincode
	if err := sdk.InstantiateCC(ccId, ccPath, ccVersion, channelId, peerNodes[0].OrgId, policy, peerNodes[0].GetHostname(), initArgs); err != nil {
		return err
	}

	return nil
}

func UpgradeCC(nc *network.NetworkConfig, ccId, ccPath, ccVersion, channelId,
	policy string, initArgs []string, sdk *sdkutil.FabricSDKDriver) error {
	peerNodes, _, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	if err := sdk.UpdateCC(ccId, ccPath, ccVersion, channelId, peerNodes[0].OrgId, policy, initArgs, peerNodes[0].GetHostname()); err != nil {
		return err
	}
	return nil
}

func deployCCByVersion(nc *network.NetworkConfig, dataDir, channelId, ccId, ccPath, ccVersion,
	ccPolicy string, ccInitParams []string, ccInitFunc string, initRequired, ccaas bool) error {
	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	switch nc.Version {
	case fabricconfig.FabricVersion_V20:
		if err := deployCCV2(nc, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initRequired, ccInitFunc, ccInitParams, true, sdk, ccaas); err != nil {
			return err
		}
	default:
		if err := InstallCC(nc, ccId, ccPath, ccVersion, channelId, sdk); err != nil {
			return err
		}
		// InstantiateCC
		if err := InstantiateCC(nc, ccId, ccPath, ccVersion, channelId,
			ccPolicy, ccInitParams, sdk); err != nil {
			return err
		}
	}
	return nil
}

func upgradeCCByVersion(nc *network.NetworkConfig, dataDir, channelId, ccId, ccPath, ccVersion,
	ccPolicy, ccInitFunc string, ccInitParams []string, initRequired bool, redeploy bool) error {

	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	switch nc.Version {
	case fabricconfig.FabricVersion_V20:
		if err := deployCCV2(nc, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initRequired, ccInitFunc, ccInitParams, redeploy, sdk, false); err != nil {
			return err
		}
	default:
		if err := InstallCC(nc, ccId, ccPath, ccVersion, channelId, sdk); err != nil {
			return err
		}
		if err := UpgradeCC(nc, ccId, ccPath, ccVersion, channelId, ccPolicy, ccInitParams, sdk); err != nil {
			return err
		}
	}
	return err
}

// ==========================cmd=========================

func DoGenerateBootstrapCommand(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc string, ccInitParams []string,
	ccPolicy string, ccInitRequired bool, sequence int64, ifCouchdb bool, peerUrls, ordererUrls []string, fVersion string, useFile bool) error {
	var err error
	networkConfig := &network.NetworkConfig{}
	if !useFile {
		networkConfig, err = GenerateNetwork(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParams, ccPolicy, ccInitRequired, sequence, ifCouchdb, peerUrls, ordererUrls, fVersion)
		if err != nil {
			return err
		}
	} else {
		networkConfig, err = network.UnmarshalNetworkConfig(dataDir)
		if err != nil {
			return err
		}
	}
	if networkName == "" {
		networkName = networkConfig.Name
	}
	config, err := network.Load()
	// if config does not exist
	if err != nil {
		config = network.NewHomeDirConfig()
	}
	if err := config.AddNetwork(networkName, dataDir); err != nil {
		return err
	}
	if err := GenerateCryptoConfig(dataDir, networkConfig); err != nil {
		return err
	}
	if err := GenerateConfigtx(dataDir, consensus, channelId, fVersion, networkConfig); err != nil {
		return err
	}
	if err := GenerateDockerCompose(dataDir, networkName, networkConfig.GetPeerNodes(), networkConfig.GetOrdererNodes(), networkConfig.GetImageTags(), ifCouchdb, false); err != nil {
		return err
	}
	if err := GenerateConnectionProfile(dataDir, channelId, networkConfig); err != nil {
		return err
	}

	if err := config.Store(); err != nil {
		return err
	}
	return nil
}

func DoStartupCommand(dataDir string, startOnly bool) error {
	if err := fabricconfig.GenerateKeyPairsAndCerts(dataDir); err != nil {
		return err
	}
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}

	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	sshUtil, err := ReadSSHConfigFromNetwork(dataDir, nc)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()

	logger.Info("begin to start network")
	for channelId := range nc.Channels {
		if err := fabricconfig.GenerateGenesisBlockAndChannelTxAndAnchorPeer(dataDir, channelId, nc); err != nil {
			return err
		}
	}

	if err := TransferFilesAllNodes(sshUtil, dataDir, nc); err != nil {
		return err
	}

	if err := StartupNetwork(sshUtil, nc); err != nil {
		return err
	}

	logger.Info("starting network complete!")
	// if only starting the fabric docker container
	if startOnly {
		return nil
	}

	if nc.Consensus == fabricconfig.OrdererType_ETCDRAFT {
		logger.Info("sleeping 15s to allow etcdraft cluster to complete booting")
		time.Sleep(time.Second * 15)
	} else {
		time.Sleep(time.Second * 5)
	}

	for channelId, channel := range nc.Channels {
		// create channel
		if err := CreateChannel(nc, dataDir, channelId, sdk); err != nil {
			return err
		}
		// join channel
		if err := JoinChannel(nc, channelId, sdk); err != nil {
			return err
		}
		if len(channel.Chaincodes) > 0 {
			for _, chaincode := range channel.Chaincodes {
				cc := nc.Chaincodes[chaincode.Name]
				if nc.Version == fabricconfig.FabricVersion_V20 {
					if err := deployCCV2(nc, dataDir, channelId, chaincode.Name, cc.Path, cc.Version, cc.Policy,
						cc.InitRequired, cc.InitFunc, cc.InitParam, true, sdk, false); err != nil {
						return err
					}
				} else {
					// install chaincode
					if err := InstallCC(nc, chaincode.Name, cc.Path, cc.Version, channelId, sdk); err != nil {
						return err
					}
					// InstantiateCC
					if err := InstantiateCC(nc, chaincode.Name, cc.Path, cc.Version, channelId,
						cc.Policy, cc.InitParam, sdk); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func DoDeployccCmd(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc string, initParams []string, initRequired, ccaas bool) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	if err = nc.ExtendChannelChaincode(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParams, initRequired); err != nil {
		return err
	}

	if err = deployCCByVersion(nc, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initParams, initFunc, initRequired, ccaas); err != nil {
		return err
	}
	return err
}

func DoUpgradeccCmd(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc string, initParams []string, initRequired bool, redeploy bool) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	if err = nc.UpgradeChaincode(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParams, initRequired); err != nil {
		return err
	}
	if err = upgradeCCByVersion(nc, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParams, initRequired, redeploy); err != nil {
		return err
	}
	return err
}

func DoShutdownCommand(dataDir string) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	sshUtil, err := ReadSSHConfigFromNetwork(dataDir, nc)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()
	if err := ShutdownNetwork(sshUtil); err != nil {
		return err
	}
	return nil
}

func DoCreateChannelCommand(dataDir, channelId string, peers []string) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	// update network config
	if err := nc.ExtendChannel(dataDir, channelId, peers); err != nil {
		return err
	}
	profile, err := connectionprofile.UnmarshalConnectionProfile(dataDir)
	if err != nil {
		return err
	}
	if err := profile.ExtendChannel(dataDir, channelId, peers); err != nil {
		return err
	}

	// update consortium of the system channel
	var nodes []*network.Node
	var peerOrgs []string
	for _, peer := range peers {
		node, ok := nc.Nodes[peer]
		if !ok {
			return errors.Errorf("node %s does not exist", peer)
		}
		nodes = append(nodes, node)
		peerOrgs = append(peerOrgs, node.OrgId)
	}
	orgs := fabricconfig.BuildConsortiumOrgs(dataDir, nodes, nc.Version)
	group, err := fabricconfig.BuildConsortiumConfigGroup(&genesisconfig.Consortium{Organizations: orgs})
	if err != nil {
		return err
	}
	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	ordererNodes := nc.GetOrdererNodes()
	if err := sdk.ExtendConsortium(fabricconfig.DefaultGenesisChannel,
		fmt.Sprintf(fabricconfig.DefaultConsortiumNameTemplate, channelId),
		ordererNodes[0].OrgId, ordererNodes[0].GetHostname(), group); err != nil {
		return err
	}

	// generate channel tx
	if err = fabricconfig.ExtendChannelProfile(dataDir, nc.Version, nc.GetPeerNodes(),
		&fabricconfig.ConfigtxApplicationChannelEntity{
			ChannelId: channelId,
			Peers:     utils.DeduplicatedSlice(peerOrgs),
		}); err != nil {
		return err
	}
	if err = fabricconfig.GenerateChannelTxAndAnchorPeer(dataDir, channelId, nc); err != nil {
		return err
	}
	sshUtil, err := ReadSSHConfigFromNetwork(dataDir, nc)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()
	if err = TransferNewChannelFiles(dataDir, channelId, sshUtil, nc); err != nil {
		return err
	}

	// create channel
	if err = CreateChannel(nc, dataDir, channelId, sdk); err != nil {
		return err
	}
	// join channel
	if err = JoinChannel(nc, channelId, sdk); err != nil {
		return err
	}
	return nil
}

func DoExtendNodeCommand(dataDir string, couchdb bool, peers, orderers []string) error {
	// extend network config file
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}

	newPeerNodes, newOrdererNodes, err := nc.ExtendNode(dataDir, couchdb, peers, orderers)
	if err != nil {
		return err
	}
	// extend crypto-config file
	if err := fabricconfig.ExtendCryptoConfigFile(dataDir, newPeerNodes, newOrdererNodes); err != nil {
		return err
	}
	// generate keypairs and certs
	if err := fabricconfig.ExtendKeyPairsAndCerts(dataDir); err != nil {
		return err
	}
	// generate docker-compose file
	if err := GenerateDockerCompose(dataDir, nc.Name, newPeerNodes, newOrdererNodes, nc.GetImageTags(), couchdb, false); err != nil {
		return err
	}
	// update connection-profile
	profile, err := connectionprofile.UnmarshalConnectionProfile(dataDir)
	if err != nil {
		return err
	}
	if err := profile.ExtendNodesAndOrgs(dataDir, newPeerNodes, newOrdererNodes); err != nil {
		return err
	}
	return nil
}

func DoStartNodeCmd(dataDir string, nodeNames ...string) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	sshUtil, err := ReadSSHConfigFromNetwork(dataDir, nc)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()
	// transfer docker-compose files
	if err := TransferFilesByNodeName(sshUtil, dataDir, nodeNames); err != nil {
		return err
	}
	// start nodes
	for _, name := range nodeNames {
		// we don't need to check nil for the client, since TransferFilesByNodeName has done it
		client := sshUtil.GetClientByName(name)
		if err := startupNode(name, client); err != nil {
			return err
		}
	}
	return nil
}

func DoStopNodeCmd(dataDir string, nodeNames ...string) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	sshUtil, err := ReadSSHConfigFromNetwork(dataDir, nc)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()
	for _, name := range nodeNames {
		client := sshUtil.GetClientByName(name)
		if err := stopNodeByNodeName(client, name); err != nil {
			return err
		}
	}
	return nil
}

func DoExistOrgPeerJoinChannel(dataDir string, channelId, nodeName string) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	// join the channel using sdk
	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()
	if err := joinChannelWithNodeName(nc, channelId, nodeName, sdk); err != nil {
		return err
	}
	return nil
}

func DoNewOrgPeerJoinChannel(dataDir string, channelId, nodeName string) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	node := nc.GetNode(nodeName)
	peerNodes, ordererNodes, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	ordererEndpoint := ordererNodes[0].GetHostname()
	var existOrgs []string
	for _, peerNode := range peerNodes {
		existOrgs = append(existOrgs, peerNode.OrgId)
	}
	orgProfile := fabricconfig.GenerateOrgProfile(dataDir, node, nc.Version)
	configGroup, err := fabricconfig.GenerateApplicationConfigGroup(orgProfile)
	if err != nil {
		return err
	}

	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()
	if err = sdk.ExtendOrShrinkChannel("extend", channelId, node.OrgId, ordererEndpoint, existOrgs, configGroup); err != nil {
		return err
	}
	// sleep 5s
	time.Sleep(5 * time.Second)
	if err := joinChannelWithNodeName(nc, channelId, nodeName, sdk); err != nil {
		return err
	}
	return nil
}

func DoCreateCANodeForOrg(dataDir, orgId, enrollId, enrollSecret string) error {
	// store ca info
	config, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}

	port := utils.GetRandomPort()
	if err = config.CreateCANode(dataDir, orgId, enrollId, enrollSecret, port); err != nil {
		return err
	}
	// generate docker-compose file of ca node
	domain := config.GetOrgDomain(orgId)
	if domain == "" {
		return errors.Errorf("org %s does not exist", domain)
	}
	if err = docker_compose.GenerateCA(dataDir, orgId, domain, port, config.CAImageTag, enrollId, enrollSecret); err != nil {
		return err
	}
	// update connection-profile
	profile, err := connectionprofile.UnmarshalConnectionProfile(dataDir)
	if err != nil {
		return err
	}
	node, caInfo := config.GetCAByOrgId(orgId)
	if err = profile.ExtendCANodeByOrg(dataDir, node, caInfo); err != nil {
		return err
	}
	// sftp docker-compose file
	sshUtil, err := ReadSSHConfigFromNetwork(dataDir, config)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()
	sshClient := sshUtil.GetClientByName(node.Name)
	dockerComposeFilePath := filepath.Join(dataDir, fmt.Sprintf("docker-compose-ca-%s.yaml", domain))
	if err = sshClient.Sftp(dockerComposeFilePath, dataDir); err != nil {
		return err
	}
	// run ca node
	if err := sshClient.RunCmd(fmt.Sprintf("docker-compose -f %s up -d", dockerComposeFilePath)); err != nil {
		logger.Info(err.Error())
	}
	time.Sleep(3 * time.Second)
	// enroll registrar
	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()
	return sdk.Enroll(orgId, enrollId, enrollSecret)
}

func DoCallFabricCAFunc(dataDir, orgId, username, secret, funcName string) (string, error) {
	config, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return "", err
	}
	var caName string
	for name, node := range config.Nodes {
		if node.Type == network.CANode && node.OrgId == orgId {
			caName = name
		}
	}

	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return "", err
	}
	defer sdk.Close()

	switch funcName {
	case "register":
		secret, err = sdk.Register(orgId, username, secret, caName)
		if err != nil {
			return "", err
		}
		if err := sdk.Enroll(orgId, username, secret); err != nil {
			return "", err
		}
	case "revoke":
		if err := sdk.Revoke(orgId, username, caName); err != nil {
			return "", err
		}
	}
	return secret, nil
}

func DoNetworkImport(name, dataDir string) error {
	hdc, err := network.Load()
	if err != nil {
		return err
	}
	networkPath := hdc.GetNetworkPath(name)
	if networkPath != "" {
		return errors.Errorf("network %s exists, current path is %s", name, dataDir)
	}

	stat, err := os.Stat(dataDir)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return errors.Errorf("%s should be a directory", dataDir)
	}

	dir, err := ioutil.ReadDir(dataDir)
	if err != nil {
		return err
	}

	fileMap := map[string]bool{
		"networkconfig.yaml":     false,
		"crypto-config.yaml":     false,
		"connection-config.yaml": false,
		"configtx.yaml":          false,
	}

	for _, info := range dir {
		fileMap[info.Name()] = true
	}
	for fName, b := range fileMap {
		if !b {
			return errors.Errorf("config file %s not exists", fName)
		}
	}

	hdc.AddNetwork(name, dataDir)
	return hdc.Store()
}

func GetNetworkPathByName(dataDir, name string) (string, error) {
	var err error
	if dataDir != "" {
		if !filepath.IsAbs(dataDir) {
			if dataDir, err = filepath.Abs(dataDir); err != nil {
				return "", err
			}
		}
		return dataDir, nil
	}
	// use network name
	if name == "" {
		return "", errors.New("network name is nil")
	}
	networkPathConf, err := network.Load()
	if err != nil {
		return "", err
	}
	path := networkPathConf.GetNetworkPath(name)
	if path == "" {
		return "", errors.New("network does not exist")
	}
	return path, nil
}

func NetworkExist(name string) bool {
	if name == "" {
		return true
	}
	networkPathConf, err := network.Load()
	if err != nil {
		return true
	}
	path := networkPathConf.GetNetworkPath(name)
	if path == "" {
		return false
	}
	return false
}
