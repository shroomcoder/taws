package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestManager(t *testing.T) *Manager {
	t.Helper()

	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "plugins")

	os.MkdirAll(pluginDir, 0755)

	return NewManager(pluginDir)
}

func TestNewManager(t *testing.T) {
	m := setupTestManager(t)

	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestInstallPlugin(t *testing.T) {
	m := setupTestManager(t)

	if err := m.Install("testplugin"); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if !m.IsInstalled("testplugin") {
		t.Error("plugin should be installed")
	}
}

func TestInstallPluginDuplicate(t *testing.T) {
	m := setupTestManager(t)

	m.Install("testplugin")

	err := m.Install("testplugin")
	if err == nil {
		t.Error("expected error for duplicate install")
	}
}

func TestRemovePlugin(t *testing.T) {
	m := setupTestManager(t)

	m.Install("testplugin")

	if err := m.Remove("testplugin"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if m.IsInstalled("testplugin") {
		t.Error("plugin should not be installed")
	}
}

func TestRemovePluginNotFound(t *testing.T) {
	m := setupTestManager(t)

	err := m.Remove("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestEnablePlugin(t *testing.T) {
	m := setupTestManager(t)

	m.Install("testplugin")

	if err := m.Enable("testplugin"); err != nil {
		t.Fatalf("Enable failed: %v", err)
	}

	if !m.IsEnabled("testplugin") {
		t.Error("plugin should be enabled")
	}
}

func TestEnablePluginNotFound(t *testing.T) {
	m := setupTestManager(t)

	err := m.Enable("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestDisablePlugin(t *testing.T) {
	m := setupTestManager(t)

	m.Install("testplugin")
	m.Enable("testplugin")

	if err := m.Disable("testplugin"); err != nil {
		t.Fatalf("Disable failed: %v", err)
	}

	if m.IsEnabled("testplugin") {
		t.Error("plugin should be disabled")
	}
}

func TestDisablePluginNotFound(t *testing.T) {
	m := setupTestManager(t)

	err := m.Disable("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestUpdatePlugin(t *testing.T) {
	m := setupTestManager(t)

	m.Install("testplugin")

	if err := m.Update("testplugin"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	plugin, _ := m.Get("testplugin")
	if plugin.Version != "1.1.0" {
		t.Errorf("expected version 1.1.0, got %s", plugin.Version)
	}
}

func TestUpdatePluginNotFound(t *testing.T) {
	m := setupTestManager(t)

	err := m.Update("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestListPlugins(t *testing.T) {
	m := setupTestManager(t)

	m.Install("plugin1")
	m.Install("plugin2")

	plugins, err := m.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(plugins))
	}
}

func TestGetPlugin(t *testing.T) {
	m := setupTestManager(t)

	m.Install("testplugin")

	plugin, err := m.Get("testplugin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if plugin.Name != "testplugin" {
		t.Errorf("expected name testplugin, got %s", plugin.Name)
	}
}

func TestGetPluginNotFound(t *testing.T) {
	m := setupTestManager(t)

	_, err := m.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestIsInstalled(t *testing.T) {
	m := setupTestManager(t)

	if m.IsInstalled("testplugin") {
		t.Error("plugin should not be installed")
	}

	m.Install("testplugin")

	if !m.IsInstalled("testplugin") {
		t.Error("plugin should be installed")
	}
}

func TestIsEnabled(t *testing.T) {
	m := setupTestManager(t)

	m.Install("testplugin")

	if m.IsEnabled("testplugin") {
		t.Error("plugin should not be enabled by default")
	}

	m.Enable("testplugin")

	if !m.IsEnabled("testplugin") {
		t.Error("plugin should be enabled")
	}
}

func TestFormatPlugins(t *testing.T) {
	plugins := []PluginInfo{
		{
			Name:        "testplugin",
			Version:     "1.0.0",
			Description: "Test plugin",
			Enabled:     true,
			Installed:   true,
		},
	}

	output := FormatPlugins(plugins)

	if output == "" {
		t.Fatal("FormatPlugins returned empty string")
	}

	if len(output) < 30 {
		t.Error("FormatPlugins output too short")
	}
}

func TestFormatPluginsEmpty(t *testing.T) {
	output := FormatPlugins([]PluginInfo{})

	if output == "" {
		t.Fatal("FormatPlugins returned empty string")
	}
}

func TestAvailablePlugins(t *testing.T) {
	if len(AvailablePlugins) == 0 {
		t.Error("AvailablePlugins should not be empty")
	}
}
