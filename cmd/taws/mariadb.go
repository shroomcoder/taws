package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"taws/internal/config"
	"taws/internal/mariadb"
	"taws/internal/validate"
)

var mariadbCmd = &cobra.Command{
	Use:   "mariadb",
	Short: "Manage MariaDB databases",
	Long:  "Create, delete, list databases and manage MariaDB.",
}

var mariadbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all databases",
	RunE:  runMariadbList,
}

var mariadbCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new database",
	Args:  cobra.ExactArgs(1),
	RunE:  runMariadbCreate,
}

var mariadbDropCmd = &cobra.Command{
	Use:   "drop <name>",
	Short: "Drop a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runMariadbDrop,
}

var mariadbTablesCmd = &cobra.Command{
	Use:   "tables <database>",
	Short: "List tables in a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runMariadbTables,
}

var mariadbGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate MariaDB configuration",
	RunE:  runMariadbGenerate,
}

func init() {
	mariadbCmd.AddCommand(mariadbListCmd)
	mariadbCmd.AddCommand(mariadbCreateCmd)
	mariadbCmd.AddCommand(mariadbDropCmd)
	mariadbCmd.AddCommand(mariadbTablesCmd)
	mariadbCmd.AddCommand(mariadbGenerateCmd)
	rootCmd.AddCommand(mariadbCmd)
}

func getMariadbManager() *mariadb.Manager {
	cfg := config.DefaultConfig()
	return mariadb.NewManager(
		cfg.Paths.Configs,
		cfg.Paths.State,
		cfg.Paths.Logs,
		cfg.Paths.TawsRoot,
		cfg.Paths.TawsRoot,
	)
}

func runMariadbList(cmd *cobra.Command, args []string) error {
	m := getMariadbManager()

	databases, err := m.ListDatabases()
	if err != nil {
		return err
	}

	fmt.Print(mariadb.FormatDatabases(databases))
	return nil
}

func runMariadbCreate(cmd *cobra.Command, args []string) error {
	if !validate.DatabaseName(args[0]) {
		return fmt.Errorf("invalid database name: must match [a-zA-Z_][a-zA-Z0-9_]*")
	}

	m := getMariadbManager()

	if err := m.CreateDatabase(args[0]); err != nil {
		return err
	}

	fmt.Printf("Database %s created successfully.\n", args[0])
	return nil
}

func runMariadbDrop(cmd *cobra.Command, args []string) error {
	if !validate.DatabaseName(args[0]) {
		return fmt.Errorf("invalid database name: must match [a-zA-Z_][a-zA-Z0-9_]*")
	}

	m := getMariadbManager()

	if err := m.DropDatabase(args[0]); err != nil {
		return err
	}

	fmt.Printf("Database %s dropped successfully.\n", args[0])
	return nil
}

func runMariadbTables(cmd *cobra.Command, args []string) error {
	if !validate.DatabaseName(args[0]) {
		return fmt.Errorf("invalid database name: must match [a-zA-Z_][a-zA-Z0-9_]*")
	}

	m := getMariadbManager()

	tables, err := m.ListTables(args[0])
	if err != nil {
		return err
	}

	fmt.Print(mariadb.FormatTables(args[0], tables))
	return nil
}

func runMariadbGenerate(cmd *cobra.Command, args []string) error {
	m := getMariadbManager()

	cfg := config.DefaultConfig()
	mariadbCfg := mariadb.DefaultConfig(cfg.Services.MariaDB.Port)

	if err := m.GenerateConfig(mariadbCfg); err != nil {
		return fmt.Errorf("generating mariadb config: %w", err)
	}

	fmt.Println("MariaDB configuration generated successfully.")
	fmt.Printf("  Config: %s/mariadb/mariadb.conf\n", cfg.Paths.Configs)

	return nil
}
