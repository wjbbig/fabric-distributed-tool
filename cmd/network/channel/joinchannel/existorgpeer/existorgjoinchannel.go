package existorgpeer

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
	channelId string
	node      string
)

var flags *pflag.FlagSet

func init() {
	resetFlags()
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&network, "network", "n", "", "The name of fabric network")
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringVarP(&channelId, "channel", "c", "", "The channel which peer wants to join in")
	flags.StringVar(&node, "node", "", "Node name")

}

var existOrgPeerJoinChannelCmd = &cobra.Command{
	Use:   "existorgjoin",
	Short: "Peers of organizations that already exist in the channel join the channel",
	Long:  "Peers of organizations that already exist in the channel join the channel",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		dataDir, err = utils.GetNetworkPathByName(dataDir, network)
		if err != nil {
			logger.Error(err.Error())
		}
		if err := utils.DoExistOrgPeerJoinChannel(dataDir, channelId, node); err != nil {
			logger.Error(err.Error())
		}
	},
}

func Cmd() *cobra.Command {
	existOrgPeerJoinChannelCmd.Flags().AddFlagSet(flags)
	return existOrgPeerJoinChannelCmd
}
