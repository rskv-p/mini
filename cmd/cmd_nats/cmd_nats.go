package cmd_nats

import (
	"context"
	"fmt"

	"github.com/rskv-p/mini/servs/s_nats/nats_client"

	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "nats",
	Short: "Interact with NATS service",
}

var echoCmd = &cobra.Command{
	Use:   "echo [message]",
	Short: "Send echo message to NATS",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			return fmt.Errorf("nats connect: %w", err)
		}
		defer nc.Close()

		client := nats_client.New(nc, "s_nats")
		reply, err := client.Echo(context.Background(), args[0])
		if err != nil {
			return fmt.Errorf("echo failed: %w", err)
		}

		fmt.Println("REPLY:", reply)
		return nil
	},
}

func init() {
	Cmd.AddCommand(echoCmd)
}
