package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/wjbbig/fabric-distributed-tool/connectionprofile"
	docker_compose "github.com/wjbbig/fabric-distributed-tool/docker-compose"
	"github.com/wjbbig/fabric-distributed-tool/fabricconfig"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/sshutil"
	"github.com/wjbbig/fabric-distributed-tool/utils"
)

// utils for network cmd
var logger = mylogger.NewLogger()

func GenerateCryptoConfig(dataDir string, peerUrls, ordererUrls []string) error {
	var peers []string
	var orderers []string
	for _, url := range peerUrls {
		index := strings.Index(url, ":")
		peers = append(peers, url[:index])
	}

	for _, url := range ordererUrls {
		index := strings.Index(url, ":")
		orderers = append(orderers, url[:index])
	}
	if err := fabricconfig.GenerateCryptoConfigFile(dataDir, peers, orderers); err != nil {
		return err
	}
	if err := fabricconfig.GenerateKeyPairsAndCerts(dataDir); err != nil {
		return errors.Wrap(err, "generate fabric keypairs and certs failed")
	}

	return nil
}

func GenerateSSHConfig(dataDir string, peerUrls, ordererUrls []string) error {
	var clients []sshutil.Client
	for _, url := range peerUrls {
		clients = append(clients, sshutil.NewClient(url, "peer"))
	}
	for _, url := range ordererUrls {
		clients = append(clients, sshutil.NewClient(url, "orderer"))
	}
	return sshutil.GenerateSSHConfig(dataDir, clients)
}

func GenerateConfigtx(dataDir, consensus, channelId string, peerUrls, ordererUrls []string) error {
	var peers []string
	var orderers []string
	var orgs []string
	for _, url := range peerUrls {
		index := strings.Index(url, "@")
		peerURl := url[:index]
		args := strings.Split(peerURl, ":")
		_, org, _ := utils.SplitNameOrgDomain(args[0])
		peers = append(peers, peerURl)
		orgs = append(orgs, org)
	}

	for _, url := range ordererUrls {
		index := strings.Index(url, "@")
		orderers = append(orderers, url[:index])
	}
	if consensus == "" {
		if len(orderers) == 1 {
			consensus = "solo"
		} else {
			consensus = "etcdraft"
		}
	}

	if err := fabricconfig.GenerateConfigtxFile(dataDir, consensus, orderers, peers); err != nil {
		return err
	}
	if err := fabricconfig.GenerateGensisBlockAndChannelTxAndAnchorPeer(dataDir, channelId, orgs); err != nil {
		return err
	}
	return nil
}

func GenerateDockerCompose(dataDir string, peerUrls, ordererUrls []string) error {
	var peers []string
	var orderers []string
	peersByOrg := make(map[string][]string)
	for _, url := range peerUrls {
		index := strings.Index(url, "@")
		peerURl := url[:index]
		args := strings.Split(peerURl, ":")
		_, org, _ := utils.SplitNameOrgDomain(args[0])
		peers = append(peers, peerURl)
		peersByOrg[org] = append(peersByOrg[org], peerURl)
	}

	for _, url := range ordererUrls {
		index := strings.Index(url, "@")
		orderers = append(orderers, url[:index])
	}
	for _, peer := range peers {
		var gossipUrl string
		args := strings.Split(peer, ":")
		_, org, _ := utils.SplitNameOrgDomain(args[0])
		orgPeers := peersByOrg[org]
		// if this org has only one peer, set this peer = gossip peer
		if len(orgPeers) == 1 {
			gossipUrl = peer
		} else {
			// this org has many peers. we choose one peer randomly, but exclude this peer
			for _, orgPeer := range orgPeers {
				if strings.Contains(peer, orgPeer) {
					continue
				}
				gossipUrl = orgPeer
				break
			}
		}
		var extraHosts []string
		extraHosts = append(extraHosts, spliceHostnameAndIP(peer, peerUrls)...)
		extraHosts = append(extraHosts, spliceHostnameAndIP(peer, ordererUrls)...)
		if err := docker_compose.GeneratePeerDockerComposeFile(dataDir, peer, gossipUrl, extraHosts); err != nil {
			return err
		}
	}
	for _, orderer := range orderers {
		var extraHosts []string
		extraHosts = append(extraHosts, spliceHostnameAndIP(orderer, peerUrls)...)
		extraHosts = append(extraHosts, spliceHostnameAndIP(orderer, ordererUrls)...)
		if err := docker_compose.GenerateOrdererDockerComposeFile(dataDir, orderer, extraHosts); err != nil {
			return err
		}
	}
	return nil
}

func spliceHostnameAndIP(excludeUrl string, urls []string) (extraHosts []string) {
	for _, url := range urls {
		if strings.Contains(url, excludeUrl) {
			continue
		}
		hostname, port, _, ip, _, _ := utils.SplitUrlParam(url)
		// if the ip is localhost or 127.0.0.1, the node will be abandoned
		isLocal, err := utils.CheckLocalIp(fmt.Sprintf("%s:%s", ip, port))
		if err != nil {
			panic(fmt.Sprintf("IP address is wrong, err=%s", err))
		}
		if isLocal {
			continue
		}

		extraHosts = append(extraHosts, fmt.Sprintf("%s:%s", hostname, ip))
	}
	return nil
}

func GenerateConnectionProfile(dataDir, channelId string, peerUrls, ordererUrls []string) error {
	var pUrls []string
	var oUrls []string

	for _, url := range peerUrls {
		hostname, port, _, ip, _, _ := utils.SplitUrlParam(url)
		pUrls = append(pUrls, fmt.Sprintf("%s:%s:%s", hostname, port, ip))
	}
	for _, url := range ordererUrls {
		hostname, port, _, ip, _, _ := utils.SplitUrlParam(url)
		oUrls = append(oUrls, fmt.Sprintf("%s:%s:%s", hostname, port, ip))
	}
	return connectionprofile.GenerateNetworkConnProfile(dataDir, channelId, pUrls, oUrls)
}

func ReadSSHConfig(dataDir string) (*sshutil.SSHUtil, error) {
	sshConfig, err := sshutil.UnmarshalSSHConfig(dataDir)
	if err != nil {
		return nil, err
	}
	sshUtil := sshutil.NewSSHUtil()
	for _, client := range sshConfig.Clients {
		if err := sshUtil.Add(client.Name, client.Username, client.Password, fmt.Sprintf("%s:%s", client.Host, client.Port), client.Type); err != nil {
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
		if client.GetNodeType() == "peer" {
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

		// TODO transfer chaincode files
	}
	return nil
}

func StartupNetwork(sshUtil *sshutil.SSHUtil, dataDir string) error {
	for name, client := range sshUtil.Clients() {
		dockerComposeFilePath := filepath.Join(dataDir, fmt.Sprintf("docker-compose-%s.yaml", strings.ReplaceAll(name, ".", "-")))
		// start nodes
		if err := client.RunCmd(fmt.Sprintf("docker-compose -f %s up -d", dockerComposeFilePath)); err != nil {
			logger.Info(err.Error())
		}
		// TODO start ca if chosen

		// TODO start couchdb if chosen
	}
	return nil
}
