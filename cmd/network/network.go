package network

import (
	"github.com/spf13/cobra"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/ca/call"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/ca/createca"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/chaincode/deploycc"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/chaincode/upgradecc"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/channel/createchannel"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/channel/joinchannel/existorgpeer"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/channel/joinchannel/neworgpeer"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/generate"
	netimport "github.com/wjbbig/fabric-distributed-tool/cmd/network/import"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/node/startnode"
	"github.com/wjbbig/fabric-distributed-tool/cmd/network/node/stopnode"
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
	networkCmd.AddCommand(startup.Cmd())
	networkCmd.AddCommand(generate.Cmd())
	networkCmd.AddCommand(shutdown.Cmd())
	networkCmd.AddCommand(deploycc.Cmd())
	networkCmd.AddCommand(stopnode.Cmd())
	networkCmd.AddCommand(upgradecc.Cmd())
	networkCmd.AddCommand(startnode.Cmd())
	networkCmd.AddCommand(neworgpeer.Cmd())
	networkCmd.AddCommand(existorgpeer.Cmd())
	networkCmd.AddCommand(createchannel.Cmd())
	networkCmd.AddCommand(createca.Cmd())
	networkCmd.AddCommand(call.Cmd())
	networkCmd.AddCommand(netimport.Cmd())
}
