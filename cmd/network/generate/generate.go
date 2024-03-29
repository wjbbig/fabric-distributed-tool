package generate

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/network"
	"os"
	"strings"
)

var logger = mylogger.NewLogger()

var (
	dataDir        string
	peerUrls       []string
	ordererUrls    []string
	networkName    string
	fabricVersion  string
	channelId      string
	consensus      string
	ccId           string
	ccPath         string
	ccVersion      string
	ccPolicy       string
	ccInitFunc     string
	ccInitParams   []string
	ccInitRequired bool
	ccSequence     int64
	ifCouchdb      bool
	bootstrap      bool
	extend         bool
	file           bool
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
	flags.StringVarP(&fabricVersion, "version", "V", "v1.4", "Version of fabric, value can be v1.4 or v2.0")
	flags.StringVarP(&channelId, "channelid", "c", "", "Fabric channel name")
	flags.StringVarP(&consensus, "consensus", "C", "", "Orderer consensus type of fabric network")
	flags.StringVarP(&ccId, "ccid", "n", "", "Chaincode name")
	flags.StringVarP(&ccPath, "ccpath", "P", "", "Chaincode path")
	flags.StringVarP(&ccVersion, "ccversion", "v", "", "chaincode version")
	flags.StringVarP(&ccPolicy, "ccpolicy", "r", "", "chaincode policy")
	flags.StringVarP(&ccInitFunc, "ccinitfunc", "f", "", "chaincode initial function")
	flags.StringSliceVar(&ccInitParams, "param", []string{}, "chaincode initial params")
	flags.StringVarP(&networkName, "network", "w", "", "Fabric network name")
	flags.BoolVar(&ifCouchdb, "couchdb", false, "If use couchdb")
	flags.BoolVar(&ccInitRequired, "initcc", false, "If chaincode needs init")
	flags.Int64VarP(&ccSequence, "seq", "s", 1, "Chaincode sequence for fabric v2.0")
	flags.BoolVar(&bootstrap, "bootstrap", false, "Initialize the network configuration file")
	flags.BoolVar(&extend, "extend", false, "Extend the network configuration file")
	flags.BoolVar(&file, "file", false, "Using exist networkconfig.yaml to generate network")
}

var (
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "generate necessary files of fabric.",
		Long:  "generate crypto-config.yaml, configtx.yaml and docker-compose.yaml.",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			dataDir, err = utils.GetNetworkPathByName(dataDir, "")
			if err != nil {
				logger.Error(err.Error())
				return
			}
			if _, err = os.Stat(dataDir); err != nil {
				_ = os.MkdirAll(dataDir, 0755)
			} else {
				config, err := network.UnmarshalNetworkConfig(dataDir)
				if err == nil {
					networkName = config.Name
					if err := utils.NetworkExist(networkName); err != nil {
						logger.Errorf(err.Error())
						return
					}
					if !file {
						var flag string
						fmt.Printf("an existing network configuration file was found in the %s，overwrite all? [y/n]\t", dataDir)
						fmt.Scanf("%s", &flag)
						if strings.ToUpper(flag) == "Y" {
							if err := os.RemoveAll(dataDir); err != nil {
								logger.Error(err.Error())
								return
							}
						}
					}
				}
			}

			switch {
			case bootstrap:
				if err := utils.DoGenerateBootstrapCommand(dataDir, networkName, channelId, consensus, ccId, ccPath, ccVersion,
					ccInitFunc, ccInitParams, ccPolicy, ccInitRequired, ccSequence, ifCouchdb, peerUrls,
					ordererUrls, fabricVersion, file); err != nil {
					logger.Error(err.Error())
				}
			case extend:
				if err := utils.DoExtendNodeCommand(dataDir, ifCouchdb, peerUrls, ordererUrls); err != nil {
					logger.Error(err.Error())
				}
			default:
				logger.Error("bootstrap or extend must be chosen one")
			}
		},
	}
)

func Cmd() *cobra.Command {
	generateCmd.Flags().AddFlagSet(flags)
	return generateCmd
}
