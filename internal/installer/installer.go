package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"taws/internal/config"
	"taws/internal/ports"
)

type Environment struct {
	IsTermux    bool
	IsAndroid   bool
	HomeDir     string
	TawsRoot    string
	WebRoot     string
	PackageManager string
}

type PackageStatus struct {
	Name       string
	Installed  bool
	Version    string
	Required   bool
	CanInstall bool
}

type SetupResult struct {
	Environment   Environment
	Packages      []PackageStatus
	ConfigCreated bool
	DirsCreated   bool
	Success       bool
	Messages      []string
}

func DetectEnvironment() (*Environment, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	env := &Environment{
		IsTermux:  runtime.GOOS == "android" && os.Getenv("PREFIX") != "",
		IsAndroid: runtime.GOOS == "android",
		HomeDir:   home,
		TawsRoot:  filepath.Join(home, ".taws"),
		WebRoot:   filepath.Join(home, "web"),
	}

	if env.IsTermux {
		env.PackageManager = "pkg"
	} else {
		switch runtime.GOOS {
		case "linux":
			env.PackageManager = detectLinuxPackageManager()
		case "darwin":
			env.PackageManager = "brew"
		default:
			env.PackageManager = "unknown"
		}
	}

	return env, nil
}

func detectLinuxPackageManager() string {
	if _, err := exec.LookPath("apt"); err == nil {
		return "apt"
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return "dnf"
	}
	if _, err := exec.LookPath("pacman"); err == nil {
		return "pacman"
	}
	if _, err := exec.LookPath("zypper"); err == nil {
		return "zypper"
	}
	return "unknown"
}

func CheckPackages(env *Environment) []PackageStatus {
	packages := []PackageStatus{
		{Name: "nginx", Required: true, CanInstall: true},
		{Name: "php", Required: true, CanInstall: true},
		{Name: "mariadb", Required: true, CanInstall: true, Version: "mariadb"},
		{Name: "composer", Required: true, CanInstall: true},
		{Name: "adminer", Required: true, CanInstall: false},
	}

	for i := range packages {
		pkg := &packages[i]
		name := pkg.Name
		if pkg.Version != "" {
			name = pkg.Version
		}

		path, err := exec.LookPath(name)
		if err == nil {
			pkg.Installed = true
			pkg.Version = getPackageVersion(name)
		} else if pkg.Name == "mariadb" {
			path, err = exec.LookPath("mysql")
			if err == nil {
				pkg.Installed = true
				pkg.Version = getPackageVersion("mysql")
			}
		}
		_ = path
	}

	return packages
}

func getPackageVersion(name string) string {
	cmds := [][]string{
		{name, "--version"},
		{name, "-v"},
		{name, "version"},
	}

	for _, args := range cmds {
		out, err := exec.Command(args[0], args[1:]...).Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
		}
	}

	return "unknown"
}

func CreateDirectories(env *Environment) error {
	dirs := []string{
		env.TawsRoot,
		filepath.Join(env.TawsRoot, "configs"),
		filepath.Join(env.TawsRoot, "state"),
		filepath.Join(env.TawsRoot, "logs"),
		filepath.Join(env.TawsRoot, "backups"),
		filepath.Join(env.TawsRoot, "certs"),
		filepath.Join(env.TawsRoot, "plugins"),
		filepath.Join(env.TawsRoot, "cache"),
		env.WebRoot,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return nil
}

func RunSetup() (*SetupResult, error) {
	result := &SetupResult{}

	env, err := DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("environment detection failed: %w", err)
	}
	result.Environment = *env

	result.Messages = append(result.Messages, fmt.Sprintf("Detected environment: %s", runtime.GOOS))
	if env.IsTermux {
		result.Messages = append(result.Messages, "Termux environment detected")
	}
	result.Messages = append(result.Messages, fmt.Sprintf("Package manager: %s", env.PackageManager))
	result.Messages = append(result.Messages, fmt.Sprintf("Home directory: %s", env.HomeDir))

	result.Messages = append(result.Messages, "\nChecking required packages...")
	result.Packages = CheckPackages(env)

	allInstalled := true
	for _, pkg := range result.Packages {
		status := "NOT INSTALLED"
		if pkg.Installed {
			status = "INSTALLED"
		} else if pkg.Required {
			allInstalled = false
		}
		result.Messages = append(result.Messages, fmt.Sprintf("  %s: %s", pkg.Name, status))
	}

	result.Messages = append(result.Messages, "\nCreating TAWS directories...")
	if err := CreateDirectories(env); err != nil {
		result.Messages = append(result.Messages, fmt.Sprintf("  Error: %v", err))
		return result, err
	}
	result.DirsCreated = true
	result.Messages = append(result.Messages, "  Directories created successfully")

	result.Messages = append(result.Messages, "\nGenerating default configuration...")
	if err := generateDefaultConfig(env); err != nil {
		result.Messages = append(result.Messages, fmt.Sprintf("  Error: %v", err))
		return result, err
	}
	result.ConfigCreated = true
	result.Messages = append(result.Messages, "  Configuration generated")

	result.Messages = append(result.Messages, "\nChecking port availability...")
	configPath := filepath.Join(env.TawsRoot, "configs", "taws.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		result.Messages = append(result.Messages, fmt.Sprintf("  Warning: could not load config for port check: %v", err))
	} else {
		origPorts := map[string]int{
			"nginx":   cfg.Services.Nginx.Port,
			"php-fpm": cfg.Services.PHP.Port,
			"mariadb": cfg.Services.MariaDB.Port,
			"adminer": cfg.Services.Adminer.Port,
		}

		assigned := ports.ResolvePorts(cfg)
		hasChanges := false
		for name, newPort := range assigned {
			if newPort > 0 && origPorts[name] != newPort {
				result.Messages = append(result.Messages, fmt.Sprintf("  %s: port %d -> %d", name, origPorts[name], newPort))
				hasChanges = true
			}
		}
		if hasChanges {
			_ = config.Save(cfg, configPath)
		} else {
			result.Messages = append(result.Messages, "  All default ports available")
		}
	}

	if !allInstalled {
		result.Messages = append(result.Messages, "\nSome packages need to be installed manually:")
		for _, pkg := range result.Packages {
			if !pkg.Installed && pkg.Required && pkg.CanInstall {
				result.Messages = append(result.Messages, fmt.Sprintf("  %s install %s", env.PackageManager, pkg.Name))
			}
		}
	}

	result.Success = true
	return result, nil
}

func generateDefaultConfig(env *Environment) error {
	cfg := config.DefaultConfig()

	cfg.Paths.Home = env.HomeDir
	cfg.Paths.TawsRoot = env.TawsRoot
	cfg.Paths.WebRoot = env.WebRoot
	cfg.Paths.Configs = filepath.Join(env.TawsRoot, "configs")
	cfg.Paths.State = filepath.Join(env.TawsRoot, "state")
	cfg.Paths.Logs = filepath.Join(env.TawsRoot, "logs")
	cfg.Paths.Backups = filepath.Join(env.TawsRoot, "backups")
	cfg.Paths.Certs = filepath.Join(env.TawsRoot, "certs")
	cfg.Paths.Plugins = filepath.Join(env.TawsRoot, "plugins")

	configPath := filepath.Join(env.TawsRoot, "configs", "taws.yaml")
	return config.Save(cfg, configPath)
}

func FormatSetupResult(result *SetupResult) string {
	var sb strings.Builder

	sb.WriteString("TAWS Setup Results\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	for _, msg := range result.Messages {
		sb.WriteString(msg + "\n")
	}

	sb.WriteString("\n" + strings.Repeat("=", 50) + "\n")

	if result.Success {
		sb.WriteString("Setup completed successfully.\n")
	} else {
		sb.WriteString("Setup encountered errors.\n")
	}

	return sb.String()
}
