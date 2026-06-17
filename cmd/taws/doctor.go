package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"taws/internal/doctor"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system dependencies and configuration",
	Long:  "Validates the development environment and runtime dependencies for TAWS.",
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	d := doctor.New()

	cfg, _, _ := loadConfigPath()
	if cfg != nil {
		d.SetConfig(cfg)
	}

	checks := d.Run()

	fmt.Print(doctor.FormatResults(checks))

	for _, c := range checks {
		if c.Status == "FAIL" || c.Status == "MISSING" {
			if c.Critical {
				return fmt.Errorf("critical dependency missing: %s", c.Name)
			}
		}
	}

	return nil
}
