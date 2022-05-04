package shutdown

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
)

var logger = mylogger.NewLogger()
var flags *pflag.FlagSet
var dataDir string
var network string

func init() {
	resetFlags()
}

var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "shutdown an exist fabric network",
	Long:  "shutdown an exist fabric network",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		dataDir, err = utils.GetNetworkPathByName(dataDir, network)
		if err != nil {
			return err
		}
		if err := utils.DoShutdownCommand(dataDir); err != nil {
			logger.Error(err.Error())
			return nil
		}
		return nil
	},
}

func Cmd() *cobra.Command {
	shutdownCmd.Flags().AddFlagSet(flags)
	return shutdownCmd
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&network, "network", "n", "", "The name of fabric network")
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
}
