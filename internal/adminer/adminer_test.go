package adminer

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestManager(t *testing.T) *Manager {
	t.Helper()

	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	logDir := filepath.Join(tmpDir, "logs")

	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(logDir, 0755)

	return NewManager(dataDir, logDir, 8081)
}

func TestNewManager(t *testing.T) {
	m := setupTestManager(t)

	if m == nil {
		t.Fatal("NewManager returned nil")
	}

	if m.port != 8081 {
		t.Errorf("expected port 8081, got %d", m.port)
	}
}

func TestIsInstalled(t *testing.T) {
	m := setupTestManager(t)

	if m.IsInstalled() {
		t.Error("Adminer should not be installed initially")
	}
}

func TestGetVersionNotInstalled(t *testing.T) {
	m := setupTestManager(t)

	version := m.GetVersion()
	if version != "not installed" {
		t.Errorf("expected 'not installed', got %s", version)
	}
}

func TestGetStatus(t *testing.T) {
	m := setupTestManager(t)

	status := m.GetStatus()

	if status == nil {
		t.Fatal("GetStatus returned nil")
	}

	if status.Installed {
		t.Error("Adminer should not be installed")
	}

	if status.Running {
		t.Error("Adminer should not be running")
	}

	if status.Port != 8081 {
		t.Errorf("expected port 8081, got %d", status.Port)
	}
}

func TestGetPublicDir(t *testing.T) {
	m := setupTestManager(t)

	dir := m.GetPublicDir()

	if dir == "" {
		t.Error("GetPublicDir returned empty string")
	}

	if !filepath.IsAbs(dir) {
		t.Error("GetPublicDir should return absolute path")
	}
}

func TestGetConfig(t *testing.T) {
	m := setupTestManager(t)

	cfg := m.GetConfig()

	if cfg["server"] != "127.0.0.1" {
		t.Errorf("expected server 127.0.0.1, got %s", cfg["server"])
	}

	if cfg["username"] != "" {
		t.Errorf("expected empty username, got %s", cfg["username"])
	}

	if cfg["password"] != "" {
		t.Errorf("expected empty password, got %s", cfg["password"])
	}
}

func TestFormatStatus(t *testing.T) {
	status := &Status{
		Installed: true,
		Running:   true,
		Port:      8081,
		URL:       "http://127.0.0.1:8081",
		Version:   "Adminer 4.8.1",
	}

	output := FormatStatus(status)

	if output == "" {
		t.Fatal("FormatStatus returned empty string")
	}

	if len(output) < 30 {
		t.Error("FormatStatus output too short")
	}
}

func TestFormatStatusNotInstalled(t *testing.T) {
	status := &Status{
		Installed: false,
		Running:   false,
		Port:      8081,
		URL:       "http://127.0.0.1:8081",
		Version:   "not installed",
	}

	output := FormatStatus(status)

	if output == "" {
		t.Fatal("FormatStatus returned empty string")
	}
}
