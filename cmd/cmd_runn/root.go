package cmd_runn

import (
	"github.com/spf13/cobra"
)

// RootCmd is the root command for the background process manager CLI
// It serves as the entry point for the CLI, with subcommands available for managing background processes
var RootCmd = &cobra.Command{
	Use:   "runn",                               // The main command name for the CLI
	Short: "CLI for background process manager", // A brief description of the CLI's purpose
}
