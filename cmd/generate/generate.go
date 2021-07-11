package generate

import "github.com/spf13/cobra"

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate necessary files of fabric.",
	Long:  "generate crypto-config.yaml, configtx.yaml and docker-compose.yaml.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Cmd() *cobra.Command {
	return generateCmd
}

func help() {

}
