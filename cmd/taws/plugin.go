package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"taws/internal/config"
	"taws/internal/plugin"
	"taws/internal/validate"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage TAWS plugins",
	Long:  "Install, remove, enable, disable, and list plugins.",
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginInstall,
}

var pluginRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginRemove,
}

var pluginEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginEnable,
}

var pluginDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginDisable,
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all plugins",
	RunE:  runPluginList,
}

var pluginAvailableCmd = &cobra.Command{
	Use:   "available",
	Short: "List available plugins",
	RunE:  runPluginAvailable,
}

func init() {
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginRemoveCmd)
	pluginCmd.AddCommand(pluginEnableCmd)
	pluginCmd.AddCommand(pluginDisableCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginAvailableCmd)
	rootCmd.AddCommand(pluginCmd)
}

func getPluginManager() *plugin.Manager {
	cfg := config.DefaultConfig()
	return plugin.NewManager(cfg.Paths.Plugins)
}

func runPluginInstall(cmd *cobra.Command, args []string) error {
	if !validate.PluginName(args[0]) {
		return fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	m := getPluginManager()

	if err := m.Install(args[0]); err != nil {
		return err
	}

	fmt.Printf("Plugin %s installed successfully.\n", args[0])
	return nil
}

func runPluginRemove(cmd *cobra.Command, args []string) error {
	if !validate.PluginName(args[0]) {
		return fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	m := getPluginManager()

	if err := m.Remove(args[0]); err != nil {
		return err
	}

	fmt.Printf("Plugin %s removed.\n", args[0])
	return nil
}

func runPluginEnable(cmd *cobra.Command, args []string) error {
	if !validate.PluginName(args[0]) {
		return fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	m := getPluginManager()

	if err := m.Enable(args[0]); err != nil {
		return err
	}

	fmt.Printf("Plugin %s enabled.\n", args[0])
	return nil
}

func runPluginDisable(cmd *cobra.Command, args []string) error {
	if !validate.PluginName(args[0]) {
		return fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	m := getPluginManager()

	if err := m.Disable(args[0]); err != nil {
		return err
	}

	fmt.Printf("Plugin %s disabled.\n", args[0])
	return nil
}

func runPluginList(cmd *cobra.Command, args []string) error {
	m := getPluginManager()

	plugins, err := m.List()
	if err != nil {
		return err
	}

	fmt.Print(plugin.FormatPlugins(plugins))
	return nil
}

func runPluginAvailable(cmd *cobra.Command, args []string) error {
	fmt.Println("Available plugins:")
	for _, p := range plugin.AvailablePlugins {
		fmt.Printf("  - %s\n", p)
	}
	return nil
}
