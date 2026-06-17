package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectEnvironment(t *testing.T) {
	env, err := DetectEnvironment()
	if err != nil {
		t.Fatalf("DetectEnvironment failed: %v", err)
	}

	if env.HomeDir == "" {
		t.Error("HomeDir should not be empty")
	}

	if env.TawsRoot == "" {
		t.Error("TawsRoot should not be empty")
	}

	if env.WebRoot == "" {
		t.Error("WebRoot should not be empty")
	}

	if runtime.GOOS == "android" {
		if !env.IsAndroid {
			t.Error("IsAndroid should be true on Android")
		}
	} else {
		if env.IsAndroid {
			t.Error("IsAndroid should be false on non-Android")
		}
	}
}

func TestCheckPackages(t *testing.T) {
	env := &Environment{
		IsTermux:  false,
		IsAndroid: false,
		HomeDir:   t.TempDir(),
	}

	packages := CheckPackages(env)

	if len(packages) == 0 {
		t.Fatal("CheckPackages returned no packages")
	}

	foundNginx := false
	foundPHP := false
	for _, pkg := range packages {
		if pkg.Name == "nginx" {
			foundNginx = true
		}
		if pkg.Name == "php" {
			foundPHP = true
		}
	}

	if !foundNginx {
		t.Error("nginx package check not found")
	}

	if !foundPHP {
		t.Error("php package check not found")
	}
}

func TestCreateDirectories(t *testing.T) {
	env := &Environment{
		HomeDir:  t.TempDir(),
		TawsRoot: filepath.Join(t.TempDir(), ".taws"),
		WebRoot:  filepath.Join(t.TempDir(), "web"),
	}

	err := CreateDirectories(env)
	if err != nil {
		t.Fatalf("CreateDirectories failed: %v", err)
	}

	dirs := []string{
		env.TawsRoot,
		filepath.Join(env.TawsRoot, "configs"),
		filepath.Join(env.TawsRoot, "state"),
		filepath.Join(env.TawsRoot, "logs"),
		env.WebRoot,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("directory %s should exist", dir)
		}
	}
}

func TestFormatSetupResult(t *testing.T) {
	result := &SetupResult{
		Success: true,
		Messages: []string{
			"Test message 1",
			"Test message 2",
		},
	}

	output := FormatSetupResult(result)

	if output == "" {
		t.Fatal("FormatSetupResult returned empty string")
	}

	if len(output) < 20 {
		t.Error("FormatSetupResult output too short")
	}
}

func TestGenerateDefaultConfig(t *testing.T) {
	env := &Environment{
		HomeDir:  t.TempDir(),
		TawsRoot: filepath.Join(t.TempDir(), ".taws"),
		WebRoot:  filepath.Join(t.TempDir(), "web"),
	}

	os.MkdirAll(filepath.Join(env.TawsRoot, "configs"), 0755)

	err := generateDefaultConfig(env)
	if err != nil {
		t.Fatalf("generateDefaultConfig failed: %v", err)
	}

	configPath := filepath.Join(env.TawsRoot, "configs", "taws.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file should exist")
	}
}
