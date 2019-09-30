package cmd

import (
	"github.com/itross/sgulgate/internal/gateway"
	"github.com/spf13/cobra"
)

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "starts the API Gateway",
	Long:  "This command configures and starts the API Gateway",
	Run: func(cmd *cobra.Command, args []string) {
		start(args)
	},
}

func init() {
	RootCmd.AddCommand(startCommand)
}

func start(args []string) {
	gateway.New().Start()
}
