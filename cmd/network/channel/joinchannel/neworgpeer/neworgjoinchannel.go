package neworgpeer

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
)

var logger = mylogger.NewLogger()

var (
	dataDir   string
	nodeName  string
	channelId string
)

var flags *pflag.FlagSet

func init() {
	resetFlags()
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringVarP(&channelId, "channel", "c", "", "The channel which peer wants to join in")
	flags.StringVarP(&nodeName, "node", "n", "", "Node name")

}

var newOrgJoinChannelCmd = &cobra.Command{
	Use:   "neworgjoin",
	Short: "Peers of organizations that not exist in the channel join the channel",
	Long:  "Peers of organizations that not exist in the channel join the channel",
	Run: func(cmd *cobra.Command, args []string) {
		if err := utils.DoNewOrgPeerJoinChannel(dataDir, channelId, nodeName); err != nil {
			logger.Error(err.Error())
		}
	},
}

func Cmd() *cobra.Command {
	newOrgJoinChannelCmd.Flags().AddFlagSet(flags)
	return newOrgJoinChannelCmd
}
