package startup

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
)

var logger = mylogger.NewLogger()

var (
	network   string
	dataDir   string
	startOnly bool
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
		var err error
		dataDir, err = utils.GetNetworkPathByName(dataDir, network)
		if err != nil {
			return err
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
	flags.StringVarP(&network, "network", "n", "", "The name of fabric network")
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.BoolVar(&startOnly, "startonly", false, "Only starting the docker container")
}
