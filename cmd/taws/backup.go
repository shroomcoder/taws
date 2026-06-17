package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"taws/internal/backup"
	"taws/internal/config"
	"taws/internal/validate"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup snapshot",
	Long:  "Creates a snapshot of TAWS configuration and state files.",
	RunE:  runBackup,
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback [snapshot-id]",
	Short: "Restore from a backup snapshot",
	Long:  "Restores TAWS configuration and state from a previous snapshot.",
	Args:  cobra.ExactArgs(1),
	RunE:  runRollback,
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all backup snapshots",
	RunE:  runBackupList,
}

var backupDeleteCmd = &cobra.Command{
	Use:   "delete <snapshot-id>",
	Short: "Delete a backup snapshot",
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupDelete,
}

func init() {
	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupDeleteCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(rollbackCmd)
}

func getBackupManager() *backup.Manager {
	cfg := config.DefaultConfig()
	return backup.NewManager(
		cfg.Paths.Backups,
		cfg.Paths.Configs,
		cfg.Paths.State,
	)
}

func runBackup(cmd *cobra.Command, args []string) error {
	m := getBackupManager()

	fmt.Println("Creating backup snapshot...")

	snapshot, err := m.CreateSnapshot("1.0.0")
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Printf("Backup created successfully.\n\n")
	fmt.Printf("  ID:        %s\n", snapshot.ID)
	fmt.Printf("  Files:     %d\n", len(snapshot.Files))
	fmt.Printf("  Size:      %d bytes\n", snapshot.Size)

	return nil
}

func runRollback(cmd *cobra.Command, args []string) error {
	if !validate.SnapshotID(args[0]) {
		return fmt.Errorf("invalid snapshot ID: expected format YYYYMMDD-HHMMSS")
	}

	m := getBackupManager()

	fmt.Printf("Restoring from snapshot %s...\n", args[0])

	if err := m.RestoreSnapshot(args[0]); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("Rollback completed successfully.")
	return nil
}

func runBackupList(cmd *cobra.Command, args []string) error {
	m := getBackupManager()

	snapshots, err := m.ListSnapshots()
	if err != nil {
		return err
	}

	fmt.Print(backup.FormatSnapshots(snapshots))
	return nil
}

func runBackupDelete(cmd *cobra.Command, args []string) error {
	if !validate.SnapshotID(args[0]) {
		return fmt.Errorf("invalid snapshot ID: expected format YYYYMMDD-HHMMSS")
	}

	m := getBackupManager()

	if err := m.DeleteSnapshot(args[0]); err != nil {
		return err
	}

	fmt.Printf("Snapshot %s deleted.\n", args[0])
	return nil
}
