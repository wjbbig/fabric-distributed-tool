package bootstrap

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"os"
)

var logger = mylogger.NewLogger()

var (
	dataDir       string
	peerUrls      []string
	ordererUrls   []string
	fabricVersion string
	channelId     string
	consensus     string
	ifCouchdb     bool
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
	flags.StringVarP(&consensus, "consensus", "C", "", "Orderer consensus type of fabric network")
	flags.BoolVar(&ifCouchdb, "couchdb", false, "If use couchdb")
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
		if err := utils.GenerateSSHConfig(dataDir, peerUrls, ordererUrls); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := utils.GenerateCryptoConfig(dataDir, peerUrls, ordererUrls); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := utils.GenerateConfigtx(dataDir, consensus, channelId, peerUrls, ordererUrls); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := utils.GenerateDockerCompose(dataDir, peerUrls, ordererUrls, ifCouchdb); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := utils.GenerateConnectionProfile(dataDir, channelId, peerUrls, ordererUrls); err != nil {
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
