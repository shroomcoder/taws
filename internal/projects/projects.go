package projects

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"taws/internal/state"
	"taws/internal/validate"
)

type Manager struct {
	webRoot  string
	stateMgr *state.Manager
}

type Project struct {
	Name      string
	Path      string
	Domain    string
	CreatedAt time.Time
}

func NewManager(webRoot string, stateMgr *state.Manager) *Manager {
	return &Manager{
		webRoot:  webRoot,
		stateMgr: stateMgr,
	}
}

func (m *Manager) Create(name string) (*Project, error) {
	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	if !isValidName(name) {
		return nil, fmt.Errorf("invalid project name: %s (use lowercase letters, numbers, and hyphens)", name)
	}

	projectPath := filepath.Join(m.webRoot, name)

	if _, err := os.Stat(projectPath); err == nil {
		return nil, fmt.Errorf("project %s already exists", name)
	}

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return nil, fmt.Errorf("creating project directory: %w", err)
	}

	indexContent := fmt.Sprintf(`<?php
// %s - Created by TAWS
// Project: %s
// Domain: %s.test

phpinfo();
`, name, name, name)

	if err := os.WriteFile(filepath.Join(projectPath, "index.php"), []byte(indexContent), 0644); err != nil {
		return nil, fmt.Errorf("creating index.php: %w", err)
	}

	domain := name + ".test"
	p := &Project{
		Name:      name,
		Path:      projectPath,
		Domain:    domain,
		CreatedAt: time.Now(),
	}

	stateProj := state.ProjectState{
		Name:      p.Name,
		Path:      p.Path,
		Domain:    p.Domain,
		CreatedAt: p.CreatedAt,
	}

	if err := m.stateMgr.AddProject(stateProj); err != nil {
		return nil, fmt.Errorf("saving project state: %w", err)
	}

	return p, nil
}

func (m *Manager) Delete(name string) error {
	if !isValidName(name) {
		return fmt.Errorf("invalid project name: %s", name)
	}

	projectPath := filepath.Join(m.webRoot, name)

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project %s does not exist", name)
	}

	if err := os.RemoveAll(projectPath); err != nil {
		return fmt.Errorf("removing project directory: %w", err)
	}

	if err := m.stateMgr.RemoveProject(name); err != nil {
		return fmt.Errorf("removing project state: %w", err)
	}

	return nil
}

func (m *Manager) List() ([]Project, error) {
	stateProj, err := m.stateMgr.Load()
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}

	var projects []Project
	for _, sp := range stateProj.Projects {
		projects = append(projects, Project{
			Name:      sp.Name,
			Path:      sp.Path,
			Domain:    sp.Domain,
			CreatedAt: sp.CreatedAt,
		})
	}

	return projects, nil
}

func (m *Manager) Get(name string) (*Project, error) {
	projects, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, p := range projects {
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("project %s not found", name)
}

func isValidName(name string) bool {
	return validate.ProjectName(name)
}

func FormatProjects(projects []Project) string {
	var sb strings.Builder

	sb.WriteString("TAWS Projects\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(projects) == 0 {
		sb.WriteString("  No projects found.\n")
		sb.WriteString("  Create one with: taws project create <name>\n")
	} else {
		for _, p := range projects {
			sb.WriteString(fmt.Sprintf("  %-15s %s.test\n", p.Name, p.Name))
			sb.WriteString(fmt.Sprintf("  %-15s %s\n", "", p.Path))
			sb.WriteString(fmt.Sprintf("  %-15s Created: %s\n\n", "", p.CreatedAt.Format("2006-01-02 15:04:05")))
		}
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")

	return sb.String()
}
