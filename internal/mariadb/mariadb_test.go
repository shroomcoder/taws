package mariadb

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestManager(t *testing.T) *Manager {
	t.Helper()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")
	dataDir := filepath.Join(tmpDir, "data")
	logDir := filepath.Join(tmpDir, "logs")
	socketDir := filepath.Join(tmpDir, "run")
	templatesDir := filepath.Join(tmpDir, "templates")

	os.MkdirAll(configDir, 0755)
	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(logDir, 0755)
	os.MkdirAll(socketDir, 0755)
	os.MkdirAll(filepath.Join(templatesDir, "mariadb"), 0755)

	tmpl := `[mysqld]
user = {{.User}}
datadir = {{.DataDir}}
port = {{.Port}}
`
	os.WriteFile(filepath.Join(templatesDir, "mariadb", "mariadb.conf.tmpl"), []byte(tmpl), 0644)

	return NewManager(configDir, dataDir, logDir, socketDir, templatesDir)
}

func TestGenerateConfig(t *testing.T) {
	m := setupTestManager(t)

	cfg := &Config{
		User:           "root",
		DataDir:        "/var/lib/mysql",
		Port:           3306,
		BindAddress:    "127.0.0.1",
		BufferPoolSize: "128M",
	}

	if err := m.GenerateConfig(cfg); err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}

	configPath := filepath.Join(m.configDir, "mariadb", "mariadb.conf")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("mariadb config file should exist")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig(3306)

	if cfg.Port != 3306 {
		t.Errorf("expected port 3306, got %d", cfg.Port)
	}

	if cfg.User != "root" {
		t.Errorf("expected user root, got %s", cfg.User)
	}

	if cfg.BindAddress != "127.0.0.1" {
		t.Errorf("expected bind address 127.0.0.1, got %s", cfg.BindAddress)
	}
}

func TestGetVersion(t *testing.T) {
	m := setupTestManager(t)

	version, err := m.GetVersion()
	if err != nil {
		t.Logf("GetVersion failed (MariaDB may not be installed): %v", err)
		return
	}

	if version == "" {
		t.Error("version should not be empty")
	}
}

func TestListDatabases(t *testing.T) {
	m := setupTestManager(t)

	databases, err := m.ListDatabases()
	if err != nil {
		t.Logf("ListDatabases failed (MariaDB may not be installed): %v", err)
		return
	}

	if databases == nil {
		t.Error("databases should not be nil")
	}
}

func TestFormatDatabases(t *testing.T) {
	databases := []string{"mysql", "information_schema", "test_db"}

	output := FormatDatabases(databases)

	if output == "" {
		t.Fatal("FormatDatabases returned empty string")
	}

	if len(output) < 30 {
		t.Error("FormatDatabases output too short")
	}
}

func TestFormatDatabasesEmpty(t *testing.T) {
	output := FormatDatabases([]string{})

	if output == "" {
		t.Fatal("FormatDatabases returned empty string")
	}
}

func TestFormatTables(t *testing.T) {
	tables := []string{"users", "posts", "comments"}

	output := FormatTables("test_db", tables)

	if output == "" {
		t.Fatal("FormatTables returned empty string")
	}

	if len(output) < 30 {
		t.Error("FormatTables output too short")
	}
}

func TestFormatTablesEmpty(t *testing.T) {
	output := FormatTables("test_db", []string{})

	if output == "" {
		t.Fatal("FormatTables returned empty string")
	}
}
