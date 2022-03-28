package call

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/utils"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"path/filepath"
)

var logger = mylogger.NewLogger()

func init() {
	resetFlags()
}

var (
	dataDir  string
	orgId    string
	username string
	secret   string
	funcName string
)

var callFuncCmd = &cobra.Command{
	Use:   "callfunc",
	Short: "deploy a new chaincode on the specified channel",
	Long:  "deploy a new chaincode on the specified channel",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if !filepath.IsAbs(dataDir) {
			if dataDir, err = filepath.Abs(dataDir); err != nil {
				logger.Error(err.Error())
			}
		}
		if secret, err = utils.DoCallFabricCAFunc(dataDir, orgId, username, secret, funcName); err != nil {
			logger.Error(err.Error())
		}
		if secret != "" {
			logger.Infof("the secret of user %s is %s", username, secret)
		}
	},
}

func Cmd() *cobra.Command {
	callFuncCmd.Flags().AddFlagSet(flags)
	return callFuncCmd
}

var flags *pflag.FlagSet

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringVar(&orgId, "org", "", "The specified org to run a ca node")
	flags.StringVar(&username, "username", "", "The name of registrar")
	flags.StringVar(&secret, "secret", "", "The password of registrar")
	flags.StringVarP(&funcName, "func", "f", "", "support 'register' and 'revoke'")
}
