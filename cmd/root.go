package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

// NewRootCommand builds the command tree. Tests can call this without exiting.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "gdlint",
		Short:         "Lint Grafana dashboard JSON files",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newCheckCommand())

	return rootCmd
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
