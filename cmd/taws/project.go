package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"taws/internal/config"
	"taws/internal/nginx"
	"taws/internal/projects"
	"taws/internal/state"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage TAWS projects",
	Long:  "Create, delete, and list web projects.",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project",
	Long:  "Creates a new web project directory with a default index.php file.",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectCreate,
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a project",
	Long:  "Removes a project directory and its metadata.",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectDelete,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  "Shows all registered TAWS projects.",
	RunE:  runProjectList,
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)
	rootCmd.AddCommand(projectCmd)
}

func getProjectManager() (*projects.Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	cfg := config.DefaultConfig()
	stateMgr := state.NewManager(cfg.Paths.State)

	return projects.NewManager(filepath.Join(home, "web"), stateMgr), nil
}

func runProjectCreate(cmd *cobra.Command, args []string) error {
	m, err := getProjectManager()
	if err != nil {
		return err
	}

	p, err := m.Create(args[0])
	if err != nil {
		return err
	}

	cfg := config.DefaultConfig()
	nginxMgr := nginx.NewManager(
		cfg.Paths.Configs,
		filepath.Join(cfg.Paths.Configs, "nginx", "sites"),
		cfg.Paths.Logs,
		cfg.Paths.TawsRoot,
	)

	vhostCfg := &nginx.Config{
		Name:       p.Name,
		Domain:     p.Domain,
		Root:       p.Path,
		Port:       cfg.Services.Nginx.Port,
		PHPFPMPort: cfg.Services.PHP.Port,
		LogDir:     cfg.Paths.Logs,
	}

	if err := nginxMgr.GenerateVhost(vhostCfg); err != nil {
		fmt.Printf("Warning: Could not generate Nginx config: %v\n", err)
	} else {
		if err := nginxMgr.EnableSite(p.Name); err != nil {
			fmt.Printf("Warning: Could not enable site: %v\n", err)
		}
	}

	fmt.Printf("Project created successfully.\n\n")
	fmt.Printf("  Name:   %s\n", p.Name)
	fmt.Printf("  Path:   %s\n", p.Path)
	fmt.Printf("  Domain: %s\n\n", p.Domain)
	fmt.Println("Next steps:")
	fmt.Printf("  1. Add your PHP files to %s\n", p.Path)
	fmt.Printf("  2. Access via http://%s\n", p.Domain)

	return nil
}

func runProjectDelete(cmd *cobra.Command, args []string) error {
	m, err := getProjectManager()
	if err != nil {
		return err
	}

	if err := m.Delete(args[0]); err != nil {
		return err
	}

	fmt.Printf("Project %s deleted.\n", args[0])
	return nil
}

func runProjectList(cmd *cobra.Command, args []string) error {
	m, err := getProjectManager()
	if err != nil {
		return err
	}

	projs, err := m.List()
	if err != nil {
		return err
	}

	fmt.Print(projects.FormatProjects(projs))
	return nil
}
