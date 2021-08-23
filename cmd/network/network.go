package network

import (
	"github.com/spf13/cobra"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/chaincode/deploycc"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/generate"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/shutdown"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/startup"
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
	networkCmd.AddCommand(startup.Cmd())
	networkCmd.AddCommand(shutdown.Cmd())
	networkCmd.AddCommand(deploycc.Cmd())
}
