package createca

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
	dataDir      string
	orgId        string
	enrollId     string
	enrollSecret string
)

var createCACmd = &cobra.Command{
	Use:   "createca",
	Short: "deploy a new chaincode on the specified channel",
	Long:  "deploy a new chaincode on the specified channel",
	Run: func(cmd *cobra.Command, args []string) {
		if err := utils.DoCreateCANodeForOrg(dataDir, orgId, enrollId, enrollSecret); err != nil {
			logger.Error(err.Error())
		}
	},
}

func Cmd() *cobra.Command {
	createCACmd.Flags().AddFlagSet(flags)
	return createCACmd
}

var flags *pflag.FlagSet

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringVar(&orgId, "org", "", "The specified org to run a ca node")
	flags.StringVar(&enrollId, "enrollid", "admin", "The name of registrar")
	flags.StringVar(&enrollSecret, "secret", "adminpw", "The password of registrar")
}
