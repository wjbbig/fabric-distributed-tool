package netimport

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
	dataDir     string
	networkName string
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import a new fabric network",
	Long:  "import a new fabric network",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if !filepath.IsAbs(dataDir) {
			if dataDir, err = filepath.Abs(dataDir); err != nil {
				logger.Error(err.Error())
			}
		}
		if networkName == "" {
			logger.Error("network name cannot be empty")
			return
		}
		if err := utils.DoNetworkImport(networkName, dataDir); err != nil {
			logger.Error(err.Error())
		}
	},
}

func Cmd() *cobra.Command {
	importCmd.Flags().AddFlagSet(flags)
	return importCmd
}

var flags *pflag.FlagSet

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.StringVarP(&networkName, "network", "n", "", "The name of fabric network")
}
