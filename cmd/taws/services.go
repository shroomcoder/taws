package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"taws/internal/config"
	"taws/internal/ports"
	"taws/internal/services"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start TAWS services",
	Long:  "Starts Nginx, PHP-FPM, and MariaDB services. Automatically resolves port conflicts.",
	RunE:  runStart,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop TAWS services",
	Long:  "Stops all running TAWS services.",
	RunE:  runStop,
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart TAWS services",
	Long:  "Restarts all TAWS services.",
	RunE:  runRestart,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show TAWS service status",
	Long:  "Displays the current status of all TAWS services.",
	RunE:  runStatus,
}

// loadConfigPath finds the TAWS config file, falling back to DefaultConfig if not found.
func loadConfigPath() (*config.Config, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return config.DefaultConfig(), "", nil
	}
	cfgPath := filepath.Join(home, ".taws", "configs", "taws.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return config.DefaultConfig(), "", nil
	}
	return cfg, cfgPath, nil
}

// resolveAndReport checks all service ports and saves updated config if any port was reassigned.
func resolveAndReport(cfg *config.Config, cfgPath string) {
	if cfgPath == "" {
		return
	}

	origPorts := map[string]int{
		"nginx":   cfg.Services.Nginx.Port,
		"php-fpm": cfg.Services.PHP.Port,
		"mariadb": cfg.Services.MariaDB.Port,
		"adminer": cfg.Services.Adminer.Port,
	}

	assigned := ports.ResolvePorts(cfg)

	changed := false
	for name, newPort := range assigned {
		if newPort > 0 && origPorts[name] != newPort {
			changed = true
			fmt.Printf("  %s: port %d -> %d\n", name, origPorts[name], newPort)
		}
	}
	if !changed {
		return
	}

	_ = config.Save(cfg, cfgPath)
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, cfgPath, _ := loadConfigPath()

	fmt.Println("Checking ports...")
	resolveAndReport(cfg, cfgPath)

	m := services.NewManager()
	fmt.Println("Starting TAWS services...")

	errs := m.StartAll()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Printf("  Warning: %v\n", err)
		}
	}

	fmt.Println("\nChecking status...")
	statuses := m.StatusAll()
	fmt.Print(services.FormatStatus(statuses))

	return nil
}

func runStop(cmd *cobra.Command, args []string) error {
	m := services.NewManager()
	fmt.Println("Stopping TAWS services...")

	errs := m.StopAll()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Printf("  Warning: %v\n", err)
		}
	}

	fmt.Println("\nChecking status...")
	statuses := m.StatusAll()
	fmt.Print(services.FormatStatus(statuses))

	return nil
}

func runRestart(cmd *cobra.Command, args []string) error {
	cfg, cfgPath, _ := loadConfigPath()

	fmt.Println("Checking ports...")
	resolveAndReport(cfg, cfgPath)

	m := services.NewManager()
	fmt.Println("Restarting TAWS services...")

	for _, svc := range []services.ServiceName{services.Nginx, services.PHP, services.MariaDB} {
		if err := m.Restart(svc); err != nil {
			fmt.Printf("  Warning: %v\n", err)
		}
	}

	fmt.Println("\nChecking status...")
	statuses := m.StatusAll()
	fmt.Print(services.FormatStatus(statuses))

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	m := services.NewManager()
	statuses := m.StatusAll()
	fmt.Print(services.FormatStatus(statuses))

	return nil
}
