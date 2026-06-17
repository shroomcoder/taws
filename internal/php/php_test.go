package php

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestManager(t *testing.T) *Manager {
	t.Helper()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")
	logDir := filepath.Join(tmpDir, "logs")
	templatesDir := filepath.Join(tmpDir, "templates")

	os.MkdirAll(configDir, 0755)
	os.MkdirAll(logDir, 0755)
	os.MkdirAll(filepath.Join(templatesDir, "php"), 0755)

	fpmTmpl := `[{{.Name}}]
listen = 127.0.0.1:{{.Port}}
pm = dynamic
pm.max_children = {{.MaxChildren}}
`
	os.WriteFile(filepath.Join(templatesDir, "php", "php-fpm.conf.tmpl"), []byte(fpmTmpl), 0644)

	iniTmpl := `error_reporting = {{.ErrorReporting}}
display_errors = {{.DisplayErrors}}
memory_limit = {{.MemoryLimit}}
`
	os.WriteFile(filepath.Join(templatesDir, "php", "php.ini.tmpl"), []byte(iniTmpl), 0644)

	return NewManager(configDir, logDir, templatesDir)
}

func TestGenerateFPMConfig(t *testing.T) {
	m := setupTestManager(t)

	cfg := &FPMConfig{
		Name:        "taws",
		User:        "www-data",
		Group:       "www-data",
		Port:        9000,
		MaxChildren: 5,
	}

	if err := m.GenerateFPMConfig(cfg); err != nil {
		t.Fatalf("GenerateFPMConfig failed: %v", err)
	}

	configPath := filepath.Join(m.configDir, "php-fpm", "pools", "taws.conf")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("FPM config file should exist")
	}
}

func TestGeneratePHPIni(t *testing.T) {
	m := setupTestManager(t)

	cfg := &PHPConfig{
		ErrorReporting: "E_ALL",
		DisplayErrors:  "Off",
		MemoryLimit:    "256M",
	}

	if err := m.GeneratePHPIni(cfg); err != nil {
		t.Fatalf("GeneratePHPIni failed: %v", err)
	}

	configPath := filepath.Join(m.configDir, "php", "php.ini")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("php.ini file should exist")
	}
}

func TestGetVersion(t *testing.T) {
	m := setupTestManager(t)

	version, err := m.GetVersion()
	if err != nil {
		t.Logf("GetVersion failed (PHP may not be installed): %v", err)
		return
	}

	if version == "" {
		t.Error("version should not be empty")
	}
}

func TestGetModules(t *testing.T) {
	m := setupTestManager(t)

	modules, err := m.GetModules()
	if err != nil {
		t.Logf("GetModules failed (PHP may not be installed): %v", err)
		return
	}

	if modules == nil {
		t.Error("modules should not be nil")
	}
}

func TestCheckModule(t *testing.T) {
	m := setupTestManager(t)

	exists, err := m.CheckModule("Core")
	if err != nil {
		t.Logf("CheckModule failed (PHP may not be installed): %v", err)
		return
	}

	if !exists {
		t.Error("Core module should exist")
	}
}

func TestDefaultFPMConfig(t *testing.T) {
	cfg := DefaultFPMConfig("taws", 9000)

	if cfg.Name != "taws" {
		t.Errorf("expected name taws, got %s", cfg.Name)
	}

	if cfg.Port != 9000 {
		t.Errorf("expected port 9000, got %d", cfg.Port)
	}

	if cfg.MaxChildren != 5 {
		t.Errorf("expected max children 5, got %d", cfg.MaxChildren)
	}
}

func TestDefaultPHPConfig(t *testing.T) {
	cfg := DefaultPHPConfig()

	if cfg.ErrorReporting != "E_ALL" {
		t.Errorf("expected error reporting E_ALL, got %s", cfg.ErrorReporting)
	}

	if cfg.MemoryLimit != "256M" {
		t.Errorf("expected memory limit 256M, got %s", cfg.MemoryLimit)
	}

	if cfg.Timezone != "UTC" {
		t.Errorf("expected timezone UTC, got %s", cfg.Timezone)
	}
}

func TestFormatPHPInfo(t *testing.T) {
	output := FormatPHPInfo("PHP 8.3.0", []string{"Core", "date", "json"})

	if output == "" {
		t.Fatal("FormatPHPInfo returned empty string")
	}

	if len(output) < 30 {
		t.Error("FormatPHPInfo output too short")
	}
}
