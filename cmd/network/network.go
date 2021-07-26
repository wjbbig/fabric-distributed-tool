package network

import (
	"github.com/spf13/cobra"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/generate"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/generate/bootstrap"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/startup"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/testnet"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Defines a series of operations of fabric",
	Long:  "Defines a series of operations of fabric",
}

func Cmd() *cobra.Command {
	return networkCmd
}

func init() {
	networkCmd.AddCommand(generate.Cmd())
	networkCmd.AddCommand(bootstrap.Cmd())
	networkCmd.AddCommand(startup.Cmd())
	networkCmd.AddCommand(testnet.Cmd())
}
