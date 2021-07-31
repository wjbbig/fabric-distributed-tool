package startup

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/sdkutil"
	"path/filepath"
)

var logger = mylogger.NewLogger()

// TODO delete some useless params and write them into the ssh config file.
var (
	dataDir          string
	ca               bool
	couchdb          bool
	channelId        string
	chaincodeId      string
	chaincodePath    string
	chaincodeVersion string
	policy           string
	initParam        string
	startOnly        bool
)

const (
	connectionConfigFileName = "connection-config.yaml"
)

func init() {
	resetFlags()
}

// startupCmd starts the fabric network
var startupCmd = &cobra.Command{
	Use:   "startup",
	Short: "start the fabric network",
	Long:  "start the fabric network",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dataDir == "" {
			logger.Error("datadir is not specified")
		}
		sshUtil, err := utils.ReadSSHConfig(dataDir)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		defer sshUtil.CloseAll()

		if err := utils.TransferFilesByPeerName(sshUtil, dataDir, couchdb); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := utils.StartupNetwork(sshUtil, dataDir, couchdb); err != nil {
			logger.Error(err.Error())
			return nil
		}
		// if only starting the fabric docker container
		if startOnly {
			return nil
		}

		sdk, err := sdkutil.NewFabricSDKDriver(filepath.Join(dataDir, connectionConfigFileName))
		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		defer sdk.Close()

		// create channel
		if err := utils.CreateChannel(sshUtil, dataDir, channelId, sdk); err != nil {
			logger.Error(err.Error())
			return nil
		}
		// join channel
		if err := utils.JoinChannel(sshUtil, channelId, sdk); err != nil {
			logger.Error(err.Error())
			return nil
		}
		// install chaincode
		if err := utils.InstallCC(sshUtil, chaincodeId, chaincodePath, chaincodeVersion, channelId, sdk); err != nil {
			logger.Error(err.Error())
			return nil
		}
		// InstantiateCC
		if err := utils.InstantiateCC(sshUtil, chaincodeId, chaincodePath, chaincodeVersion, channelId,
			policy, initParam, sdk); err != nil {
			logger.Error(err.Error())
			return nil
		}
		logger.Info("DONE!")
		return nil
	},
}

func Cmd() *cobra.Command {
	startupCmd.Flags().AddFlagSet(flags)
	return startupCmd
}

var flags *pflag.FlagSet

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.BoolVar(&ca, "ca", false, "If start the fabric ca container")
	flags.BoolVar(&couchdb, "couchdb", false, "If use couchdb")
	flags.BoolVar(&startOnly, "startonly", false, "Just start the docker container")
	flags.StringVarP(&channelId, "channelid", "c", "", "Channel Name")
	flags.StringVarP(&chaincodeId, "ccid", "n", "", "Chaincode Name")
	flags.StringVarP(&chaincodePath, "ccpath", "p", "", "Chaincode path")
	flags.StringVarP(&chaincodeVersion, "ccver", "v", "", "Chaincode version")
	flags.StringVar(&policy, "policy", "", "Chaincode policy")
	flags.StringVar(&initParam, "params", "init", "Chaincode Init params")
}
