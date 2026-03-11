package main

import (
	"fmt"
	"os"

	"github.com/pinchtab/pinchtab/internal/config"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "pinchtab",
	Short: "PinchTab - Browser for AI agents",
	Long: `PinchTab provides a lightweight, API-driven way for AI agents to control 
browsers, manage tabs, and perform interactive tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action: start the server
		cfg := config.Load()
		runDashboard(cfg)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("pinchtab {{.Version}}\n")
	rootCmd.AddCommand(config.ConfigCmd)
}
