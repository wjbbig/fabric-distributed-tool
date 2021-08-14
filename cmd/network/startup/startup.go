package startup

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
)

var logger = mylogger.NewLogger()

// TODO delete some useless params and write them into the ssh config file.
var (
	dataDir          string
	startOnly        bool
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
		if err := utils.DoStartupCommand(dataDir, startOnly); err != nil {
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
	flags.BoolVar(&startOnly, "startonly", false, "Just start the docker container")
}
