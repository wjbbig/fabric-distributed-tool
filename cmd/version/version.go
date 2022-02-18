package version

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print fabric-distributed-tool version.",
	Long:  "Print fabric-distributed-tool version.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("trailing args detected")
		}
		fmt.Println(getInfo())
		return nil
	},
}

func Cmd() *cobra.Command {
	return versionCmd
}

func getInfo() string {
	return fmt.Sprintf("fdt version: v0.1\nsupported fabric version: v1.4.x|v2.x")
}
