package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"taws/internal/config"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Completely remove TAWS and all its files",
	Long:  "Removes all TAWS configuration, state, logs, backups, caches, and installed plugins. User projects in ~/web/ are preserved by default.",
	RunE:  runUninstall,
}

var uninstallKeepProjects bool

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallKeepProjects, "keep-projects", true, "Keep user projects in ~/web/")
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	cfg := config.DefaultConfig()

	fmt.Println("TAWS Uninstaller")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	fmt.Println("The following will be removed:")
	fmt.Printf("  - %s (configs, state, logs, backups, certs, plugins, cache)\n", cfg.Paths.TawsRoot)
	fmt.Println()

	if uninstallKeepProjects {
		fmt.Printf("  - User projects in %s will be KEPT\n", cfg.Paths.WebRoot)
	} else {
		fmt.Printf("  - User projects in %s will be REMOVED\n", cfg.Paths.WebRoot)
	}

	fmt.Println()
	fmt.Print("Are you sure you want to continue? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "yes" && response != "y" {
		fmt.Println("Uninstall cancelled.")
		return nil
	}

	fmt.Println()
	fmt.Println("Removing TAWS files...")

	removed := 0

	dirsToRemove := []string{
		filepath.Join(cfg.Paths.TawsRoot, "configs"),
		filepath.Join(cfg.Paths.TawsRoot, "state"),
		filepath.Join(cfg.Paths.TawsRoot, "logs"),
		filepath.Join(cfg.Paths.TawsRoot, "backups"),
		filepath.Join(cfg.Paths.TawsRoot, "certs"),
		filepath.Join(cfg.Paths.TawsRoot, "plugins"),
		filepath.Join(cfg.Paths.TawsRoot, "cache"),
		cfg.Paths.TawsRoot,
	}

	for _, dir := range dirsToRemove {
		if _, err := os.Stat(dir); err == nil {
			if err := os.RemoveAll(dir); err != nil {
				fmt.Printf("  Warning: Could not remove %s: %v\n", dir, err)
			} else {
				fmt.Printf("  Removed: %s\n", dir)
				removed++
			}
		}
	}

	if !uninstallKeepProjects {
		if _, err := os.Stat(cfg.Paths.WebRoot); err == nil {
			if err := os.RemoveAll(cfg.Paths.WebRoot); err != nil {
				fmt.Printf("  Warning: Could not remove %s: %v\n", cfg.Paths.WebRoot, err)
			} else {
				fmt.Printf("  Removed: %s\n", cfg.Paths.WebRoot)
				removed++
			}
		}
	}

	tawsBin, _ := filepath.Abs("taws")
	if _, err := os.Stat(tawsBin); err == nil {
		os.Remove(tawsBin)
		fmt.Printf("  Removed: %s\n", tawsBin)
		removed++
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))

	if removed > 0 {
		fmt.Printf("TAWS uninstalled successfully. %d items removed.\n", removed)
	} else {
		fmt.Println("TAWS uninstalled successfully. No files were found to remove.")
	}

	if uninstallKeepProjects {
		fmt.Printf("\nYour projects in %s have been preserved.\n", cfg.Paths.WebRoot)
	}

	return nil
}
