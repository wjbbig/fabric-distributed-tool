package generate

import (
	"github.com/spf13/cobra"
	"github.com/wjbbig/fabric-distributed-tool/cmd/generate/bootstrap"
	"github.com/wjbbig/fabric-distributed-tool/cmd/generate/extend"
)

var (
	generateCmd = &cobra.Command{
		Use: "generate",
	}
)

func Cmd() *cobra.Command {
	generateCmd.AddCommand(bootstrap.Cmd())
	generateCmd.AddCommand(extend.Cmd())
	return generateCmd
}
