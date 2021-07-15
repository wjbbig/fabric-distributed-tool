package extend

import "github.com/spf13/cobra"

var extendCmd = &cobra.Command{
	Use:   "extend",
	Short: "",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}

func Cmd() *cobra.Command {
	return extendCmd
}
