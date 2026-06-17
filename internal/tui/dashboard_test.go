package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel()

	if m.quitting {
		t.Error("new model should not be quitting")
	}

	if m.err != nil {
		t.Error("new model should have no error")
	}
}

func TestModelInit(t *testing.T) {
	m := NewModel()

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	m := NewModel()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, _ := m.Update(msg)

	model := updated.(Model)
	if model.width != 80 {
		t.Errorf("expected width 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("expected height 24, got %d", model.height)
	}
}

func TestModelUpdateQuit(t *testing.T) {
	m := NewModel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updated, _ := m.Update(msg)

	model := updated.(Model)
	if !model.quitting {
		t.Error("model should be quitting after 'q' key")
	}
}

func TestModelUpdateCtrlC(t *testing.T) {
	m := NewModel()

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updated, _ := m.Update(msg)

	model := updated.(Model)
	if !model.quitting {
		t.Error("model should be quitting after ctrl+c")
	}
}

func TestModelUpdateDashboardMsg(t *testing.T) {
	m := NewModel()

	data := DashboardData{
		Version:  "1.0.0",
		Hostname: "test-host",
	}

	msg := dashboardMsg{data: data}
	updated, _ := m.Update(msg)

	model := updated.(Model)
	if model.data.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", model.data.Version)
	}

	if model.data.Hostname != "test-host" {
		t.Errorf("expected hostname test-host, got %s", model.data.Hostname)
	}
}

func TestModelUpdateErrMsg(t *testing.T) {
	m := NewModel()

	msg := errMsg{err: nil}
	updated, _ := m.Update(msg)

	model := updated.(Model)
	if model.err != nil {
		t.Error("error should be nil")
	}
}

func TestModelView(t *testing.T) {
	m := NewModel()

	view := m.View()
	if view == "" {
		t.Error("View should return non-empty string")
	}
}

func TestModelViewQuitting(t *testing.T) {
	m := NewModel()
	m.quitting = true

	view := m.View()
	if view != "" {
		t.Error("View should return empty string when quitting")
	}
}

func TestModelViewWithError(t *testing.T) {
	m := NewModel()
	m.err = nil

	view := m.View()
	if view == "" {
		t.Error("View should return non-empty string")
	}
}

func TestGetHostname(t *testing.T) {
	hostname := getHostname()
	if hostname == "" {
		t.Error("getHostname should return non-empty string")
	}
}

func TestGetServices(t *testing.T) {
	services := getServices()
	if len(services) == 0 {
		t.Error("getServices should return at least one service")
	}
}

func TestGetProjects(t *testing.T) {
	projects := getProjects()
	if projects == nil {
		t.Error("getProjects should return non-nil slice")
	}
}

func TestGetSnapshots(t *testing.T) {
	snapshots := getSnapshots()
	if snapshots == nil {
		t.Error("getSnapshots should return non-nil slice")
	}
}

func TestGetMemoryInfo(t *testing.T) {
	used, total := getMemoryInfo()
	if used == "" || total == "" {
		t.Error("getMemoryInfo should return non-empty strings")
	}
}

func TestGetDiskInfo(t *testing.T) {
	disk := getDiskInfo()
	if disk == "" {
		t.Error("getDiskInfo should return non-empty string")
	}
}

func TestGetLoadAvg(t *testing.T) {
	load := getLoadAvg()
	if load == "" {
		t.Error("getLoadAvg should return non-empty string")
	}
}

func TestServiceInfo(t *testing.T) {
	svc := ServiceInfo{
		Name:    "nginx",
		Running: true,
		Status:  "running",
	}

	if svc.Name != "nginx" {
		t.Errorf("expected name nginx, got %s", svc.Name)
	}

	if !svc.Running {
		t.Error("expected running to be true")
	}
}

func TestProjectInfo(t *testing.T) {
	proj := ProjectInfo{
		Name:   "blog",
		Domain: "blog.test",
		Path:   "/home/user/web/blog",
	}

	if proj.Name != "blog" {
		t.Errorf("expected name blog, got %s", proj.Name)
	}
}

func TestSnapshotInfo(t *testing.T) {
	snap := SnapshotInfo{
		ID:        "20260101-120000",
		Timestamp: "2026-01-01 12:00:00",
		Version:   "1.0.0",
	}

	if snap.ID != "20260101-120000" {
		t.Errorf("expected ID 20260101-120000, got %s", snap.ID)
	}
}
