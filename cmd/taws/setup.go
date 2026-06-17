package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"taws/internal/installer"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initial setup and installation of TAWS",
	Long:  "Verifies the environment, checks dependencies, creates directories, and generates initial configuration.",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	result, err := installer.RunSetup()
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	fmt.Print(installer.FormatSetupResult(result))

	if !result.Success {
		return fmt.Errorf("setup completed with errors")
	}

	return nil
}
