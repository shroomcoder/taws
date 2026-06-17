package main

import (
	"github.com/spf13/cobra"

	"taws/internal/tui"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show the TAWS dashboard",
	Long:  "Displays a real-time terminal dashboard with service status, projects, and system information.",
	RunE:  runDashboard,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

func runDashboard(cmd *cobra.Command, args []string) error {
	return tui.RunDashboard()
}
