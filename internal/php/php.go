package php

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type FPMConfig struct {
	Name             string
	User             string
	Group            string
	Port             int
	MaxChildren      int
	StartServers     int
	MinSpareServers  int
	MaxSpareServers  int
	MaxRequests      int
	LogDir           string
	UploadMaxFilesize string
	PostMaxSize      string
	MemoryLimit      string
	MaxExecutionTime int
	MaxInputTime     int
}

type PHPConfig struct {
	ErrorReporting      string
	DisplayErrors       string
	DisplayStartupErrors string
	LogDir              string
	MemoryLimit         string
	MaxExecutionTime    int
	MaxInputTime        int
	UploadMaxFilesize   string
	PostMaxSize         string
	SessionPath         string
	Timezone            string
}

type Manager struct {
	configDir    string
	logDir       string
	templatesDir string
}

func NewManager(configDir, logDir, templatesDir string) *Manager {
	return &Manager{
		configDir:    configDir,
		logDir:       logDir,
		templatesDir: templatesDir,
	}
}

func (m *Manager) GenerateFPMConfig(cfg *FPMConfig) error {
	tmplPath := filepath.Join(m.templatesDir, "php", "php-fpm.conf.tmpl")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("parsing fpm template: %w", err)
	}

	outputDir := filepath.Join(m.configDir, "php-fpm", "pools")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, cfg.Name+".conf")

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

func (m *Manager) GeneratePHPIni(cfg *PHPConfig) error {
	tmplPath := filepath.Join(m.templatesDir, "php", "php.ini.tmpl")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("parsing php.ini template: %w", err)
	}

	outputDir := filepath.Join(m.configDir, "php")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, "php.ini")

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
	cmd := exec.Command("php", "-v")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("php version check failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "unknown", nil
}

func (m *Manager) GetModules() ([]string, error) {
	cmd := exec.Command("php", "-m")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("php modules check failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var modules []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "[") {
			modules = append(modules, line)
		}
	}

	return modules, nil
}

func (m *Manager) CheckModule(module string) (bool, error) {
	cmd := exec.Command("php", "-m")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("php modules check failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == module {
			return true, nil
		}
	}

	return false, nil
}

func (m *Manager) TestConfig() error {
	cmd := exec.Command("php", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("php config test failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func DefaultFPMConfig(name string, port int) *FPMConfig {
	home, _ := os.UserHomeDir()

	return &FPMConfig{
		Name:             name,
		User:             "www-data",
		Group:            "www-data",
		Port:             port,
		MaxChildren:      5,
		StartServers:     2,
		MinSpareServers:  1,
		MaxSpareServers:  3,
		MaxRequests:      500,
		LogDir:           filepath.Join(home, ".taws", "logs"),
		UploadMaxFilesize: "64M",
		PostMaxSize:      "64M",
		MemoryLimit:      "256M",
		MaxExecutionTime: 30,
		MaxInputTime:     60,
	}
}

func DefaultPHPConfig() *PHPConfig {
	home, _ := os.UserHomeDir()

	return &PHPConfig{
		ErrorReporting:      "E_ALL",
		DisplayErrors:       "Off",
		DisplayStartupErrors: "Off",
		LogDir:              filepath.Join(home, ".taws", "logs"),
		MemoryLimit:         "256M",
		MaxExecutionTime:    30,
		MaxInputTime:        60,
		UploadMaxFilesize:   "64M",
		PostMaxSize:         "64M",
		SessionPath:         filepath.Join(home, ".taws", "cache", "sessions"),
		Timezone:            "UTC",
	}
}

func FormatPHPInfo(version string, modules []string) string {
	var sb strings.Builder

	sb.WriteString("TAWS PHP Information\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	sb.WriteString(fmt.Sprintf("  Version: %s\n\n", version))

	sb.WriteString("  Modules:\n")
	for _, m := range modules {
		sb.WriteString(fmt.Sprintf("    - %s\n", m))
	}

	sb.WriteString("\n" + strings.Repeat("=", 50) + "\n")

	return sb.String()
}
