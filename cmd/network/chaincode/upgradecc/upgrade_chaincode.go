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
	channelId    string
	dataDir      string
	ccId         string
	ccPath       string
	ccVersion    string
	ccPolicy     string
	initRequired bool
	initFunc     string
	initParam    string
)

var upgradeChaincodeCmd = &cobra.Command{
	Use:   "upgradecc",
	Short: "upgrade a exist chaincode on the specified channel",
	Long:  "upgrade a exist chaincode on the specified channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := utils.DoUpgradeccCmd(dataDir, channelId, ccId, ccPath, ccVersion, ccPolicy, initFunc, initParam, initRequired); err != nil {
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
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringVarP(&channelId, "channelid", "c", "", "The specified channel to deploy new chaincode")
	flags.StringVarP(&ccId, "ccid", "n", "", "The name of new chaincode")
	flags.StringVarP(&ccPath, "ccpath", "p", "", "The path of new chaincode")
	flags.StringVarP(&ccVersion, "ccversion", "v", "", "The version of new chaincode")
	flags.BoolVarP(&initRequired, "initrequired", "r", false, "If the new chaincode needs initialization")
	flags.StringVarP(&ccPolicy, "ccpolicy", "P", "", "The endorsement policy of new chaincode")
	flags.StringVarP(&initParam, "initparam", "i", "", "The initial param of new chaincode")
	flags.StringVarP(&initFunc, "initfunc", "f", "", "The initial function of new chaincode")
}
