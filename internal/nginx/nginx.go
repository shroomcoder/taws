package nginx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"taws/internal/validate"
)

type Config struct {
	Name       string
	Domain     string
	Root       string
	Port       int
	PHPFPMPort int
	LogDir     string
}

type Manager struct {
	configDir  string
	sitesDir   string
	logDir     string
	templatesDir string
}

func NewManager(configDir, sitesDir, logDir, templatesDir string) *Manager {
	return &Manager{
		configDir:    configDir,
		sitesDir:     sitesDir,
		logDir:       logDir,
		templatesDir: templatesDir,
	}
}

func (m *Manager) GenerateVhost(cfg *Config) error {
	if !validate.ProjectName(cfg.Name) {
		return fmt.Errorf("invalid vhost name: use lowercase letters, numbers, and hyphens")
	}
	if !validate.SafeTemplateString(cfg.Domain) {
		return fmt.Errorf("invalid domain: must not contain newlines or control characters")
	}
	if !validate.SafeTemplateString(cfg.Root) {
		return fmt.Errorf("invalid root path: must not contain newlines or control characters")
	}

	tmplPath := filepath.Join(m.templatesDir, "nginx", "vhost.conf.tmpl")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	outputPath := filepath.Join(m.sitesDir, cfg.Name+".conf")

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

func (m *Manager) RemoveVhost(name string) error {
	if !validate.ProjectName(name) {
		return fmt.Errorf("invalid vhost name: use lowercase letters, numbers, and hyphens")
	}

	configPath := filepath.Join(m.sitesDir, name+".conf")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("vhost %s does not exist", name)
	}

	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("removing vhost config: %w", err)
	}

	return nil
}

func (m *Manager) ListVhosts() ([]string, error) {
	entries, err := os.ReadDir(m.sitesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("reading sites directory: %w", err)
	}

	var vhosts []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".conf") {
			name := strings.TrimSuffix(entry.Name(), ".conf")
			vhosts = append(vhosts, name)
		}
	}

	return vhosts, nil
}

func (m *Manager) EnableSite(name string) error {
	if !validate.ProjectName(name) {
		return fmt.Errorf("invalid site name: use lowercase letters, numbers, and hyphens")
	}

	src := filepath.Join(m.sitesDir, name+".conf")
	dst := filepath.Join(m.configDir, "sites-enabled", name+".conf")

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("site %s does not exist", name)
	}

	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating sites-enabled directory: %w", err)
	}

	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing existing symlink: %w", err)
	}

	if err := os.Symlink(src, dst); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}

	return nil
}

func (m *Manager) DisableSite(name string) error {
	if !validate.ProjectName(name) {
		return fmt.Errorf("invalid site name: use lowercase letters, numbers, and hyphens")
	}

	link := filepath.Join(m.configDir, "sites-enabled", name+".conf")

	if _, err := os.Stat(link); os.IsNotExist(err) {
		return fmt.Errorf("site %s is not enabled", name)
	}

	if err := os.Remove(link); err != nil {
		return fmt.Errorf("removing symlink: %w", err)
	}

	return nil
}

func (m *Manager) TestConfig() error {
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx config test failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func (m *Manager) Reload() error {
	cmd := exec.Command("sudo", "nginx", "-s", "reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx reload failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func FormatVhosts(vhosts []string) string {
	var sb strings.Builder

	sb.WriteString("TAWS Nginx Virtual Hosts\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(vhosts) == 0 {
		sb.WriteString("  No virtual hosts configured.\n")
	} else {
		for _, v := range vhosts {
			sb.WriteString(fmt.Sprintf("  %-20s %s.test\n", v, v))
		}
	}

	sb.WriteString("\n" + strings.Repeat("=", 50) + "\n")

	return sb.String()
}
