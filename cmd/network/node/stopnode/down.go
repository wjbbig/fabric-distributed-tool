package stopnode

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
)

var (
	logger    = mylogger.NewLogger()
	dataDir   string
	nodeNames []string
	flags     *pflag.FlagSet
)

func init() {
	resetFlags()
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	// generate -p a -p b -p c
	flags.StringArrayVarP(&nodeNames, "name", "n", nil, "Hostname of fabric node")
}

var stopNodeCmd = &cobra.Command{
	Use:   "stopnode",
	Short: "Stop a node with the specified name",
	Long:  "Stop a node with the specified name",
	Run: func(cmd *cobra.Command, args []string) {
		if err := utils.DoStopNodeCmd(dataDir, nodeNames...); err != nil {
			logger.Error(err.Error())
		}
	},
}

func Cmd() *cobra.Command {
	stopNodeCmd.Flags().AddFlagSet(flags)
	return stopNodeCmd
}
