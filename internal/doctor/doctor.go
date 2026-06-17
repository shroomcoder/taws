package doctor

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"taws/internal/config"
)

type Check struct {
	Name     string
	Status   string
	Message  string
	Critical bool
}

type Doctor struct {
	checks []Check
	cfg    *config.Config
}

func New() *Doctor {
	return &Doctor{}
}

func (d *Doctor) SetConfig(cfg *config.Config) {
	d.cfg = cfg
}

func (d *Doctor) Run() []Check {
	d.checks = []Check{}

	d.checkEnvironment()
	d.checkGo()
	d.checkGit()
	d.checkMake()
	d.checkGCC()
	d.checkTermux()
	d.checkNginx()
	d.checkPHP()
	d.checkMariaDB()
	d.checkComposer()
	d.checkStorage()
	d.checkPorts()

	return d.checks
}

func (d *Doctor) addCheck(name, status, message string, critical bool) {
	d.checks = append(d.checks, Check{
		Name:     name,
		Status:   status,
		Message:  message,
		Critical: critical,
	})
}

func (d *Doctor) checkEnvironment() {
	home, err := os.UserHomeDir()
	if err != nil {
		d.addCheck("Environment", "FAIL", "Cannot determine home directory", true)
		return
	}

	tawsRoot := filepath.Join(home, ".taws")
	if _, err := os.Stat(tawsRoot); os.IsNotExist(err) {
		d.addCheck("TAWS Directory", "WARN", "~/.taws/ does not exist (will be created on setup)", false)
	} else {
		d.addCheck("TAWS Directory", "OK", tawsRoot, false)
	}
}

func (d *Doctor) checkGo() {
	path, err := exec.LookPath("go")
	if err != nil {
		d.addCheck("Go", "MISSING", "Go is not installed or not in PATH", true)
		return
	}

	out, err := exec.Command(path, "version").Output()
	if err != nil {
		d.addCheck("Go", "FAIL", "Go installed but version check failed", true)
		return
	}

	version := strings.TrimSpace(string(out))
	d.addCheck("Go", "OK", version, true)
}

func (d *Doctor) checkGit() {
	path, err := exec.LookPath("git")
	if err != nil {
		d.addCheck("Git", "MISSING", "Git is not installed or not in PATH", false)
		return
	}

	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		d.addCheck("Git", "FAIL", "Git installed but version check failed", false)
		return
	}

	version := strings.TrimSpace(string(out))
	d.addCheck("Git", "OK", version, false)
}

func (d *Doctor) checkMake() {
	path, err := exec.LookPath("make")
	if err != nil {
		d.addCheck("Make", "MISSING", "Make is not installed or not in PATH", false)
		return
	}

	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		d.addCheck("Make", "FAIL", "Make installed but version check failed", false)
		return
	}

	lines := strings.Split(string(out), "\n")
	version := strings.TrimSpace(lines[0])
	d.addCheck("Make", "OK", version, false)
}

func (d *Doctor) checkGCC() {
	path, err := exec.LookPath("gcc")
	if err != nil {
		d.addCheck("GCC", "MISSING", "GCC is not installed or not in PATH", false)
		return
	}

	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		d.addCheck("GCC", "FAIL", "GCC installed but version check failed", false)
		return
	}

	lines := strings.Split(string(out), "\n")
	version := strings.TrimSpace(lines[0])
	d.addCheck("GCC", "OK", version, false)
}

func (d *Doctor) checkTermux() {
	if runtime.GOOS == "android" {
		termuxPrefix := os.Getenv("PREFIX")
		if termuxPrefix != "" {
			d.addCheck("Termux", "OK", "Termux environment detected", false)
		} else {
			d.addCheck("Termux", "WARN", "Running on Android but PREFIX not set", false)
		}
	} else {
		d.addCheck("Termux", "SKIP", fmt.Sprintf("Not running on Android (OS: %s)", runtime.GOOS), false)
	}
}

func (d *Doctor) checkNginx() {
	path, err := exec.LookPath("nginx")
	if err != nil {
		d.addCheck("Nginx", "NOT INSTALLED", "Nginx not found in PATH", false)
		return
	}

	out, err := exec.Command(path, "-v").Output()
	if err != nil {
		d.addCheck("Nginx", "WARN", "Nginx found but version check failed", false)
		return
	}

	version := strings.TrimSpace(string(out))
	d.addCheck("Nginx", "OK", version, false)
}

func (d *Doctor) checkPHP() {
	path, err := exec.LookPath("php")
	if err != nil {
		d.addCheck("PHP", "NOT INSTALLED", "PHP not found in PATH", false)
		return
	}

	out, err := exec.Command(path, "-v").Output()
	if err != nil {
		d.addCheck("PHP", "WARN", "PHP found but version check failed", false)
		return
	}

	lines := strings.Split(string(out), "\n")
	version := strings.TrimSpace(lines[0])
	d.addCheck("PHP", "OK", version, false)
}

func (d *Doctor) checkMariaDB() {
	path, err := exec.LookPath("mariadb")
	if err != nil {
		path, err = exec.LookPath("mysql")
		if err != nil {
			d.addCheck("MariaDB", "NOT INSTALLED", "MariaDB/MySQL not found in PATH", false)
			return
		}
	}

	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		d.addCheck("MariaDB", "WARN", "MariaDB found but version check failed", false)
		return
	}

	version := strings.TrimSpace(string(out))
	d.addCheck("MariaDB", "OK", version, false)
}

func (d *Doctor) checkComposer() {
	path, err := exec.LookPath("composer")
	if err != nil {
		d.addCheck("Composer", "NOT INSTALLED", "Composer not found in PATH", false)
		return
	}

	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		d.addCheck("Composer", "WARN", "Composer found but version check failed", false)
		return
	}

	version := strings.TrimSpace(string(out))
	d.addCheck("Composer", "OK", version, false)
}

func (d *Doctor) checkStorage() {
	home, err := os.UserHomeDir()
	if err != nil {
		d.addCheck("Storage", "FAIL", "Cannot determine home directory", true)
		return
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(home, &stat); err != nil {
		d.addCheck("Storage", "FAIL", "Cannot check disk space", false)
		return
	}

	available := stat.Bavail * uint64(stat.Bsize)
	availableGB := float64(available) / (1024 * 1024 * 1024)

	if availableGB < 1 {
		d.addCheck("Storage", "WARN", fmt.Sprintf("Low disk space: %.1f GB available", availableGB), false)
	} else {
		d.addCheck("Storage", "OK", fmt.Sprintf("%.1f GB available", availableGB), false)
	}
}

func (d *Doctor) checkPorts() {
	type portDef struct {
		name string
		port int
	}

	var portsList []portDef
	if d.cfg != nil {
		portsList = []portDef{
			{"Nginx", d.cfg.Services.Nginx.Port},
			{"PHP-FPM", d.cfg.Services.PHP.Port},
			{"MariaDB", d.cfg.Services.MariaDB.Port},
			{"Adminer", d.cfg.Services.Adminer.Port},
		}
	} else {
		portsList = []portDef{
			{"Nginx", 8080},
			{"PHP-FPM", 9000},
			{"MariaDB", 3306},
			{"Adminer", 8081},
		}
	}

	usedPorts := make(map[int]string)
	for _, p := range portsList {
		if p.port == 0 {
			continue
		}
		if existing, ok := usedPorts[p.port]; ok {
			d.addCheck(fmt.Sprintf("Port %d (%s)", p.port, p.name), "CONFLICT",
				fmt.Sprintf("Port %d is already assigned to %s", p.port, existing), true)
			continue
		}
		usedPorts[p.port] = p.name

		addr := fmt.Sprintf("127.0.0.1:%d", p.port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			d.addCheck(fmt.Sprintf("Port %d (%s)", p.port, p.name), "IN USE",
				fmt.Sprintf("Port %d is already in use", p.port), false)
		} else {
			ln.Close()
			d.addCheck(fmt.Sprintf("Port %d (%s)", p.port, p.name), "OK", "Available", false)
		}
	}
}

func FormatResults(checks []Check) string {
	var sb strings.Builder

	sb.WriteString("TAWS Doctor Results\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	critical := 0
	warnings := 0

	for _, c := range checks {
		icon := "[OK]"
		switch c.Status {
		case "FAIL", "MISSING":
			icon = "[!!]"
			critical++
		case "WARN", "NOT INSTALLED":
			icon = "[~~]"
			warnings++
		case "SKIP":
			icon = "[--]"
		case "IN USE":
			icon = "[**]"
		}

		sb.WriteString(fmt.Sprintf("  %s %-20s %s\n", icon, c.Name, c.Message))
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("  Total: %d | OK: %d | Warnings: %d | Critical: %d\n",
		len(checks), len(checks)-critical-warnings, warnings, critical))

	return sb.String()
}
