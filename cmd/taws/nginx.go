package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"taws/internal/config"
	"taws/internal/nginx"
)

var nginxCmd = &cobra.Command{
	Use:   "nginx",
	Short: "Manage Nginx virtual hosts",
	Long:  "Generate, enable, disable, and list Nginx virtual host configurations.",
}

var nginxListCmd = &cobra.Command{
	Use:   "list",
	Short: "List virtual hosts",
	RunE:  runNginxList,
}

var nginxTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test Nginx configuration",
	RunE:  runNginxTest,
}

var nginxReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload Nginx configuration",
	RunE:  runNginxReload,
}

func init() {
	nginxCmd.AddCommand(nginxListCmd)
	nginxCmd.AddCommand(nginxTestCmd)
	nginxCmd.AddCommand(nginxReloadCmd)
	rootCmd.AddCommand(nginxCmd)
}

func getNginxManager() *nginx.Manager {
	cfg := config.DefaultConfig()
	return nginx.NewManager(
		cfg.Paths.Configs,
		filepath.Join(cfg.Paths.Configs, "nginx", "sites"),
		cfg.Paths.Logs,
		cfg.Paths.TawsRoot,
	)
}

func runNginxList(cmd *cobra.Command, args []string) error {
	m := getNginxManager()

	vhosts, err := m.ListVhosts()
	if err != nil {
		return err
	}

	fmt.Print(nginx.FormatVhosts(vhosts))
	return nil
}

func runNginxTest(cmd *cobra.Command, args []string) error {
	m := getNginxManager()

	if err := m.TestConfig(); err != nil {
		return err
	}

	fmt.Println("Nginx configuration test passed.")
	return nil
}

func runNginxReload(cmd *cobra.Command, args []string) error {
	m := getNginxManager()

	if err := m.Reload(); err != nil {
		return err
	}

	fmt.Println("Nginx reloaded successfully.")
	return nil
}
