package generate

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	dataDir       string
	peerUrls      []string
	ordererUrls   []string
	fabricVersion string
	channelId     string
)

var flags *pflag.FlagSet

func init() {
	resetFlags()
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVarP(&dataDir, "datadir", "d", "", "Path to file containing fabric needed")
	// ./fdt generate -p a -p b -p c
	flags.StringArrayVarP(&peerUrls, "peerurls", "p", nil, "Urls of peers, splitting by ','")
	flags.StringArrayVarP(&ordererUrls, "ordererurls", "o", nil, "Urls of orderers, splitting by ','")
	flags.StringVarP(&fabricVersion, "version", "v", "2.0", "Version of fabric")
	flags.StringVarP(&channelId, "channelid", "c", "", "fabric channel name")
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate necessary files of fabric.",
	Long:  "generate crypto-config.yaml, configtx.yaml and docker-compose.yaml.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO 完成方法
		return nil
	},
}

func Cmd() *cobra.Command {
	generateCmd.Flags().AddFlagSet(flags)
	return generateCmd
}

func generateCryptoConfig() {

}

func generateConfigtx() {

}
