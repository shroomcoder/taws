package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taws/internal/config"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	okStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD43B"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)
)

type ServiceInfo struct {
	Name    string
	Running bool
	Status  string
	Port    int
}

type ProjectInfo struct {
	Name   string
	Domain string
	Path   string
}

type SnapshotInfo struct {
	ID        string
	Timestamp string
	Version   string
}

type DashboardData struct {
	Services   []ServiceInfo
	Projects   []ProjectInfo
	Snapshots  []SnapshotInfo
	Version    string
	Hostname   string
	Uptime     string
	MemoryUsed string
	MemoryTotal string
	Diskspace  string
	LoadAvg    string
}

type tickMsg struct{}

type Model struct {
	data     DashboardData
	width    int
	height   int
	quitting bool
	err      error
}

func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchData(), tickEvery(5*time.Second))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case dashboardMsg:
		m.data = msg.data
		return m, tickEvery(5*time.Second)

	case tickMsg:
		return m, m.fetchData()

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var sb strings.Builder

	sb.WriteString(titleStyle.Render(" TAWS Dashboard "))
	sb.WriteString("\n\n")

	sb.WriteString(m.renderHeader())
	sb.WriteString("\n")
	sb.WriteString(m.renderServices())
	sb.WriteString("\n")
	sb.WriteString(m.renderProjects())
	sb.WriteString("\n")
	sb.WriteString(m.renderBackups())
	sb.WriteString("\n")
	sb.WriteString(m.renderFooter())

	return sb.String()
}

func (m Model) renderHeader() string {
	var sb strings.Builder

	sb.WriteString(sectionStyle.Render("System Information"))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("  Hostname:  %s\n", m.data.Hostname))
	sb.WriteString(fmt.Sprintf("  Version:   %s\n", m.data.Version))
	sb.WriteString(fmt.Sprintf("  OS:        %s/%s\n", runtime.GOOS, runtime.GOARCH))
	sb.WriteString(fmt.Sprintf("  Memory:    %s / %s\n", m.data.MemoryUsed, m.data.MemoryTotal))
	sb.WriteString(fmt.Sprintf("  Disk:      %s\n", m.data.Diskspace))
	sb.WriteString(fmt.Sprintf("  Load:      %s\n", m.data.LoadAvg))

	return sb.String()
}

func (m Model) renderServices() string {
	var sb strings.Builder

	sb.WriteString(sectionStyle.Render("Services"))
	sb.WriteString("\n")

	for _, svc := range m.data.Services {
		icon := errorStyle.Render("[STOPPED]")
		if svc.Running {
			icon = okStyle.Render("[RUNNING]")
		}

		portInfo := ""
		if svc.Port > 0 {
			portInfo = fmt.Sprintf(" :%d", svc.Port)
		}

		sb.WriteString(fmt.Sprintf("  %-12s %s%s\n", svc.Name, icon, portInfo))
	}

	return sb.String()
}

func (m Model) renderProjects() string {
	var sb strings.Builder

	sb.WriteString(sectionStyle.Render("Projects"))
	sb.WriteString("\n")

	if len(m.data.Projects) == 0 {
		sb.WriteString("  No projects configured.\n")
	} else {
		for _, proj := range m.data.Projects {
			sb.WriteString(fmt.Sprintf("  %-15s %s.test\n", proj.Name, proj.Name))
		}
	}

	return sb.String()
}

func (m Model) renderBackups() string {
	var sb strings.Builder

	sb.WriteString(sectionStyle.Render("Backups"))
	sb.WriteString("\n")

	if len(m.data.Snapshots) == 0 {
		sb.WriteString("  No snapshots found.\n")
	} else {
		for _, snap := range m.data.Snapshots {
			sb.WriteString(fmt.Sprintf("  %s (v%s) - %s\n", snap.ID, snap.Version, snap.Timestamp))
		}
	}

	return sb.String()
}

func (m Model) renderFooter() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("Press 'q' to quit | Refreshes every 5 seconds")
}

type dashboardMsg struct {
	data DashboardData
}

type errMsg struct {
	err error
}

func (m Model) fetchData() tea.Cmd {
	return func() tea.Msg {
		data := DashboardData{
			Version:   "1.0.0",
			Hostname:  getHostname(),
			Services:  getServices(),
			Projects:  getProjects(),
			Snapshots: getSnapshots(),
		}

		data.MemoryUsed, data.MemoryTotal = getMemoryInfo()
		data.Diskspace = getDiskInfo()
		data.LoadAvg = getLoadAvg()

		return dashboardMsg{data: data}
	}
}

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func getServices() []ServiceInfo {
	cfg, _ := loadDashboardConfig()

	services := []ServiceInfo{
		{Name: "nginx", Running: false, Status: "stopped"},
		{Name: "php-fpm", Running: false, Status: "stopped"},
		{Name: "mariadb", Running: false, Status: "stopped"},
		{Name: "adminer", Running: false, Status: "stopped"},
	}

	if cfg != nil {
		services[0].Port = cfg.Services.Nginx.Port
		services[1].Port = cfg.Services.PHP.Port
		services[2].Port = cfg.Services.MariaDB.Port
		services[3].Port = cfg.Services.Adminer.Port
	}

	return services
}

func loadDashboardConfig() (*config.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cfgPath := filepath.Join(home, ".taws", "configs", "taws.yaml")
	return config.Load(cfgPath)
}

func getProjects() []ProjectInfo {
	return []ProjectInfo{}
}

func getSnapshots() []SnapshotInfo {
	return []SnapshotInfo{}
}

func getMemoryInfo() (string, string) {
	return "N/A", "N/A"
}

func getDiskInfo() string {
	return "N/A"
}

func getLoadAvg() string {
	return "N/A"
}

func RunDashboard() error {
	m := NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("dashboard error: %w", err)
	}

	return nil
}
