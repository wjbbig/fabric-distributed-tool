package extend

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
)

var logger = mylogger.NewLogger()

var (
	dataDir     string
	peerUrls    []string
	ordererUrls []string
	ifCouchdb   bool
)

// peer0.org1.example.com:7050@username@127.0.0.1:22:password
var flags *pflag.FlagSet

func init() {
	resetFlags()
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringArrayVarP(&peerUrls, "peerurls", "p", nil, "Urls of fabric peers")
	flags.StringArrayVarP(&ordererUrls, "ordererurls", "o", nil, "Urls of fabric orderers")
	flags.BoolVar(&ifCouchdb, "couchdb", false, "If use couchdb")
}

var extendCmd = &cobra.Command{
	Use:   "extend",
	Short: "extend peer or orderer node.",
	Long:  "extend peer or orderer node.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dataDir == "" {
			logger.Error("datadir is not specified")
		}

		return nil
	},
}

func Cmd() *cobra.Command {
	return extendCmd
}
