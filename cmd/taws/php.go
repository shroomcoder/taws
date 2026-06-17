package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"taws/internal/config"
	"taws/internal/php"
)

var phpCmd = &cobra.Command{
	Use:   "php",
	Short: "Manage PHP configuration",
	Long:  "View PHP information and manage PHP-FPM configuration.",
}

var phpInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show PHP information",
	RunE:  runPHPInfo,
}

var phpModulesCmd = &cobra.Command{
	Use:   "modules",
	Short: "List PHP modules",
	RunE:  runPHPModules,
}

var phpGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate PHP configuration files",
	RunE:  runPHPGenerate,
}

func init() {
	phpCmd.AddCommand(phpInfoCmd)
	phpCmd.AddCommand(phpModulesCmd)
	phpCmd.AddCommand(phpGenerateCmd)
	rootCmd.AddCommand(phpCmd)
}

func getPHPManager() *php.Manager {
	cfg := config.DefaultConfig()
	return php.NewManager(
		cfg.Paths.Configs,
		cfg.Paths.Logs,
		cfg.Paths.TawsRoot,
	)
}

func runPHPInfo(cmd *cobra.Command, args []string) error {
	m := getPHPManager()

	version, err := m.GetVersion()
	if err != nil {
		return err
	}

	modules, err := m.GetModules()
	if err != nil {
		return err
	}

	fmt.Print(php.FormatPHPInfo(version, modules))
	return nil
}

func runPHPModules(cmd *cobra.Command, args []string) error {
	m := getPHPManager()

	modules, err := m.GetModules()
	if err != nil {
		return err
	}

	for _, mod := range modules {
		fmt.Println(mod)
	}
	return nil
}

func runPHPGenerate(cmd *cobra.Command, args []string) error {
	m := getPHPManager()

	cfg := config.DefaultConfig()

	fpmCfg := php.DefaultFPMConfig("taws", cfg.Services.PHP.Port)
	if err := m.GenerateFPMConfig(fpmCfg); err != nil {
		return fmt.Errorf("generating FPM config: %w", err)
	}

	phpCfg := php.DefaultPHPConfig()
	if err := m.GeneratePHPIni(phpCfg); err != nil {
		return fmt.Errorf("generating php.ini: %w", err)
	}

	fmt.Println("PHP configuration files generated successfully.")
	fmt.Printf("  FPM config: %s/php-fpm/pools/taws.conf\n", cfg.Paths.Configs)
	fmt.Printf("  php.ini:    %s/php/php.ini\n", cfg.Paths.Configs)

	return nil
}
