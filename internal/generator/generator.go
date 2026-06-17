package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"taws/internal/validate"
)

type Generator struct {
	templatesDir string
	outputDir    string
}

type TemplateData struct {
	Config   interface{}
	Project  *ProjectData
	Services map[string]interface{}
}

type ProjectData struct {
	Name   string
	Domain string
	Path   string
	Port   int
}

func NewGenerator(templatesDir, outputDir string) *Generator {
	return &Generator{
		templatesDir: templatesDir,
		outputDir:    outputDir,
	}
}

func (g *Generator) Generate(templateName string, data *TemplateData, outputPath string) error {
	if data.Project != nil && data.Project.Name != "" {
		if !validate.ProjectName(data.Project.Name) {
			return fmt.Errorf("invalid project name in template data")
		}
		if !validate.SafeTemplateString(data.Project.Domain) {
			return fmt.Errorf("invalid domain in template data")
		}
		if !validate.SafeTemplateString(data.Project.Path) {
			return fmt.Errorf("invalid path in template data")
		}
	}

	tmplPath := filepath.Join(g.templatesDir, templateName)

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", templateName, err)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

func (g *Generator) GenerateToString(templateName string, data *TemplateData) (string, error) {
	if data.Project != nil && data.Project.Name != "" {
		if !validate.ProjectName(data.Project.Name) {
			return "", fmt.Errorf("invalid project name in template data")
		}
		if !validate.SafeTemplateString(data.Project.Domain) {
			return "", fmt.Errorf("invalid domain in template data")
		}
		if !validate.SafeTemplateString(data.Project.Path) {
			return "", fmt.Errorf("invalid path in template data")
		}
	}

	tmplPath := filepath.Join(g.templatesDir, templateName)

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", templateName, err)
	}

	var buf []byte
	writer := &byteWriter{buf: &buf}

	if err := tmpl.Execute(writer, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return string(buf), nil
}

func (g *Generator) ListTemplates() ([]string, error) {
	entries, err := os.ReadDir(g.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("reading templates directory: %w", err)
	}

	var templates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			templates = append(templates, entry.Name())
		}
	}

	return templates, nil
}

type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
