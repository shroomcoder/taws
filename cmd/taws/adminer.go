package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"taws/internal/adminer"
	"taws/internal/config"
)

var adminerCmd = &cobra.Command{
	Use:   "adminer",
	Short: "Manage Adminer web database manager",
	Long:  "Download, configure, and manage Adminer for web-based database management.",
}

var adminerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Adminer status",
	RunE:  runAdminerStatus,
}

var adminerInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Download and install Adminer",
	RunE:  runAdminerInstall,
}

func init() {
	adminerCmd.AddCommand(adminerStatusCmd)
	adminerCmd.AddCommand(adminerInstallCmd)
	rootCmd.AddCommand(adminerCmd)
}

func getAdminerManager() *adminer.Manager {
	cfg := config.DefaultConfig()
	return adminer.NewManager(
		cfg.Paths.State,
		cfg.Paths.Logs,
		cfg.Services.Adminer.Port,
	)
}

func runAdminerStatus(cmd *cobra.Command, args []string) error {
	m := getAdminerManager()

	status := m.GetStatus()
	fmt.Print(adminer.FormatStatus(status))
	return nil
}

func runAdminerInstall(cmd *cobra.Command, args []string) error {
	m := getAdminerManager()

	fmt.Println("Downloading Adminer...")

	if err := m.Download(); err != nil {
		return err
	}

	fmt.Println("Adminer installed successfully.")
	fmt.Printf("  Location: %s\n", m.GetPublicDir())

	status := m.GetStatus()
	fmt.Printf("  URL:      %s\n", status.URL)

	return nil
}
