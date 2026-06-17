package ports

import (
	"fmt"
	"net"

	"taws/internal/config"
)

const (
	MinPort     = 1
	MaxPort     = 65535
	MaxAttempts = 100
)

// IsAvailable checks whether a TCP port on 127.0.0.1 is free.
func IsAvailable(port int) bool {
	if port < MinPort || port > MaxPort {
		return false
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// FindAvailable returns the first free TCP port starting from base.
// It skips any port present in skip (e.g. ports already assigned to other services).
// It scans up to maxAttempts consecutive ports, returning an error if none are available.
func FindAvailable(base, maxAttempts int, skip map[int]bool) (int, error) {
	for i := 0; i < maxAttempts; i++ {
		port := base + i
		if port > MaxPort {
			break
		}
		if skip != nil && skip[port] {
			continue
		}
		if IsAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found starting from %d after %d attempts", base, maxAttempts)
}

// ResolvePorts checks every configured service port for availability.
// If a port is in use it finds the next free port and updates cfg in place.
// Returns a map of service name → assigned port (only included if the port differs from original).
func ResolvePorts(cfg *config.Config) map[string]int {
	type servicePort struct {
		name string
		port *int
	}

	services := []servicePort{
		{"nginx", &cfg.Services.Nginx.Port},
		{"php-fpm", &cfg.Services.PHP.Port},
		{"mariadb", &cfg.Services.MariaDB.Port},
		{"adminer", &cfg.Services.Adminer.Port},
	}

	assigned := make(map[string]int)
	usedPorts := make(map[int]bool)

	for _, svc := range services {
		if *svc.port == 0 {
			continue
		}
		if IsAvailable(*svc.port) && !usedPorts[*svc.port] {
			usedPorts[*svc.port] = true
			assigned[svc.name] = *svc.port
			continue
		}
		newPort, err := FindAvailable(*svc.port+1, MaxAttempts, usedPorts)
		if err != nil {
			assigned[svc.name] = -1
			continue
		}
		usedPorts[newPort] = true
		*svc.port = newPort
		assigned[svc.name] = newPort
	}

	return assigned
}
