package testnet

import "github.com/spf13/cobra"

const (
	peerUrl     = ""
	ordererUrl  = ""
	dataDir     = "/opt/fdt"
	channelId   = "mychannel"
	chaincodeId = "mycc"
)

var testnetCmd = &cobra.Command{
	Use:   "testnet",
	Short: "Start a test fabric network with one peer and one orderer.",
	Long:  "Start a test fabric network with one peer and one orderer.",
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}

func Cmd() *cobra.Command {
	return testnetCmd
}
