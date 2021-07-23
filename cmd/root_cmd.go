package main

import (
	"github.com/wjbbig/fabric-distributed-tool/cmd/network"
	"os"

	"github.com/spf13/cobra"
	"github.com/wjbbig/fabric-distributed-tool/cmd/version"
)

// 总命令
var rootCmd = &cobra.Command{
	Use: "fdt",
}

func main() {
	rootCmd.AddCommand(version.Cmd())
	rootCmd.AddCommand(network.Cmd())
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
