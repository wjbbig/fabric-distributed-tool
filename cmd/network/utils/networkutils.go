package utils

import (
	"fmt"
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

const (
	connectionConfigFileName = "connection-config.yaml"
)

func GenerateCryptoConfig(dataDir string, networkConfig *network.NetworkConfig) error {
	if err := fabricconfig.GenerateCryptoConfigFile(dataDir, networkConfig.GetPeerNodes(), networkConfig.GetOrdererNodes()); err != nil {
		return err
	}
	return nil
}

func GenerateNetwork(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitParam, ccPolicy string, ccInitRequired bool, sequence int64, couchdb bool, peerUrls, ordererUrls []string) (*network.NetworkConfig, error) {
	return network.GenerateNetworkConfig(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitParam, ccPolicy, ccInitRequired, sequence, couchdb, peerUrls, ordererUrls)
}

func GenerateConfigtx(dataDir, consensus, channelId string, networkConfig *network.NetworkConfig) error {
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
	if err := fabricconfig.GenerateConfigtxFile(dataDir, consensus, ordererNodes, peerNodes); err != nil {
		return err
	}
	return nil
}

func GenerateDockerCompose(dataDir string, networkConfig *network.NetworkConfig, couchdb bool) error {
	peerNodes := networkConfig.GetPeerNodes()
	ordererNodes := networkConfig.GetOrdererNodes()
	peersByOrg := make(map[string][]*network.Node)
	for _, peerNode := range networkConfig.GetPeerNodes() {
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
		if err := docker_compose.GeneratePeerDockerComposeFile(dataDir, peer, gossipUrl, extraHosts, couchdb); err != nil {
			return err
		}
	}
	for _, orderer := range ordererNodes {
		var extraHosts []string
		extraHosts = append(extraHosts, spliceHostnameAndIP(orderer, peerNodes)...)
		extraHosts = append(extraHosts, spliceHostnameAndIP(orderer, ordererNodes)...)
		if err := docker_compose.GenerateOrdererDockerComposeFile(dataDir, orderer, extraHosts); err != nil {
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
		if err := sshUtil.Add(node.GetHostname(), node.Username, node.Password, fmt.Sprintf("%s:%d", node.Host, node.SSHPort), node.Type, node.Couch); err != nil {
			return nil, err
		}
	}
	return sshUtil, nil
}

func TransferFilesByPeerName(sshUtil *sshutil.SSHUtil, dataDir string) error {
	ordererCryptoConfigPrefix := filepath.Join(dataDir, "crypto-config", "ordererOrganizations")
	peerCryptoConfigPrefix := filepath.Join(dataDir, "crypto-config", "peerOrganizations")
	for name, client := range sshUtil.Clients() {
		_, orgName, _ := utils.SplitNameOrgDomain(name)
		// send node self keypairs and certs
		var certDir string
		if client.NodeType == network.PeerNode {
			certDir = filepath.Join(peerCryptoConfigPrefix, orgName)
		} else {
			certDir = filepath.Join(ordererCryptoConfigPrefix, orgName)
		}
		err := client.Sftp(certDir, certDir)
		if err != nil {
			return err
		}
		// send genesis.block, channel.tx and anchor.tx
		channelArtifactsPath := filepath.Join(dataDir, "channel-artifacts")
		if err = client.Sftp(channelArtifactsPath, channelArtifactsPath); err != nil {
			return err
		}

		dockerComposeFilePath := filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(name, ".", "-")))
		if err = client.Sftp(dockerComposeFilePath, dataDir); err != nil {
			return err
		}
		if client.NeedCouch {
			dockerComposeFilePath = filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s-couchdb.yaml", strings.ReplaceAll(name, ".", "-")))
			if err = client.Sftp(dockerComposeFilePath, dataDir); err != nil {
				return err
			}
		}
	}
	return nil
}

func StartupNetwork(sshUtil *sshutil.SSHUtil, dataDir string) error {
	logger.Info("begin to start network")
	for name, client := range sshUtil.Clients() {
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
		// TODO start ca if chosen
	}
	logger.Info("starting network complete!")
	return nil
}

func ShutdownNetwork(sshUtil *sshutil.SSHUtil, dataDir string) error {
	for name, client := range sshUtil.Clients() {
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

		// TODO: delete chaincode container and image
		// if client.NodeType == network.PeerNode {
		// 	if err := client.RunCmd("docker rm -f $(docker ps -a -q |  grep \"dev*\"  | awk '{print $1}')"); err != nil {
		// 		logger.Info(err.Error())
		// 	}
		// }
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

// ==========================cmd=========================

func DoGenerateBootstrapCommand(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitParam, ccPolicy string, ccInitRequired bool, sequence int64, ifCouchdb bool, peerUrls, ordererUrls []string) error {
	networkConfig, err := GenerateNetwork(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion, ccInitParam, ccPolicy, ccInitRequired, sequence, ifCouchdb, peerUrls, ordererUrls)
	if err != nil {
		return err
	}
	if err := GenerateCryptoConfig(dataDir, networkConfig); err != nil {
		return err
	}
	if err := GenerateConfigtx(dataDir, consensus, channelId, networkConfig); err != nil {
		return err
	}
	if err := GenerateDockerCompose(dataDir, networkConfig, ifCouchdb); err != nil {
		return err
	}
	if err := GenerateConnectionProfile(dataDir, channelId, networkConfig); err != nil {
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
	var channelId, ccId, ccPath, ccInitParam, ccVersion, ccPolicy, consensus string
	var ifInstallCC bool
	for name, channel := range nc.Channels {
		channelId = name
		if len(channel.Chaincodes) > 0 {
			ifInstallCC = true
		} else {
			break
		}
		ccPath = nc.Chaincodes[channel.Chaincodes[0].Name].Path
		ccInitParam = nc.Chaincodes[channel.Chaincodes[0].Name].InitParam
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

	if err := TransferFilesByPeerName(sshUtil, dataDir); err != nil {
		return err
	}
	if err := StartupNetwork(sshUtil, dataDir); err != nil {
		return err
	}
	// if only starting the fabric docker container
	if startOnly {
		return nil
	}

	sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionConfigFileName))
	if err != nil {
		return err
	}
	defer sdk.Close()

	if consensus == fabricconfig.OrdererType_ETCDRAFT {
		logger.Info("sleeping 15s to allow etcdraft cluster to complete booting")
		time.Sleep(time.Second * 15)
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
	return nil
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
