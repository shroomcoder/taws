package services

import (
	"fmt"
	"os/exec"
	"strings"
)

type ServiceName string

const (
	Nginx    ServiceName = "nginx"
	PHP      ServiceName = "php-fpm"
	MariaDB  ServiceName = "mariadb"
	Adminer  ServiceName = "adminer"
)

type ServiceStatus struct {
	Name    ServiceName
	Running bool
	PID     string
	Error   string
}

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Start(service ServiceName) error {
	name := string(service)
	if service == MariaDB {
		name = "mariadb"
	}

	cmd := exec.Command("sudo", "service", name, "start")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start %s: %s - %w", service, strings.TrimSpace(string(output)), err)
	}

	return nil
}

func (m *Manager) Stop(service ServiceName) error {
	name := string(service)
	if service == MariaDB {
		name = "mariadb"
	}

	cmd := exec.Command("sudo", "service", name, "stop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop %s: %s - %w", service, strings.TrimSpace(string(output)), err)
	}

	return nil
}

func (m *Manager) Restart(service ServiceName) error {
	name := string(service)
	if service == MariaDB {
		name = "mariadb"
	}

	cmd := exec.Command("sudo", "service", name, "restart")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart %s: %s - %w", service, strings.TrimSpace(string(output)), err)
	}

	return nil
}

func (m *Manager) Status(service ServiceName) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name: service,
	}

	name := string(service)
	if service == MariaDB {
		name = "mariadb"
	}

	cmd := exec.Command("service", name, "status")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		if strings.Contains(outputStr, "not found") || strings.Contains(outputStr, "not running") {
			status.Running = false
			return status, nil
		}
		status.Error = strings.TrimSpace(outputStr)
		return status, nil
	}

	status.Running = strings.Contains(outputStr, "running") || strings.Contains(outputStr, "active")

	if pid := extractPID(outputStr); pid != "" {
		status.PID = pid
	}

	return status, nil
}

func extractPID(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Main PID:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				pid := parts[2]
				pid = strings.TrimSuffix(pid, ",")
				return pid
			}
		}
		if idx := strings.Index(line, "(pid "); idx != -1 {
			rest := line[idx+5:]
			if endIdx := strings.Index(rest, ")"); endIdx != -1 {
				return rest[:endIdx]
			}
		}
		if idx := strings.Index(line, "PID: "); idx != -1 {
			rest := line[idx+5:]
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				return strings.TrimSuffix(parts[0], ",")
			}
		}
	}
	return ""
}

func (m *Manager) StartAll() []error {
	services := []ServiceName{Nginx, PHP, MariaDB}
	var errs []error

	for _, svc := range services {
		if err := m.Start(svc); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", svc, err))
		}
	}

	return errs
}

func (m *Manager) StopAll() []error {
	services := []ServiceName{Adminer, PHP, Nginx, MariaDB}
	var errs []error

	for _, svc := range services {
		if err := m.Stop(svc); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", svc, err))
		}
	}

	return errs
}

func (m *Manager) StatusAll() []ServiceStatus {
	services := []ServiceName{Nginx, PHP, MariaDB, Adminer}
	statuses := make([]ServiceStatus, 0, len(services))

	for _, svc := range services {
		status, _ := m.Status(svc)
		if status != nil {
			statuses = append(statuses, *status)
		}
	}

	return statuses
}

func FormatStatus(statuses []ServiceStatus) string {
	var sb strings.Builder

	sb.WriteString("TAWS Service Status\n")
	sb.WriteString(strings.Repeat("=", 45) + "\n\n")

	for _, s := range statuses {
		icon := "[STOPPED]"
		if s.Running {
			icon = "[RUNNING]"
		}

		line := fmt.Sprintf("  %-12s %s", s.Name, icon)
		if s.PID != "" {
			line += fmt.Sprintf(" (PID: %s)", s.PID)
		}
		if s.Error != "" {
			line += fmt.Sprintf(" - %s", s.Error)
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString(strings.Repeat("=", 45) + "\n")

	return sb.String()
}
