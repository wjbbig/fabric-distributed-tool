package generate

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/fabricconfig"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/sshutil"
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
			return err
		}

		if err := generateCryptoConfig(); err != nil {
			return err
		}

		if err := generateConfigtx(); err != nil {
			return err
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
	if err := fabricconfig.GenerateKeyPairsAndCerts(dataDir, fabricVersion); err != nil {
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
	for _, url := range peerUrls {
		index := strings.Index(url, "@")
		peers = append(peers, url[:index])
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

	return nil
}
