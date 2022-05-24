package installcc

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
)

var logger = mylogger.NewLogger()

func init() {
	resetFlags()
}

var (
	network   string
	channelId string
	dataDir   string
	ccId      string
	peer      string
)

var installChaincodeCmd = &cobra.Command{
	Use:   "installcc",
	Short: "deploy a new chaincode on the specified channel",
	Long:  "deploy a new chaincode on the specified channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		dataDir, err = utils.GetNetworkPathByName(dataDir, network)
		if err != nil {
			return err
		}
		if err := utils.DoInstallOnlyCmd(dataDir, channelId, ccId, peer); err != nil {
			logger.Error(err.Error())
		}
		return nil
	},
}

func Cmd() *cobra.Command {
	installChaincodeCmd.Flags().AddFlagSet(flags)
	return installChaincodeCmd
}

var flags *pflag.FlagSet

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&network, "network", "n", "", "The name of fabric network")
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringVarP(&channelId, "channelid", "c", "", "The specified channel to deploy new chaincode")
	flags.StringVar(&ccId, "ccid", "", "The name of new chaincode")
	flags.StringVarP(&peer, "peer", "p", "", "The name of peer")
}
