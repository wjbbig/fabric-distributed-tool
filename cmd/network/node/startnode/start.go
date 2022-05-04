package startnode

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

var startNodeCmd = &cobra.Command{
	Use:   "startnode",
	Short: "Start a node with the specified name",
	Long:  "Start a node with the specified name",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		dataDir, err = utils.GetNetworkPathByName(dataDir, network)
		if err != nil {
			logger.Error(err.Error())
		}
		if err := utils.DoStartNodeCmd(dataDir, peers...); err != nil {
			logger.Error(err.Error())
		}
	},
}

func Cmd() *cobra.Command {
	startNodeCmd.Flags().AddFlagSet(flags)
	return startNodeCmd
}
