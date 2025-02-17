package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run pulpod",
	Run: func(_ *cobra.Command, _ []string) {
		slog.Info("run pulpod, run")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
