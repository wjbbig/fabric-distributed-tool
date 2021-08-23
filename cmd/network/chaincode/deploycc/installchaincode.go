package deploycc

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	initParam    string
)

var installChaincodeCmd = &cobra.Command{
	Use:   "deploycc",
	Short: "deploy a new chaincode on the specified channel",
	Long:  "deploy a new chaincode on the specified channel",
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}

func Cmd() *cobra.Command {
	return installChaincodeCmd
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
}
