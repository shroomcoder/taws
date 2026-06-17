package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Paths    Paths    `json:"paths" yaml:"paths"`
	Services Services `json:"services" yaml:"services"`
}

type Paths struct {
	Home     string `json:"home" yaml:"home"`
	TawsRoot string `json:"taws_root" yaml:"taws_root"`
	WebRoot  string `json:"web_root" yaml:"web_root"`
	Configs  string `json:"configs" yaml:"configs"`
	State    string `json:"state" yaml:"state"`
	Logs     string `json:"logs" yaml:"logs"`
	Backups  string `json:"backups" yaml:"backups"`
	Certs    string `json:"certs" yaml:"certs"`
	Plugins  string `json:"plugins" yaml:"plugins"`
}

type Services struct {
	Nginx    ServiceConfig `json:"nginx" yaml:"nginx"`
	PHP      ServiceConfig `json:"php" yaml:"php"`
	MariaDB  ServiceConfig `json:"mariadb" yaml:"mariadb"`
	Composer ServiceConfig `json:"composer" yaml:"composer"`
	Adminer  ServiceConfig `json:"adminer" yaml:"adminer"`
}

type ServiceConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Port    int    `json:"port,omitempty" yaml:"port,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	tawsRoot := filepath.Join(home, ".taws")

	return &Config{
		Paths: Paths{
			Home:     home,
			TawsRoot: tawsRoot,
			WebRoot:  filepath.Join(home, "web"),
			Configs:  filepath.Join(tawsRoot, "configs"),
			State:    filepath.Join(tawsRoot, "state"),
			Logs:     filepath.Join(tawsRoot, "logs"),
			Backups:  filepath.Join(tawsRoot, "backups"),
			Certs:    filepath.Join(tawsRoot, "certs"),
			Plugins:  filepath.Join(tawsRoot, "plugins"),
		},
		Services: Services{
			Nginx:    ServiceConfig{Enabled: true, Port: 8080},
			PHP:      ServiceConfig{Enabled: true, Port: 9000},
			MariaDB:  ServiceConfig{Enabled: true, Port: 3306},
			Composer: ServiceConfig{Enabled: true},
			Adminer:  ServiceConfig{Enabled: true, Port: 8081},
		},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{}
	ext := filepath.Ext(path)

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing yaml config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing json config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format: %s", ext)
	}

	return cfg, nil
}

func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

func EnsureDirs(cfg *Config) error {
	dirs := []string{
		cfg.Paths.TawsRoot,
		cfg.Paths.Configs,
		cfg.Paths.State,
		cfg.Paths.Logs,
		cfg.Paths.Backups,
		cfg.Paths.Certs,
		cfg.Paths.Plugins,
		cfg.Paths.WebRoot,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return nil
}
