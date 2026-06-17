package validate

import "testing"

func TestDatabaseName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"valid_name", true},
		{"validName123", true},
		{"_private", true},
		{"name_123", true},
		{"", false},
		{"123invalid", false},
		{"name-with-dash", false},
		{"name.with.dot", false},
		{"name;drop", false},
		{"name`backtick", false},
		{"name'quote", false},
		{"name space", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DatabaseName(tt.name)
			if result != tt.expected {
				t.Errorf("DatabaseName(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestSnapshotID(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"20260101-120000", true},
		{"20261231-235959", true},
		{"", false},
		{"20260101", false},
		{"20260101-12000", false},
		{"2026-01-01-120000", false},
		{"abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SnapshotID(tt.name)
			if result != tt.expected {
				t.Errorf("SnapshotID(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestPluginName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"valid", true},
		{"valid-name", true},
		{"valid_name", true},
		{"name123", true},
		{"", false},
		{"name with space", false},
		{"name@special", false},
		{"../traversal", false},
		{"name.dll", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PluginName(tt.name)
			if result != tt.expected {
				t.Errorf("PluginName(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestProjectName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"valid", true},
		{"valid-name", true},
		{"name123", true},
		{"123name", true},
		{"", false},
		{"-invalid", false},
		{"invalid-", false},
		{"Invalid", false},
		{"has space", false},
		{"special!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProjectName(tt.name)
			if result != tt.expected {
				t.Errorf("ProjectName(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestSafeTemplateString(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"simple", true},
		{"with-dash", true},
		{"with_underscore", true},
		{"", false},
		{"with\nnewline", false},
		{"with\rcarriage", false},
		{"with\ttab", false},
		{"multiple\nlines\nhere", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeTemplateString(tt.name)
			if result != tt.expected {
				t.Errorf("SafeTemplateString(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestDatabaseNameLengthLimit(t *testing.T) {
	longName := make([]byte, maxDatabaseNameLen+1)
	for i := range longName {
		longName[i] = 'a'
	}
	if DatabaseName(string(longName)) {
		t.Error("expected false for too-long database name")
	}

	validName := make([]byte, maxDatabaseNameLen)
	for i := range validName {
		validName[i] = 'a'
	}
	if !DatabaseName(string(validName)) {
		t.Error("expected true for max-length database name")
	}
}

func TestPluginNameLengthLimit(t *testing.T) {
	longName := make([]byte, maxPluginNameLen+1)
	for i := range longName {
		longName[i] = 'a'
	}
	if PluginName(string(longName)) {
		t.Error("expected false for too-long plugin name")
	}
}

func TestProjectNameLengthLimit(t *testing.T) {
	longName := make([]byte, maxProjectNameLen+1)
	for i := range longName {
		longName[i] = 'a'
	}
	if ProjectName(string(longName)) {
		t.Error("expected false for too-long project name")
	}
}

func TestSafeTemplateStringLengthLimit(t *testing.T) {
	longStr := make([]byte, maxTemplateStrLen+1)
	for i := range longStr {
		longStr[i] = 'a'
	}
	if SafeTemplateString(string(longStr)) {
		t.Error("expected false for too-long template string")
	}
}