package createchannel

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
	peers     []string
	channelId string
)

var createChannelCmd = &cobra.Command{
	Use:   "createchannel",
	Short: "Create a new channel in the specified fabric network.",
	// TODO:
	Long: "Create a new channel in the specified fabric network, only support existing peer or orderer.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		dataDir, err = utils.GetNetworkPathByName(dataDir, network)
		if err != nil {
			return err
		}
		if err := utils.DoCreateChannelCommand(dataDir, channelId, peers); err != nil {
			logger.Error(err.Error())
		}
		return nil
	},
}

func Cmd() *cobra.Command {
	createChannelCmd.Flags().AddFlagSet(flags)
	return createChannelCmd
}

var flags *pflag.FlagSet

func init() {
	resetFlags()
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&network, "network", "n", "", "The name of fabric network")
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	// generate -p a -p b -p c
	flags.StringArrayVarP(&peers, "peer", "p", nil, "Hostname of fabric peers")
	flags.StringVarP(&channelId, "channelid", "c", "", "Fabric channel name")
}
