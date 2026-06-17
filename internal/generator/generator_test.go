package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	os.MkdirAll(templatesDir, 0755)

	tmpl := `server {
    listen {{.Project.Port}};
    server_name {{.Project.Domain}};
    root {{.Project.Path}};
}`
	os.WriteFile(filepath.Join(templatesDir, "nginx.conf.tmpl"), []byte(tmpl), 0644)

	g := NewGenerator(templatesDir, outputDir)

	data := &TemplateData{
		Project: &ProjectData{
			Name:   "blog",
			Domain: "blog.test",
			Path:   "/home/user/web/blog",
			Port:   8080,
		},
	}

	outputPath := filepath.Join(outputDir, "nginx", "blog.conf")
	if err := g.Generate("nginx.conf.tmpl", data, outputPath); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output failed: %v", err)
	}

	expected := `server {
    listen 8080;
    server_name blog.test;
    root /home/user/web/blog;
}`
	if string(content) != expected {
		t.Errorf("unexpected output:\ngot:\n%s\nexpected:\n%s", string(content), expected)
	}
}

func TestGenerateToString(t *testing.T) {
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	os.MkdirAll(templatesDir, 0755)

	tmpl := `port: {{.Project.Port}}
domain: {{.Project.Domain}}`
	os.WriteFile(filepath.Join(templatesDir, "test.tmpl"), []byte(tmpl), 0644)

	g := NewGenerator(templatesDir, "")

	data := &TemplateData{
		Project: &ProjectData{
			Domain: "test.test",
			Port:   3000,
		},
	}

	result, err := g.GenerateToString("test.tmpl", data)
	if err != nil {
		t.Fatalf("GenerateToString failed: %v", err)
	}

	expected := "port: 3000\ndomain: test.test"
	if result != expected {
		t.Errorf("unexpected output:\ngot:\n%s\nexpected:\n%s", result, expected)
	}
}

func TestListTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	os.MkdirAll(templatesDir, 0755)

	os.WriteFile(filepath.Join(templatesDir, "a.tmpl"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(templatesDir, "b.tmpl"), []byte("b"), 0644)
	os.Mkdir(filepath.Join(templatesDir, "subdir"), 0755)

	g := NewGenerator(templatesDir, "")

	templates, err := g.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
}

func TestGenerateNonexistentTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	g := NewGenerator(tmpDir, tmpDir)

	err := g.Generate("nonexistent.tmpl", &TemplateData{}, filepath.Join(tmpDir, "out"))
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}
