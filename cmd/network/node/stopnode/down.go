package stopnode

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
)

var (
	logger  = mylogger.NewLogger()
	dataDir string
	network string
	peers   []string
	flags   *pflag.FlagSet
)

func init() {
	resetFlags()
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&network, "network", "n", "", "The name of fabric network")
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringArrayVarP(&peers, "peers", "p", nil, "Hostname of fabric node")
}

var stopNodeCmd = &cobra.Command{
	Use:   "stopnode",
	Short: "Stop a node with the specified name",
	Long:  "Stop a node with the specified name",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		dataDir, err = utils.GetNetworkPathByName(dataDir, network)
		if err != nil {
			logger.Error(err.Error())
		}
		if err := utils.DoStopNodeCmd(dataDir, peers...); err != nil {
			logger.Error(err.Error())
		}
	},
}

func Cmd() *cobra.Command {
	stopNodeCmd.Flags().AddFlagSet(flags)
	return stopNodeCmd
}
