package cmd

import (
	"github.com/rskv-p/mini/cmd/cmd_nats"
	"github.com/rskv-p/mini/cmd/cmd_runn"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "github.com/rskv-p/mini",
	Short: "Microservice platform",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(cmd_runn.Cmd)
	rootCmd.AddCommand(cmd_nats.Cmd)
}
