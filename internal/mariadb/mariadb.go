package mariadb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"taws/internal/validate"
)

type Config struct {
	User              string
	DataDir           string
	SocketDir         string
	PIDFile           string
	BindAddress       string
	Port              int
	LogDir            string
	BufferPoolSize    string
	LogFileSize       string
	MaxConnections    int
	MaxAllowedPacket  string
}

type Manager struct {
	configDir    string
	dataDir      string
	logDir       string
	socketDir    string
	templatesDir string
}

func NewManager(configDir, dataDir, logDir, socketDir, templatesDir string) *Manager {
	return &Manager{
		configDir:    configDir,
		dataDir:      dataDir,
		logDir:       logDir,
		socketDir:    socketDir,
		templatesDir: templatesDir,
	}
}

func (m *Manager) GenerateConfig(cfg *Config) error {
	tmplPath := filepath.Join(m.templatesDir, "mariadb", "mariadb.conf.tmpl")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("parsing mariadb template: %w", err)
	}

	outputDir := filepath.Join(m.configDir, "mariadb")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, "mariadb.conf")

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

func (m *Manager) GetVersion() (string, error) {
	cmds := [][]string{
		{"mariadb", "--version"},
		{"mysql", "--version"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	}

	return "", fmt.Errorf("mariadb/mysql not found")
}

func (m *Manager) Execute(query string) (string, error) {
	cmds := [][]string{
		{"mariadb", "-e", query},
		{"mysql", "-e", query},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return string(output), nil
		}
	}

	return "", fmt.Errorf("failed to execute query")
}

func (m *Manager) ExecuteWithDB(db, query string) (string, error) {
	cmds := [][]string{
		{"mariadb", db, "-e", query},
		{"mysql", db, "-e", query},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return string(output), nil
		}
	}

	return "", fmt.Errorf("failed to execute query on database %s", db)
}

func (m *Manager) ListDatabases() ([]string, error) {
	output, err := m.Execute("SHOW DATABASES")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var databases []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "Database" {
			databases = append(databases, line)
		}
	}

	return databases, nil
}

func (m *Manager) CreateDatabase(name string) error {
	if !validate.DatabaseName(name) {
		return fmt.Errorf("invalid database name: must match [a-zA-Z_][a-zA-Z0-9_]*")
	}
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", name)
	_, err := m.Execute(query)
	if err != nil {
		return fmt.Errorf("creating database %s: %w", name, err)
	}
	return nil
}

func (m *Manager) DropDatabase(name string) error {
	if !validate.DatabaseName(name) {
		return fmt.Errorf("invalid database name: must match [a-zA-Z_][a-zA-Z0-9_]*")
	}
	query := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", name)
	_, err := m.Execute(query)
	if err != nil {
		return fmt.Errorf("dropping database %s: %w", name, err)
	}
	return nil
}

func (m *Manager) DatabaseExists(name string) (bool, error) {
	databases, err := m.ListDatabases()
	if err != nil {
		return false, err
	}

	for _, db := range databases {
		if db == name {
			return true, nil
		}
	}

	return false, nil
}

func (m *Manager) ListTables(db string) ([]string, error) {
	output, err := m.ExecuteWithDB(db, "SHOW TABLES")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var tables []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "Tables_in_") {
			tables = append(tables, line)
		}
	}

	return tables, nil
}

func (m *Manager) ImportSQL(db, filePath string) error {
	var cmd *exec.Cmd

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("SQL file not found: %s", filePath)
	}

	cmds := [][]string{
		{"mariadb", db},
		{"mysql", db},
	}

	for _, args := range cmds {
		cmd = exec.Command(args[0], args[1:]...)
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("opening SQL file: %w", err)
		}
		cmd.Stdin = f
		output, err := cmd.CombinedOutput()
		f.Close()
		if err == nil {
			return nil
		}
		_ = output
	}

	return fmt.Errorf("failed to import SQL file")
}

func (m *Manager) ExportDatabase(db, outputPath string) error {
	cmds := [][]string{
		{"mariadb-dump", db},
		{"mysqldump", db},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.Output()
		if err == nil {
			return os.WriteFile(outputPath, output, 0600)
		}
	}

	return fmt.Errorf("failed to export database")
}

func DefaultConfig(port int) *Config {
	home, _ := os.UserHomeDir()
	tawsRoot := filepath.Join(home, ".taws")

	return &Config{
		User:             "root",
		DataDir:          filepath.Join(tawsRoot, "data", "mariadb"),
		SocketDir:        filepath.Join(tawsRoot, "run"),
		PIDFile:          filepath.Join(tawsRoot, "run", "mariadb.pid"),
		BindAddress:      "127.0.0.1",
		Port:             port,
		LogDir:           filepath.Join(tawsRoot, "logs"),
		BufferPoolSize:   "128M",
		LogFileSize:      "48M",
		MaxConnections:   100,
		MaxAllowedPacket: "64M",
	}
}

func FormatDatabases(databases []string) string {
	var sb strings.Builder

	sb.WriteString("TAWS MariaDB Databases\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(databases) == 0 {
		sb.WriteString("  No databases found.\n")
	} else {
		for _, db := range databases {
			sb.WriteString(fmt.Sprintf("  - %s\n", db))
		}
	}

	sb.WriteString("\n" + strings.Repeat("=", 50) + "\n")

	return sb.String()
}

func FormatTables(db string, tables []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("TAWS MariaDB Tables: %s\n", db))
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(tables) == 0 {
		sb.WriteString("  No tables found.\n")
	} else {
		for _, t := range tables {
			sb.WriteString(fmt.Sprintf("  - %s\n", t))
		}
	}

	sb.WriteString("\n" + strings.Repeat("=", 50) + "\n")

	return sb.String()
}
