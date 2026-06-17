package nginx

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestManager(t *testing.T) *Manager {
	t.Helper()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")
	sitesDir := filepath.Join(configDir, "nginx", "sites")
	logDir := filepath.Join(tmpDir, "logs")
	templatesDir := filepath.Join(tmpDir, "templates")

	os.MkdirAll(sitesDir, 0755)
	os.MkdirAll(logDir, 0755)
	os.MkdirAll(filepath.Join(templatesDir, "nginx"), 0755)

	tmpl := `server {
    listen {{.Port}};
    server_name {{.Domain}};
    root {{.Root}};
}`
	os.WriteFile(filepath.Join(templatesDir, "nginx", "vhost.conf.tmpl"), []byte(tmpl), 0644)

	return NewManager(configDir, sitesDir, logDir, templatesDir)
}

func TestGenerateVhost(t *testing.T) {
	m := setupTestManager(t)

	cfg := &Config{
		Name:       "blog",
		Domain:     "blog.test",
		Root:       "/home/user/web/blog",
		Port:       8080,
		PHPFPMPort: 9000,
		LogDir:     "/home/user/.taws/logs",
	}

	if err := m.GenerateVhost(cfg); err != nil {
		t.Fatalf("GenerateVhost failed: %v", err)
	}

	configPath := filepath.Join(m.sitesDir, "blog.conf")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("vhost config file should exist")
	}
}

func TestRemoveVhost(t *testing.T) {
	m := setupTestManager(t)

	cfg := &Config{
		Name:   "blog",
		Domain: "blog.test",
		Root:   "/home/user/web/blog",
		Port:   8080,
	}

	if err := m.GenerateVhost(cfg); err != nil {
		t.Fatalf("GenerateVhost failed: %v", err)
	}

	if err := m.RemoveVhost("blog"); err != nil {
		t.Fatalf("RemoveVhost failed: %v", err)
	}

	configPath := filepath.Join(m.sitesDir, "blog.conf")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("vhost config file should be removed")
	}
}

func TestRemoveVhostNotFound(t *testing.T) {
	m := setupTestManager(t)

	err := m.RemoveVhost("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent vhost")
	}
}

func TestListVhosts(t *testing.T) {
	m := setupTestManager(t)

	cfg1 := &Config{Name: "blog", Domain: "blog.test", Root: "/home/user/web/blog", Port: 8080}
	cfg2 := &Config{Name: "api", Domain: "api.test", Root: "/home/user/web/api", Port: 8080}

	m.GenerateVhost(cfg1)
	m.GenerateVhost(cfg2)

	vhosts, err := m.ListVhosts()
	if err != nil {
		t.Fatalf("ListVhosts failed: %v", err)
	}

	if len(vhosts) != 2 {
		t.Errorf("expected 2 vhosts, got %d", len(vhosts))
	}
}

func TestListVhostsEmpty(t *testing.T) {
	m := setupTestManager(t)

	vhosts, err := m.ListVhosts()
	if err != nil {
		t.Fatalf("ListVhosts failed: %v", err)
	}

	if len(vhosts) != 0 {
		t.Errorf("expected 0 vhosts, got %d", len(vhosts))
	}
}

func TestEnableDisableSite(t *testing.T) {
	m := setupTestManager(t)

	os.MkdirAll(filepath.Join(m.configDir, "sites-enabled"), 0755)

	cfg := &Config{Name: "blog", Domain: "blog.test", Root: "/home/user/web/blog", Port: 8080}
	m.GenerateVhost(cfg)

	if err := m.EnableSite("blog"); err != nil {
		t.Fatalf("EnableSite failed: %v", err)
	}

	link := filepath.Join(m.configDir, "sites-enabled", "blog.conf")
	if _, err := os.Lstat(link); os.IsNotExist(err) {
		t.Error("symlink should exist")
	}

	if err := m.DisableSite("blog"); err != nil {
		t.Fatalf("DisableSite failed: %v", err)
	}

	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Error("symlink should be removed")
	}
}

func TestEnableSiteNotFound(t *testing.T) {
	m := setupTestManager(t)

	err := m.EnableSite("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent site")
	}
}

func TestDisableSiteNotFound(t *testing.T) {
	m := setupTestManager(t)

	err := m.DisableSite("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent site")
	}
}

func TestFormatVhosts(t *testing.T) {
	vhosts := []string{"blog", "api"}

	output := FormatVhosts(vhosts)

	if output == "" {
		t.Fatal("FormatVhosts returned empty string")
	}

	if len(output) < 30 {
		t.Error("FormatVhosts output too short")
	}
}

func TestFormatVhostsEmpty(t *testing.T) {
	output := FormatVhosts([]string{})

	if output == "" {
		t.Fatal("FormatVhosts returned empty string")
	}
}
