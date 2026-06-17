package adminer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	dataDir   string
	logDir    string
	port      int
	publicDir string
}

type Status struct {
	Installed bool
	Running   bool
	Port      int
	URL       string
	Version   string
}

func NewManager(dataDir, logDir string, port int) *Manager {
	return &Manager{
		dataDir:   dataDir,
		logDir:    logDir,
		port:      port,
		publicDir: filepath.Join(dataDir, "adminer"),
	}
}

const adminerChecksum = "2fd7e6d8f987b243ab1839249551f62adce19704c47d3d0c8dd9e57ea5b9c6b3"

func (m *Manager) Download() error {
	if err := os.MkdirAll(m.publicDir, 0755); err != nil {
		return fmt.Errorf("creating adminer directory: %w", err)
	}

	adminerFile := filepath.Join(m.publicDir, "index.php")

	if _, err := os.Stat(adminerFile); err == nil {
		return nil
	}

	cmd := exec.Command("wget", "-q", "https://github.com/vrana/adminer/releases/download/v4.8.1/adminer-4.8.1.php", "-O", adminerFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("downloading adminer: %s - %w", strings.TrimSpace(string(output)), err)
	}

	if err := m.verifyChecksum(adminerFile, adminerChecksum); err != nil {
		os.Remove(adminerFile)
		return fmt.Errorf("adminer checksum verification failed: %w", err)
	}

	return nil
}

func (m *Manager) verifyChecksum(filePath, expectedChecksum string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(data)
	actualChecksum := hex.EncodeToString(hash[:])

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

func (m *Manager) IsInstalled() bool {
	adminerFile := filepath.Join(m.publicDir, "index.php")
	_, err := os.Stat(adminerFile)
	return err == nil
}

func (m *Manager) GetVersion() string {
	if !m.IsInstalled() {
		return "not installed"
	}

	adminerFile := filepath.Join(m.publicDir, "index.php")
	data, err := os.ReadFile(adminerFile)
	if err != nil {
		return "unknown"
	}

	content := string(data)
	if idx := strings.Index(content, "Adminer "); idx != -1 {
		end := strings.Index(content[idx:], "\"")
		if end != -1 {
			return content[idx : idx+end]
		}
	}

	return "installed"
}

func (m *Manager) GetStatus() *Status {
	return &Status{
		Installed: m.IsInstalled(),
		Running:   false,
		Port:      m.port,
		URL:       fmt.Sprintf("http://127.0.0.1:%d", m.port),
		Version:   m.GetVersion(),
	}
}

func (m *Manager) GetPublicDir() string {
	return m.publicDir
}

func (m *Manager) GetConfig() map[string]string {
	return map[string]string{
		"server":   "127.0.0.1",
		"db":       "",
		"username": "",
		"password": "",
	}
}

func FormatStatus(status *Status) string {
	var sb strings.Builder

	sb.WriteString("TAWS Adminer Status\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	installed := "NO"
	if status.Installed {
		installed = "YES"
	}

	running := "NO"
	if status.Running {
		running = "YES"
	}

	sb.WriteString(fmt.Sprintf("  Installed:   %s\n", installed))
	sb.WriteString(fmt.Sprintf("  Running:     %s\n", running))
	sb.WriteString(fmt.Sprintf("  Port:        %d\n", status.Port))
	sb.WriteString(fmt.Sprintf("  URL:         %s\n", status.URL))
	sb.WriteString(fmt.Sprintf("  Version:     %s\n", status.Version))

	sb.WriteString("\n" + strings.Repeat("=", 50) + "\n")

	return sb.String()
}
