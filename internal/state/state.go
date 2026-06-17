package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Manager struct {
	stateDir string
}

type State struct {
	Installed   bool              `json:"installed"`
	Version     string            `json:"version"`
	LastUpdated time.Time         `json:"last_updated"`
	Services    map[string]bool   `json:"services"`
	Projects    []ProjectState    `json:"projects"`
}

type ProjectState struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
}

func NewManager(stateDir string) *Manager {
	return &Manager{stateDir: stateDir}
}

func (m *Manager) stateFile() string {
	return filepath.Join(m.stateDir, "taws.json")
}

func (m *Manager) Load() (*State, error) {
	data, err := os.ReadFile(m.stateFile())
	if err != nil {
		if os.IsNotExist(err) {
			return m.defaultState(), nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	s := &State{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}

	return s, nil
}

func (m *Manager) Save(s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.MkdirAll(m.stateDir, 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	if err := os.WriteFile(m.stateFile(), data, 0600); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}

func (m *Manager) SetInstalled(version string) error {
	s, err := m.Load()
	if err != nil {
		return err
	}

	s.Installed = true
	s.Version = version
	s.LastUpdated = time.Now()

	return m.Save(s)
}

func (m *Manager) SetService(name string, running bool) error {
	s, err := m.Load()
	if err != nil {
		return err
	}

	if s.Services == nil {
		s.Services = make(map[string]bool)
	}

	s.Services[name] = running
	return m.Save(s)
}

func (m *Manager) AddProject(p ProjectState) error {
	s, err := m.Load()
	if err != nil {
		return err
	}

	s.Projects = append(s.Projects, p)
	return m.Save(s)
}

func (m *Manager) RemoveProject(name string) error {
	s, err := m.Load()
	if err != nil {
		return err
	}

	for i, p := range s.Projects {
		if p.Name == name {
			s.Projects = append(s.Projects[:i], s.Projects[i+1:]...)
			break
		}
	}

	return m.Save(s)
}

func (m *Manager) defaultState() *State {
	return &State{
		Installed: false,
		Version:   "",
		Services:  make(map[string]bool),
		Projects:  []ProjectState{},
	}
}
