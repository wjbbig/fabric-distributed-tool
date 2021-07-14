package generate

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	docker_compose "github.com/wjbbig/fabric-distributed-tool/docker-compose"
	"github.com/wjbbig/fabric-distributed-tool/fabricconfig"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/sshutil"
	"github.com/wjbbig/fabric-distributed-tool/util"
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
	flags.StringVarP(&channelId, "channelid", "c", "", "fabric channel name")
	flags.StringVarP(&consensus, "consensus", "s", "", "orderer consensus type of fabric network")
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate necessary files of fabric.",
	Long:  "generate crypto-config.yaml, configtx.yaml and docker-compose.yaml.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO 完成方法
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
		return nil
	},
}

func Cmd() *cobra.Command {
	generateCmd.Flags().AddFlagSet(flags)
	return generateCmd
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
	var urls []string
	urls = append(urls, peerUrls...)
	urls = append(urls, ordererUrls...)
	return sshutil.GenerateSSHConfig(dataDir, urls)
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
		index := strings.Index(url, ":")
		orderers = append(orderers, url[:index])
	}
	for _, peer := range peers {
		if err := docker_compose.GeneratePeerDockerComposeFile(dataDir, peer, "", nil); err != nil {
			return err
		}
	}
	for _, orderer := range orderers {
		if err := docker_compose.GenerateOrdererDockerComposeFile(dataDir, orderer, nil); err != nil {
			return err
		}
	}
	return nil
}

//
func spliceHostnamePortAndIP(urls []string) []string {

	return nil
}
