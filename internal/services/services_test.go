package services

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestServiceConstants(t *testing.T) {
	if Nginx != "nginx" {
		t.Errorf("expected Nginx to be 'nginx', got %s", Nginx)
	}

	if PHP != "php-fpm" {
		t.Errorf("expected PHP to be 'php-fpm', got %s", PHP)
	}

	if MariaDB != "mariadb" {
		t.Errorf("expected MariaDB to be 'mariadb', got %s", MariaDB)
	}

	if Adminer != "adminer" {
		t.Errorf("expected Adminer to be 'adminer', got %s", Adminer)
	}
}

func TestExtractPID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "systemd format",
			input:    "Main PID: 1234 (nginx)",
			expected: "1234",
		},
		{
			name:     "sysvinit format",
			input:    "nginx is running (pid 1234).",
			expected: "1234",
		},
		{
			name:     "no pid",
			input:    "nginx is stopped",
			expected: "",
		},
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPID(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	statuses := []ServiceStatus{
		{Name: Nginx, Running: true, PID: "1234"},
		{Name: PHP, Running: false},
		{Name: MariaDB, Running: true},
	}

	output := FormatStatus(statuses)

	if output == "" {
		t.Fatal("FormatStatus returned empty string")
	}

	if len(output) < 30 {
		t.Error("FormatStatus output too short")
	}
}

func TestServiceStatusStruct(t *testing.T) {
	status := ServiceStatus{
		Name:    Nginx,
		Running: true,
		PID:     "1234",
	}

	if status.Name != Nginx {
		t.Errorf("expected name %s, got %s", Nginx, status.Name)
	}

	if !status.Running {
		t.Error("expected Running to be true")
	}

	if status.PID != "1234" {
		t.Errorf("expected PID 1234, got %s", status.PID)
	}
}
