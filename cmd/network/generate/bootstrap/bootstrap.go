package bootstrap

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/connectionprofile"
	docker_compose "github.com/wjbbig/fabric-distributed-tool/docker-compose"
	"github.com/wjbbig/fabric-distributed-tool/fabricconfig"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/sshutil"
	"github.com/wjbbig/fabric-distributed-tool/util"
	"os"
	"strings"
)

var logger = mylogger.NewLogger()

var (
	dataDir       string
	peerUrls      []string
	ordererUrls   []string
	fabricVersion string
	channelId     string
	consensus     string
)

// peer0.org1.example.com:7050@username@127.0.0.1:22:password
var flags *pflag.FlagSet

func init() {
	resetFlags()
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	// generate -p a -p b -p c
	flags.StringArrayVarP(&peerUrls, "peerurls", "p", nil, "Urls of fabric peers")
	flags.StringArrayVarP(&ordererUrls, "ordererurls", "o", nil, "Urls of fabric orderers")
	flags.StringVarP(&fabricVersion, "version", "v", "1.4", "Version of fabric, value can be 1.4 or 2.0")
	flags.StringVarP(&channelId, "channelid", "c", "", "Fabric channel name")
	flags.StringVarP(&consensus, "consensus", "s", "", "Orderer consensus type of fabric network")
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "generate necessary files of fabric.",
	Long:  "generate crypto-config.yaml, configtx.yaml and docker-compose.yaml.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dataDir == "" {
			logger.Error("datadir is not specified")
		}
		if _, err := os.Stat(dataDir); err != nil {
			_ = os.MkdirAll(dataDir, 0755)
		}
		if err := generateSSHConfig(); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := generateCryptoConfig(); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := generateConfigtx(); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := generateDockerCompose(); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := generateConnectionProfile(); err != nil {
			logger.Error(err.Error())
			return nil
		}
		return nil
	},
}

func Cmd() *cobra.Command {
	bootstrapCmd.Flags().AddFlagSet(flags)
	return bootstrapCmd
}

func generateCryptoConfig() error {
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

func generateSSHConfig() error {
	var clients []sshutil.Client
	for _, url := range peerUrls {
		clients = append(clients, sshutil.NewClient(url, "peer"))
	}
	for _, url := range ordererUrls {
		clients = append(clients, sshutil.NewClient(url, "orderer"))
	}
	return sshutil.GenerateSSHConfig(dataDir, clients)
}

func generateConfigtx() error {
	var peers []string
	var orderers []string
	var orgs []string
	for _, url := range peerUrls {
		index := strings.Index(url, "@")
		peerURl := url[:index]
		args := strings.Split(peerURl, ":")
		_, org, _ := util.SplitNameOrgDomain(args[0])
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

func generateDockerCompose() error {
	var peers []string
	var orderers []string
	peersByOrg := make(map[string][]string)
	for _, url := range peerUrls {
		index := strings.Index(url, "@")
		peerURl := url[:index]
		args := strings.Split(peerURl, ":")
		_, org, _ := util.SplitNameOrgDomain(args[0])
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
		_, org, _ := util.SplitNameOrgDomain(args[0])
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
		if err := docker_compose.GenerateOrdererDockerComposeFile(dataDir, orderer, nil); err != nil {
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
		hostname, port, _, ip, _, _ := util.SplitUrlParam(url)
		isLocal, err := util.CheckLocalIp(fmt.Sprintf("%s:%s", ip, port))
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

func generateConnectionProfile() error {
	var pUrls []string
	var oUrls []string

	for _, url := range peerUrls {
		hostname, port, _, ip, _, _ := util.SplitUrlParam(url)
		pUrls = append(pUrls, fmt.Sprintf("%s:%s:%s", hostname, port, ip))
	}
	for _, url := range ordererUrls {
		hostname, port, _, ip, _, _ := util.SplitUrlParam(url)
		oUrls = append(oUrls, fmt.Sprintf("%s:%s:%s", hostname, port, ip))
	}
	return connectionprofile.GenerateNetworkConnProfile(dataDir, channelId, pUrls, oUrls)
}
