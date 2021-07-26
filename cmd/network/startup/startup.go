package startup

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
)

var logger = mylogger.NewLogger()
var (
	dataDir       string
	ca            bool
	stateDB       string
	channelId     string
	chaincodeId   string
	chaincodePath string
	initParam     string
	startOnly     bool
)

func init() {
	resetFlags()
}

// 启动fabric网络
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

		if err := utils.TransferFilesByPeerName(sshUtil, dataDir); err != nil {
			logger.Error(err.Error())
			return nil
		}
		if err := utils.StartupNetwork(sshUtil, dataDir); err != nil {
			logger.Error(err.Error())
			return nil
		}
		// if only starting the fabric docker container
		if startOnly {
			return nil
		}

		// todo
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
	flags.StringVar(&dataDir, "datadir", "", "Path to file containing fabric needed")
	flags.BoolVar(&ca, "ca", false, "If start the fabric ca container")
	flags.StringVar(&stateDB, "statedb", "leveldb", "Which type of statedb that fabric uses")
	flags.BoolVar(&startOnly, "startonly", false, "Just start the docker container")
	flags.StringVar(&channelId, "channel", "", "Channel Name")
	flags.StringVar(&chaincodeId, "chaincodeid", "", "Chaincode Name")
	flags.StringVar(&chaincodePath, "path", "", "Chaincode path")
	flags.StringVar(&initParam, "initparam", "", "Chaincode Init params")
}
