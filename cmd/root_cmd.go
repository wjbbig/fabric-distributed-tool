package main

import (
	"github.com/spf13/cobra"
	"github.com/wjbbig/fabric-distributed-tool/cmd/generate"
	"github.com/wjbbig/fabric-distributed-tool/cmd/version"
	"os"
)

// 总命令
var rootCmd = &cobra.Command{
	Use: "fdt",
}

func main() {
	//mainFlags := rootCmd.PersistentFlags()
	rootCmd.AddCommand(version.Cmd())
	rootCmd.AddCommand(generate.Cmd())
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
