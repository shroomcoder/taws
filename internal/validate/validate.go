package validate

import (
	"regexp"
	"strings"
)

const (
	maxDatabaseNameLen = 64
	maxPluginNameLen   = 64
	maxProjectNameLen  = 64
	maxTemplateStrLen  = 1024
)

var (
	databaseNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	snapshotIDRegex   = regexp.MustCompile(`^\d{8}-\d{6}$`)
	pluginNameRegex   = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

func DatabaseName(name string) bool {
	if len(name) > maxDatabaseNameLen {
		return false
	}
	return databaseNameRegex.MatchString(name)
}

func SnapshotID(id string) bool {
	return snapshotIDRegex.MatchString(id)
}

func PluginName(name string) bool {
	if len(name) > maxPluginNameLen {
		return false
	}
	return pluginNameRegex.MatchString(name)
}

func ProjectName(name string) bool {
	if name == "" || len(name) > maxProjectNameLen {
		return false
	}

	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}

	return !strings.HasPrefix(name, "-") && !strings.HasSuffix(name, "-")
}

func SafeTemplateString(s string) bool {
	if s == "" || len(s) > maxTemplateStrLen {
		return false
	}

	for _, c := range s {
		if c == '\n' || c == '\r' || c == '\t' {
			return false
		}
	}
	return true
}