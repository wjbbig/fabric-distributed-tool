package startup

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	mylogger "github.com/wjbbig/fabric-distributed-tool/logger"
	"github.com/wjbbig/fabric-distributed-tool/sshutil"
	"path/filepath"
)

var logger = mylogger.NewLogger()
var (
	dataDir string
	ca      bool
	stateDB string
)

func init() {
	resetFlags()
}

// 启动fabric网络
var startupCmd = &cobra.Command{
	Use:   "startup",
	Short: "start the fabric network",
	Long:  "start the fabric network",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dataDir == "" {
			logger.Error("datadir is not specified")
		}
		sshUtil, err := readSSHConfig()
		if err != nil {
			logger.Error(err.Error())
			return nil
		}

		return nil
	},
}

func Cmd() *cobra.Command {
	startupCmd.Flags().AddFlagSet(flags)
	return startupCmd
}

var flags *pflag.FlagSet

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	flags.BoolVarP(&ca, "ca", "c", false, "If start the fabric ca container")
	flags.StringVarP(&stateDB, "statedb", "s", "leveldb", "Which type of statedb that fabric uses")
}

func readSSHConfig() (*sshutil.SSHUtil, error) {
	configPath := filepath.Join(dataDir, "ssh-config.yaml")
	sshConfig, err := sshutil.UnmarshalSSHConfig(configPath)
	if err != nil {
		return nil, err
	}
	sshUtil := sshutil.NewSSHUtil()
	for _, client := range sshConfig.Clients {
		if err := sshUtil.Add(client.Name, client.Username, client.Password, fmt.Sprintf("%s:%s", client.Host, client.Port)); err != nil {
			return nil, err
		}
	}
	return nil, nil
}
