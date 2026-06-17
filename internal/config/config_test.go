package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Paths.TawsRoot == "" {
		t.Error("TawsRoot should not be empty")
	}

	if cfg.Paths.WebRoot == "" {
		t.Error("WebRoot should not be empty")
	}

	if cfg.Services.Nginx.Port != 8080 {
		t.Errorf("expected nginx port 8080, got %d", cfg.Services.Nginx.Port)
	}

	if cfg.Services.MariaDB.Port != 3306 {
		t.Errorf("expected mariadb port 3306, got %d", cfg.Services.MariaDB.Port)
	}
}

func TestSaveAndLoadYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Services.Nginx.Port = 9090

	if err := Save(cfg, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Services.Nginx.Port != 9090 {
		t.Errorf("expected nginx port 9090, got %d", loaded.Services.Nginx.Port)
	}
}

func TestSaveAndLoadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := []byte(`{"paths":{"home":"/home/test"},"services":{"nginx":{"enabled":true,"port":8888}}}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if cfg.Services.Nginx.Port != 8888 {
		t.Errorf("expected nginx port 8888, got %d", cfg.Services.Nginx.Port)
	}
}

func TestLoadUnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.txt")

	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestEnsureDirs(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Paths.TawsRoot = filepath.Join(dir, ".taws")
	cfg.Paths.Configs = filepath.Join(dir, ".taws", "configs")
	cfg.Paths.State = filepath.Join(dir, ".taws", "state")
	cfg.Paths.Logs = filepath.Join(dir, ".taws", "logs")
	cfg.Paths.Backups = filepath.Join(dir, ".taws", "backups")
	cfg.Paths.Certs = filepath.Join(dir, ".taws", "certs")
	cfg.Paths.Plugins = filepath.Join(dir, ".taws", "plugins")
	cfg.Paths.WebRoot = filepath.Join(dir, "web")

	if err := EnsureDirs(cfg); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	for _, p := range []string{
		cfg.Paths.TawsRoot,
		cfg.Paths.Configs,
		cfg.Paths.State,
		cfg.Paths.Logs,
		cfg.Paths.WebRoot,
	} {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("directory %s should exist", p)
		}
	}
}
