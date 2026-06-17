package projects

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"taws/internal/state"
)

func setupTestManager(t *testing.T) (*Manager, string) {
	t.Helper()

	tmpDir := t.TempDir()
	webRoot := filepath.Join(tmpDir, "web")
	stateDir := filepath.Join(tmpDir, "state")

	os.MkdirAll(webRoot, 0755)
	os.MkdirAll(stateDir, 0755)

	stateMgr := state.NewManager(stateDir)
	m := NewManager(webRoot, stateMgr)

	return m, webRoot
}

func TestCreateProject(t *testing.T) {
	m, webRoot := setupTestManager(t)

	p, err := m.Create("testproject")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if p.Name != "testproject" {
		t.Errorf("expected name testproject, got %s", p.Name)
	}

	if p.Domain != "testproject.test" {
		t.Errorf("expected domain testproject.test, got %s", p.Domain)
	}

	projectDir := filepath.Join(webRoot, "testproject")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Error("project directory should exist")
	}

	indexFile := filepath.Join(projectDir, "index.php")
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		t.Error("index.php should exist")
	}
}

func TestCreateProjectEmptyName(t *testing.T) {
	m, _ := setupTestManager(t)

	_, err := m.Create("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestCreateProjectInvalidName(t *testing.T) {
	m, _ := setupTestManager(t)

	tests := []string{
		"-invalid",
		"invalid-",
		"Invalid",
		"has space",
		"special!",
	}

	for _, name := range tests {
		_, err := m.Create(name)
		if err == nil {
			t.Errorf("expected error for invalid name: %s", name)
		}
	}
}

func TestCreateProjectDuplicate(t *testing.T) {
	m, _ := setupTestManager(t)

	_, err := m.Create("testproject")
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	_, err = m.Create("testproject")
	if err == nil {
		t.Error("expected error for duplicate project")
	}
}

func TestDeleteProject(t *testing.T) {
	m, webRoot := setupTestManager(t)

	_, err := m.Create("testproject")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := m.Delete("testproject"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	projectDir := filepath.Join(webRoot, "testproject")
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Error("project directory should be removed")
	}
}

func TestDeleteProjectEmptyName(t *testing.T) {
	m, _ := setupTestManager(t)

	err := m.Delete("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestDeleteProjectNotFound(t *testing.T) {
	m, _ := setupTestManager(t)

	err := m.Delete("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestListProjects(t *testing.T) {
	m, _ := setupTestManager(t)

	_, err := m.Create("project1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	_, err = m.Create("project2")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	projects, err := m.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestListProjectsEmpty(t *testing.T) {
	m, _ := setupTestManager(t)

	projects, err := m.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestGetProject(t *testing.T) {
	m, _ := setupTestManager(t)

	_, err := m.Create("testproject")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	p, err := m.Get("testproject")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if p.Name != "testproject" {
		t.Errorf("expected name testproject, got %s", p.Name)
	}
}

func TestGetProjectNotFound(t *testing.T) {
	m, _ := setupTestManager(t)

	_, err := m.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"valid", true},
		{"valid-name", true},
		{"name123", true},
		{"123name", true},
		{"-invalid", false},
		{"invalid-", false},
		{"Invalid", false},
		{"has space", false},
		{"special!", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidName(tt.name)
			if result != tt.expected {
				t.Errorf("isValidName(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestFormatProjects(t *testing.T) {
	projects := []Project{
		{
			Name:      "blog",
			Path:      "/home/user/web/blog",
			Domain:    "blog.test",
			CreatedAt: time.Now(),
		},
	}

	output := FormatProjects(projects)

	if output == "" {
		t.Fatal("FormatProjects returned empty string")
	}

	if len(output) < 30 {
		t.Error("FormatProjects output too short")
	}
}

func TestFormatProjectsEmpty(t *testing.T) {
	output := FormatProjects([]Project{})

	if output == "" {
		t.Fatal("FormatProjects returned empty string")
	}
}
