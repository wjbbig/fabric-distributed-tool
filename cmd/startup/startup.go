package startup

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	dataDir string
	ca      bool
	stateDB string
)

// 启动fabric网络
var startupCmd = &cobra.Command{
	Use:   "startup",
	Short: "start the fabric network",
	Long:  "start the fabric network",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Cmd() *cobra.Command {
	return startupCmd
}

var flags *pflag.FlagSet

func resetFlags() {

}
