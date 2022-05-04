package utils

import (
	"fmt"
	"github.com/pkg/errors"
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

func GenerateNetwork(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParam, ccPolicy string, ccInitRequired bool, sequence int64, couchdb bool, peerUrls, ordererUrls []string, networkVersion string) (*network.NetworkConfig, error) {
	return network.GenerateNetworkConfig(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParam, ccPolicy, ccInitRequired, sequence, couchdb, peerUrls, ordererUrls, networkVersion)
}

func GenerateConfigtx(dataDir, consensus, channelId, fversion string, networkConfig *network.NetworkConfig) error {
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
	if err := fabricconfig.GenerateConfigtxFile(dataDir, consensus, ordererNodes, peerNodes, fversion); err != nil {
		return err
	}
	return nil
}

func GenerateDockerCompose(dataDir string, peerNodes, ordererNodes []*network.Node, imageTags []string, couchdb bool, ca bool) error {
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
	var pUrls []string
	var oUrls []string
	peerNodes, ordererNodes, err := networkConfig.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	for _, node := range peerNodes {
		pUrls = append(pUrls, fmt.Sprintf("%s:%d:%s", node.GetHostname(), node.NodePort, node.Host))
	}
	for _, node := range ordererNodes {
		oUrls = append(oUrls, fmt.Sprintf("%s:%d:%s", node.GetHostname(), node.NodePort, node.Host))
	}
	return connectionprofile.GenerateNetworkConnProfile(dataDir, channelId, pUrls, oUrls)
}

func ReadSSHConfigFromNetwork(networkConfig *network.NetworkConfig) (*sshutil.SSHUtil, error) {
	sshUtil := sshutil.NewSSHUtil()
	for _, node := range networkConfig.GetPeerNodes() {
		if err := sshUtil.Add(node.GetHostname(), node.Username, node.Password, fmt.Sprintf("%s:%d", node.Host, node.SSHPort), node.Type, node.Couch); err != nil {
			return nil, err
		}
	}
	for _, node := range networkConfig.GetOrdererNodes() {
		if err := sshUtil.Add(node.GetHostname(), node.Username, node.Password, fmt.Sprintf("%s:%d", node.Host, node.SSHPort), node.Type, false); err != nil {
			return nil, err
		}
	}

	for _, node := range networkConfig.GetCANodes() {
		if err := sshUtil.Add(node.GetHostname(), node.Username, node.Password, fmt.Sprintf("%s:%d", node.Host, node.SSHPort), node.Type, false); err != nil {
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

func TransferFilesAllNodes(sshUtil *sshutil.SSHUtil, dataDir string) error {
	for name, client := range sshUtil.Clients() {
		if err := transferFiles(client, name, dataDir); err != nil {
			return err
		}
	}
	return nil
}

func transferFiles(sshClient *sshutil.SSHClient, nodeName, dataDir string) error {
	ordererCryptoConfigPrefix := filepath.Join(dataDir, "crypto-config", "ordererOrganizations")
	peerCryptoConfigPrefix := filepath.Join(dataDir, "crypto-config", "peerOrganizations")
	_, orgName, _ := utils.SplitNameOrgDomain(nodeName)
	// send node self keypairs and certs
	var certDir string
	if sshClient.NodeType == network.PeerNode {
		certDir = filepath.Join(peerCryptoConfigPrefix, orgName)
	} else {
		certDir = filepath.Join(ordererCryptoConfigPrefix, orgName)
	}
	err := sshClient.Sftp(certDir, certDir)
	if err != nil {
		return err
	}
	// send genesis.block, channel.tx and anchor.tx
	channelArtifactsPath := filepath.Join(dataDir, "channel-artifacts")
	if err = sshClient.Sftp(channelArtifactsPath, channelArtifactsPath); err != nil {
		return err
	}

	dockerComposeFilePath := filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(nodeName, ".", "-")))
	if err = sshClient.Sftp(dockerComposeFilePath, dataDir); err != nil {
		return err
	}
	if sshClient.NeedCouch {
		dockerComposeFilePath = filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s-couchdb.yaml", strings.ReplaceAll(nodeName, ".", "-")))
		if err = sshClient.Sftp(dockerComposeFilePath, dataDir); err != nil {
			return err
		}
	}
	return nil
}

func TransferNewChannelFiles(dataDir string, channelId string, sshUtil *sshutil.SSHUtil, nc *network.NetworkConfig) error {
	channelTxPath := filepath.Join(dataDir, "channel-artifacts", fmt.Sprintf("%s.tx", channelId))
	peerNodes, ordererNodes, err := nc.GetNodesByChannel(channelId)
	if err != nil {
		return err
	}
	sshClients := sshUtil.Clients()
	for _, node := range peerNodes {
		if err := sshClients[node.GetHostname()].Sftp(channelTxPath, channelTxPath); err != nil {
			return err
		}
	}
	for _, node := range ordererNodes {
		if err := sshClients[node.GetHostname()].Sftp(channelTxPath, channelTxPath); err != nil {
			return err
		}
	}
	return nil
}

func StartupNetwork(sshUtil *sshutil.SSHUtil, dataDir string) error {
	logger.Info("begin to start network")
	for name, client := range sshUtil.Clients() {
		if err := startupNode(dataDir, name, client); err != nil {
			return err
		}
	}
	logger.Info("starting network complete!")
	return nil
}

func startupNode(dataDir, name string, client *sshutil.SSHClient) error {
	var dockerComposeFilePath string
	if client.NeedCouch && client.NodeType == network.PeerNode {
		// start couchdb first
		dockerComposeFilePath = filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s-couchdb.yaml", strings.ReplaceAll(name, ".", "-")))
		if err := client.RunCmd(fmt.Sprintf("docker-compose -f %s up -d", dockerComposeFilePath)); err != nil {
			logger.Info(err.Error())
		}
	}
	dockerComposeFilePath = filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(name, ".", "-")))
	// start node
	if err := client.RunCmd(fmt.Sprintf("docker-compose -f %s up -d", dockerComposeFilePath)); err != nil {
		logger.Info(err.Error())
	}
	return nil
}

func ShutdownNetwork(sshUtil *sshutil.SSHUtil, dataDir string) error {
	for name, client := range sshUtil.Clients() {
		if err := stopNodeByNodeName(dataDir, client, name); err != nil {
			return err
		}
	}
	return nil
}

func stopNodeByNodeName(dataDir string, client *sshutil.SSHClient, name string) error {
	if client == nil {
		return errors.Errorf("node %s does not exist", name)
	}
	if strings.HasPrefix(name, "ca_") {
		name = strings.ReplaceAll(name, "_", "-")
	}
	dockerComposeFilePath := filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(name, ".", "-")))
	// shutdown nodes
	if err := client.RunCmd(fmt.Sprintf("docker-compose -f %s down -v", dockerComposeFilePath)); err != nil {
		logger.Info(err.Error())
	}

	// delete couchdb container
	if client.NeedCouch {
		dockerComposeFilePath := filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s-couchdb.yaml", strings.ReplaceAll(name, ".", "-")))
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

func deployCCV2(nc *network.NetworkConfig, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy string, initRequired bool, initFunc string, initParams string, redeploy bool, sdk *sdkutil.FabricSDKDriver, ccaas bool) error {
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
		initArgs := strings.Split(initParams, ",")
		if err := sdk.InitCC(ccId, channelId, peerNodes[0].OrgId, initFunc, initArgs, peers); err != nil {
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
	policy, initArgsStr string, sdk *sdkutil.FabricSDKDriver) error {
	initArgs := strings.Split(initArgsStr, ",")

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
	policy, initArgsStr string, sdk *sdkutil.FabricSDKDriver) error {
	initArgs := strings.Split(initArgsStr, ",")
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
	ccPolicy, ccInitParam string, ccInitFunc string, initRequired, ccaas bool) error {
	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	switch nc.Version {
	case fabricconfig.FabricVersion_V20:
		if err := deployCCV2(nc, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initRequired, ccInitFunc, ccInitParam, true, sdk, ccaas); err != nil {
			return err
		}
	default:
		if err := InstallCC(nc, ccId, ccPath, ccVersion, channelId, sdk); err != nil {
			return err
		}
		// InstantiateCC
		if err := InstantiateCC(nc, ccId, ccPath, ccVersion, channelId,
			ccPolicy, ccInitParam, sdk); err != nil {
			return err
		}
	}
	return nil
}

func upgradeCCByVersion(nc *network.NetworkConfig, dataDir, channelId, ccId, ccPath, ccVersion,
	ccPolicy, ccInitFunc, ccInitParam string, initRequired bool, redeploy bool) error {

	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	switch nc.Version {
	case fabricconfig.FabricVersion_V20:
		if err := deployCCV2(nc, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initRequired, ccInitFunc, ccInitParam, redeploy, sdk, false); err != nil {
			return err
		}
	default:
		if err := InstallCC(nc, ccId, ccPath, ccVersion, channelId, sdk); err != nil {
			return err
		}
		if err := UpgradeCC(nc, ccId, ccPath, ccVersion, channelId, ccPolicy, ccInitParam, sdk); err != nil {
			return err
		}
	}
	return err
}

// ==========================cmd=========================

func DoGenerateBootstrapCommand(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParam,
	ccPolicy string, ccInitRequired bool, sequence int64, ifCouchdb bool, peerUrls, ordererUrls []string, fVersion string) error {
	config, err := network.Load()
	// config does not exist
	if err != nil {
		config = network.NewHomeDirConfig()
	}
	if err := config.AddNetwork(networkName, dataDir); err != nil {
		return err
	}
	networkConfig, err := GenerateNetwork(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitFunc, ccInitParam, ccPolicy, ccInitRequired, sequence, ifCouchdb, peerUrls, ordererUrls, fVersion)
	if err != nil {
		return err
	}
	if err := GenerateCryptoConfig(dataDir, networkConfig); err != nil {
		return err
	}
	if err := GenerateConfigtx(dataDir, consensus, channelId, fVersion, networkConfig); err != nil {
		return err
	}
	if err := GenerateDockerCompose(dataDir, networkConfig.GetPeerNodes(), networkConfig.GetOrdererNodes(), networkConfig.GetImageTags(), ifCouchdb, false); err != nil {
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
	// startup command only starts a fabric network with one channel and one chaincode right now
	// TODO: support multi channels
	var channelId, ccId, ccPath, ccInitParam, ccVersion, ccPolicy, ccInitFunc, consensus string
	var ifInstallCC, ccInitRequired bool
	for name, channel := range nc.Channels {
		channelId = name
		if len(channel.Chaincodes) > 0 {
			ifInstallCC = true
		} else {
			break
		}
		ccPath = nc.Chaincodes[channel.Chaincodes[0].Name].Path
		ccInitParam = nc.Chaincodes[channel.Chaincodes[0].Name].InitParam
		ccInitFunc = nc.Chaincodes[channel.Chaincodes[0].Name].InitFunc
		ccInitRequired = nc.Chaincodes[channel.Chaincodes[0].Name].InitRequired
		ccVersion = nc.Chaincodes[channel.Chaincodes[0].Name].Version
		ccPolicy = nc.Chaincodes[channel.Chaincodes[0].Name].Policy
		ccId = channel.Chaincodes[0].Name
		consensus = channel.Consensus
	}
	if err := fabricconfig.GenerateGenesisBlockAndChannelTxAndAnchorPeer(dataDir, channelId, nc); err != nil {
		return err
	}
	sshUtil, err := ReadSSHConfigFromNetwork(nc)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()

	if err := TransferFilesAllNodes(sshUtil, dataDir); err != nil {
		return err
	}
	if err := StartupNetwork(sshUtil, dataDir); err != nil {
		return err
	}
	// if only starting the fabric docker container
	if startOnly {
		return nil
	}

	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	if consensus == fabricconfig.OrdererType_ETCDRAFT {
		logger.Info("sleeping 15s to allow etcdraft cluster to complete booting")
		time.Sleep(time.Second * 15)
	} else {
		time.Sleep(time.Second * 5)
	}

	// create channel
	if err := CreateChannel(nc, dataDir, channelId, sdk); err != nil {
		return err
	}
	// join channel
	if err := JoinChannel(nc, channelId, sdk); err != nil {
		return err
	}
	if ifInstallCC {
		if nc.Version == fabricconfig.FabricVersion_V20 {
			if err := deployCCV2(nc, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, ccInitRequired, ccInitFunc, ccInitParam, true, sdk, false); err != nil {
				return err
			}
		} else {
			// install chaincode
			if err := InstallCC(nc, ccId, ccPath, ccVersion, channelId, sdk); err != nil {
				return err
			}
			// InstantiateCC
			if err := InstantiateCC(nc, ccId, ccPath, ccVersion, channelId,
				ccPolicy, ccInitParam, sdk); err != nil {
				return err
			}
		}
	}
	return nil
}

func DoDeployccCmd(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParam string, initRequired, ccaas bool) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	if err = nc.ExtendChannelChaincode(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParam, initRequired); err != nil {
		return err
	}

	if err = deployCCByVersion(nc, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initParam, initFunc, initRequired, ccaas); err != nil {
		return err
	}
	return err
}

func DoUpgradeccCmd(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParam string, initRequired bool, redeploy bool) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	if err = nc.UpgradeChaincode(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParam, initRequired); err != nil {
		return err
	}
	if err = upgradeCCByVersion(nc, dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParam, initRequired, redeploy); err != nil {
		return err
	}
	return err
}

func DoShutdownCommand(dataDir string) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	sshUtil, err := ReadSSHConfigFromNetwork(nc)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()
	if err := ShutdownNetwork(sshUtil, dataDir); err != nil {
		return err
	}
	return nil
}

func DoCreateChannelCommand(dataDir, channelId, consensus string, peers, orderers []string) error {
	nc, err := network.UnmarshalNetworkConfig(dataDir)
	if err != nil {
		return err
	}
	if err := nc.ExtendChannel(dataDir, channelId, consensus, peers, orderers); err != nil {
		return err
	}
	profile, err := connectionprofile.UnmarshalConnectionProfile(dataDir)
	if err != nil {
		return err
	}
	if err := profile.ExtendChannel(dataDir, channelId, peers); err != nil {
		return err
	}
	if err := fabricconfig.GenerateChannelTxAndAnchorPeer(dataDir, channelId, nc); err != nil {
		return err
	}
	sshUtil, err := ReadSSHConfigFromNetwork(nc)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()
	if err := TransferNewChannelFiles(dataDir, channelId, sshUtil, nc); err != nil {
		return err
	}
	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionprofile.DefaultConnProfileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	// create channel
	if err := CreateChannel(nc, dataDir, channelId, sdk); err != nil {
		return err
	}
	// join channel
	if err := JoinChannel(nc, channelId, sdk); err != nil {
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
	if err := GenerateDockerCompose(dataDir, newPeerNodes, newOrdererNodes, nc.GetImageTags(), couchdb, false); err != nil {
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
	sshUtil, err := ReadSSHConfigFromNetwork(nc)
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
		if err := startupNode(dataDir, name, client); err != nil {
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
	sshUtil, err := ReadSSHConfigFromNetwork(nc)
	if err != nil {
		return err
	}
	defer sshUtil.CloseAll()
	for _, name := range nodeNames {
		client := sshUtil.GetClientByName(name)
		if err := stopNodeByNodeName(dataDir, client, name); err != nil {
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
	configGroup, err := fabricconfig.GenerateConfigGroup(orgProfile)
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
	// sleep 5s,
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
	sshUtil, err := ReadSSHConfigFromNetwork(config)
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
