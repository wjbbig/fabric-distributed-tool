package upgradecc

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
	network      string
	channelId    string
	dataDir      string
	ccId         string
	ccPath       string
	ccVersion    string
	ccPolicy     string
	redeploy     bool
	initRequired bool
	initFunc     string
	initParam    string
)

var upgradeChaincodeCmd = &cobra.Command{
	Use:   "upgradecc",
	Short: "upgrade a exist chaincode on the specified channel",
	Long:  "upgrade a exist chaincode on the specified channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		dataDir, err = utils.GetNetworkPathByName(dataDir, network)
		if err != nil {
			return err
		}
		if err := utils.DoUpgradeccCmd(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParam, initRequired, redeploy); err != nil {
			logger.Error(err.Error())
		}
		return nil
	},
}

func Cmd() *cobra.Command {
	upgradeChaincodeCmd.Flags().AddFlagSet(flags)
	return upgradeChaincodeCmd
}

var flags *pflag.FlagSet

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&network, "network", "n", "", "The name of fabric network")
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringVarP(&channelId, "channelid", "c", "", "The specified channel to deploy new chaincode")
	flags.StringVar(&ccId, "ccid", "", "The name of new chaincode")
	flags.StringVarP(&ccPath, "ccpath", "p", "", "The path of new chaincode")
	flags.StringVarP(&ccVersion, "ccversion", "v", "", "The version of new chaincode")
	flags.BoolVar(&initRequired, "initcc", false, "If the new chaincode needs initialization")
	flags.BoolVar(&redeploy, "redeploy", false, "If the new chaincode needs redeploy. This option is only used by fabric v2.0, "+
		"if we only update chaincode policy, redeploy should be false")
	flags.StringVar(&ccPolicy, "policy", "", "The endorsement policy of new chaincode")
	flags.StringVar(&initParam, "param", "", "The initial param of new chaincode")
	flags.StringVar(&initFunc, "func", "", "The initial function of new chaincode")
}
