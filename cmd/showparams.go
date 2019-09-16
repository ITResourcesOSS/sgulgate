package cmd

import (
	"github.com/ITResourcesOSS/sgulgate/internal/gateway"
	"github.com/spf13/cobra"
)

var showparamsCommand = &cobra.Command{
	Use:   "show",
	Short: "shows gateway params",
	Long:  "This command prints out all the gateway params after configuration",
	Run: func(cmd *cobra.Command, args []string) {
		show(args)
	},
}

func init() {
	RootCmd.AddCommand(showparamsCommand)
}

func show(args []string) {
	gw := gateway.New()
	gw.PrintParams()
	gw.PrintApis()
}
