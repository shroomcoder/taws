package state

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefault(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	s, err := m.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if s.Installed {
		t.Error("expected installed to be false")
	}

	if len(s.Projects) != 0 {
		t.Error("expected empty projects")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	s := &State{
		Installed:   true,
		Version:     "1.0.0",
		LastUpdated: time.Now(),
		Services:    map[string]bool{"nginx": true},
		Projects:    []ProjectState{{Name: "test", Domain: "test.test"}},
	}

	if err := m.Save(s); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := m.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if !loaded.Installed {
		t.Error("expected installed to be true")
	}

	if loaded.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", loaded.Version)
	}

	if !loaded.Services["nginx"] {
		t.Error("expected nginx to be true")
	}

	if len(loaded.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(loaded.Projects))
	}
}

func TestSetInstalled(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	if err := m.SetInstalled("1.0.0"); err != nil {
		t.Fatalf("SetInstalled failed: %v", err)
	}

	s, _ := m.Load()
	if !s.Installed {
		t.Error("expected installed to be true")
	}
}

func TestSetService(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	if err := m.SetService("nginx", true); err != nil {
		t.Fatalf("SetService failed: %v", err)
	}

	s, _ := m.Load()
	if !s.Services["nginx"] {
		t.Error("expected nginx to be true")
	}
}

func TestAddAndRemoveProject(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	p := ProjectState{
		Name:      "blog",
		Path:      "/home/user/web/blog",
		Domain:    "blog.test",
		CreatedAt: time.Now(),
	}

	if err := m.AddProject(p); err != nil {
		t.Fatalf("AddProject failed: %v", err)
	}

	s, _ := m.Load()
	if len(s.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(s.Projects))
	}

	if err := m.RemoveProject("blog"); err != nil {
		t.Fatalf("RemoveProject failed: %v", err)
	}

	s, _ = m.Load()
	if len(s.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(s.Projects))
	}
}

func TestStateFileLocation(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)

	expected := filepath.Join(dir, "taws.json")
	actual := m.stateFile()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}
