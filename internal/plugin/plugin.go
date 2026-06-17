package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"taws/internal/validate"
)

type Manager struct {
	pluginDir string
	stateFile string
}

type PluginInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Enabled     bool      `json:"enabled"`
	Installed   bool      `json:"installed"`
	InstalledAt time.Time `json:"installed_at,omitempty"`
}

type PluginState struct {
	Plugins []PluginInfo `json:"plugins"`
}

type Plugin interface {
	Name() string
	Version() string
	Description() string
	Author() string
	Init(configDir string) error
	Start() error
	Stop() error
	Status() string
}

func NewManager(pluginDir string) *Manager {
	return &Manager{
		pluginDir: pluginDir,
		stateFile: filepath.Join(pluginDir, "plugins.json"),
	}
}

func (m *Manager) Install(name string) error {
	if !validate.PluginName(name) {
		return fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	if err := os.MkdirAll(m.pluginDir, 0755); err != nil {
		return fmt.Errorf("creating plugin directory: %w", err)
	}

	state, err := m.loadState()
	if err != nil {
		return err
	}

	for _, p := range state.Plugins {
		if p.Name == name {
			if p.Installed {
				return fmt.Errorf("plugin %s is already installed", name)
			}
		}
	}

	pluginInfo := PluginInfo{
		Name:        name,
		Version:     "1.0.0",
		Description: fmt.Sprintf("Plugin: %s", name),
		Installed:   true,
		Enabled:     false,
		InstalledAt: time.Now(),
	}

	state.Plugins = append(state.Plugins, pluginInfo)

	if err := m.saveState(state); err != nil {
		return err
	}

	return nil
}

func (m *Manager) Remove(name string) error {
	if !validate.PluginName(name) {
		return fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	state, err := m.loadState()
	if err != nil {
		return err
	}

	for i, p := range state.Plugins {
		if p.Name == name {
			if !p.Installed {
				return fmt.Errorf("plugin %s is not installed", name)
			}

			state.Plugins = append(state.Plugins[:i], state.Plugins[i+1:]...)

			pluginPath := filepath.Join(m.pluginDir, name+".so")
			absPluginPath, err := filepath.Abs(pluginPath)
			if err != nil {
				return fmt.Errorf("resolving plugin path: %w", err)
			}
			absPluginDir, err := filepath.Abs(m.pluginDir)
			if err != nil {
				return fmt.Errorf("resolving plugin dir: %w", err)
			}
			if !strings.HasPrefix(absPluginPath, absPluginDir+string(filepath.Separator)) && absPluginPath != absPluginDir {
				return fmt.Errorf("path traversal attempt detected")
			}

			os.Remove(pluginPath)

			return m.saveState(state)
		}
	}

	return fmt.Errorf("plugin %s not found", name)
}

func (m *Manager) Enable(name string) error {
	if !validate.PluginName(name) {
		return fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	state, err := m.loadState()
	if err != nil {
		return err
	}

	for i, p := range state.Plugins {
		if p.Name == name {
			if !p.Installed {
				return fmt.Errorf("plugin %s is not installed", name)
			}

			state.Plugins[i].Enabled = true
			return m.saveState(state)
		}
	}

	return fmt.Errorf("plugin %s not found", name)
}

func (m *Manager) Disable(name string) error {
	if !validate.PluginName(name) {
		return fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	state, err := m.loadState()
	if err != nil {
		return err
	}

	for i, p := range state.Plugins {
		if p.Name == name {
			if !p.Installed {
				return fmt.Errorf("plugin %s is not installed", name)
			}

			state.Plugins[i].Enabled = false
			return m.saveState(state)
		}
	}

	return fmt.Errorf("plugin %s not found", name)
}

func (m *Manager) Update(name string) error {
	if !validate.PluginName(name) {
		return fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	state, err := m.loadState()
	if err != nil {
		return err
	}

	for i, p := range state.Plugins {
		if p.Name == name {
			if !p.Installed {
				return fmt.Errorf("plugin %s is not installed", name)
			}

			state.Plugins[i].Version = "1.1.0"
			return m.saveState(state)
		}
	}

	return fmt.Errorf("plugin %s not found", name)
}

func (m *Manager) List() ([]PluginInfo, error) {
	state, err := m.loadState()
	if err != nil {
		return nil, err
	}

	return state.Plugins, nil
}

func (m *Manager) Get(name string) (*PluginInfo, error) {
	if !validate.PluginName(name) {
		return nil, fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	state, err := m.loadState()
	if err != nil {
		return nil, err
	}

	for i := range state.Plugins {
		if state.Plugins[i].Name == name {
			return &state.Plugins[i], nil
		}
	}

	return nil, fmt.Errorf("plugin %s not found", name)
}

func (m *Manager) IsInstalled(name string) bool {
	if !validate.PluginName(name) {
		return false
	}

	state, err := m.loadState()
	if err != nil {
		return false
	}

	for _, p := range state.Plugins {
		if p.Name == name && p.Installed {
			return true
		}
	}

	return false
}

func (m *Manager) IsEnabled(name string) bool {
	if !validate.PluginName(name) {
		return false
	}

	state, err := m.loadState()
	if err != nil {
		return false
	}

	for _, p := range state.Plugins {
		if p.Name == name && p.Enabled {
			return true
		}
	}

	return false
}

func (m *Manager) LoadPlugin(name string) (Plugin, error) {
	if !validate.PluginName(name) {
		return nil, fmt.Errorf("invalid plugin name: use letters, numbers, hyphens, underscores")
	}

	pluginPath := filepath.Join(m.pluginDir, name+".so")

	absPluginPath, err := filepath.Abs(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("resolving plugin path: %w", err)
	}
	absPluginDir, err := filepath.Abs(m.pluginDir)
	if err != nil {
		return nil, fmt.Errorf("resolving plugin dir: %w", err)
	}
	if !strings.HasPrefix(absPluginPath, absPluginDir+string(filepath.Separator)) && absPluginPath != absPluginDir {
		return nil, fmt.Errorf("path traversal attempt detected")
	}

	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("opening plugin %s: %w", name, err)
	}

	symbol, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("looking up Plugin symbol: %w", err)
	}

	pluginInstance, ok := symbol.(Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin %s does not implement Plugin interface", name)
	}

	return pluginInstance, nil
}

func (m *Manager) loadState() (*PluginState, error) {
	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &PluginState{Plugins: []PluginInfo{}}, nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	state := &PluginState{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}

	return state, nil
}

func (m *Manager) saveState(state *PluginState) error {
	if err := os.MkdirAll(m.pluginDir, 0755); err != nil {
		return fmt.Errorf("creating plugin directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(m.stateFile, data, 0600); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}

func FormatPlugins(plugins []PluginInfo) string {
	var sb strings.Builder

	sb.WriteString("TAWS Plugins\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(plugins) == 0 {
		sb.WriteString("  No plugins installed.\n")
		sb.WriteString("  Install one with: taws plugin install <name>\n")
	} else {
		for _, p := range plugins {
			status := "DISABLED"
			if p.Enabled {
				status = "ENABLED"
			}

			sb.WriteString(fmt.Sprintf("  %-15s v%-8s [%s]\n", p.Name, p.Version, status))
			sb.WriteString(fmt.Sprintf("  %-15s %s\n\n", "", p.Description))
		}
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")

	return sb.String()
}

var AvailablePlugins = []string{
	"apache",
	"postgresql",
	"redis",
	"nodejs",
	"phpmyadmin",
	"mailhog",
}
