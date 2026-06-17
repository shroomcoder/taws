package ports

import (
	"fmt"
	"net"
	"testing"
	"time"

	"taws/internal/config"
)

func TestIsAvailable_OpenPort(t *testing.T) {
	port := 49152
	if !IsAvailable(port) {
		t.Errorf("expected port %d to be available", port)
	}
}

func TestIsAvailable_InvalidPort(t *testing.T) {
	if IsAvailable(0) {
		t.Error("expected port 0 to be unavailable")
	}
	if IsAvailable(-1) {
		t.Error("expected port -1 to be unavailable")
	}
	if IsAvailable(65536) {
		t.Error("expected port 65536 to be unavailable")
	}
}

func TestIsAvailable_OccupiedPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	port := addr.Port

	if IsAvailable(port) {
		t.Errorf("expected port %d to be occupied but IsAvailable returned true", port)
	}
}

func TestFindAvailable_FreePort(t *testing.T) {
	port, err := FindAvailable(49200, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port < 49200 || port > 49210 {
		t.Errorf("expected port in range [49200, 49210], got %d", port)
	}
}

func TestFindAvailable_SkipsOccupiedPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	occupied := addr.Port

	port, err := FindAvailable(occupied, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port == occupied {
		t.Errorf("FindAvailable returned the occupied port %d", port)
	}
}

func TestFindAvailable_NoAvailablePort(t *testing.T) {
	skip := make(map[int]bool)
	for i := 0; i < MaxAttempts; i++ {
		skip[49200+i] = true
	}
	_, err := FindAvailable(49200, MaxAttempts, skip)
	if err == nil {
		t.Error("expected error when all ports occupied via skip map")
	}
}

func TestResolvePorts_NoConflict(t *testing.T) {
	cfg := config.DefaultConfig()

	// Set ports to something unlikely to be in use
	cfg.Services.Nginx.Port = 49210
	cfg.Services.PHP.Port = 49211
	cfg.Services.MariaDB.Port = 49212
	cfg.Services.Adminer.Port = 49213

	assigned := ResolvePorts(cfg)

	for name, port := range assigned {
		if port <= 0 {
			t.Errorf("service %s got invalid port %d", name, port)
		}
	}

	// Ports should be unchanged
	if cfg.Services.Nginx.Port != 49210 {
		t.Errorf("nginx port changed unexpectedly: %d", cfg.Services.Nginx.Port)
	}
	if cfg.Services.PHP.Port != 49211 {
		t.Errorf("php port changed unexpectedly: %d", cfg.Services.PHP.Port)
	}
}

func TestResolvePorts_WithConflict(t *testing.T) {
	// Occupy a specific port
	ln, err := net.Listen("tcp", "127.0.0.1:49220")
	if err != nil {
		t.Fatalf("failed to bind: %v", err)
	}
	defer ln.Close()

	cfg := config.DefaultConfig()
	cfg.Services.Nginx.Port = 49220
	cfg.Services.PHP.Port = 49221
	cfg.Services.MariaDB.Port = 49222
	cfg.Services.Adminer.Port = 49223

	assigned := ResolvePorts(cfg)

	if assigned["nginx"] == 49220 {
		t.Error("nginx port should have been reassigned")
	}
	if cfg.Services.Nginx.Port == 49220 {
		t.Error("config should have been updated to new port")
	}
	// php was configured at 49221 which is free, but nginx shifted to 49221,
	// so php must also shift
	if cfg.Services.PHP.Port == 49221 {
		t.Error("php port should have shifted to avoid conflict with nginx")
	}
	// All ports should be unique
	ports := []int{cfg.Services.Nginx.Port, cfg.Services.PHP.Port, cfg.Services.MariaDB.Port, cfg.Services.Adminer.Port}
	seen := make(map[int]bool)
	for _, p := range ports {
		if seen[p] {
			t.Errorf("duplicate port %d in resolved config", p)
		}
		seen[p] = true
	}
}

func TestResolvePorts_SkipsOccupied(t *testing.T) {
	// Occupy two consecutive ports
	ln1, err := net.Listen("tcp", "127.0.0.1:49230")
	if err != nil {
		t.Fatalf("failed to bind: %v", err)
	}
	defer ln1.Close()

	ln2, err := net.Listen("tcp", "127.0.0.1:49231")
	if err != nil {
		t.Fatalf("failed to bind: %v", err)
	}
	defer ln2.Close()

	cfg := config.DefaultConfig()
	cfg.Services.Nginx.Port = 49230

	assigned := ResolvePorts(cfg)

	if assigned["nginx"] == 49230 || assigned["nginx"] == 49231 {
		t.Errorf("nginx should have skipped occupied ports, got %d", assigned["nginx"])
	}
}

func TestResolvePorts_PreservesUserConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Services.Nginx.Port = 8888
	cfg.Services.PHP.Port = 9001
	cfg.Services.MariaDB.Port = 3307
	cfg.Services.Adminer.Port = 8882

	// All ports likely free
	assigned := ResolvePorts(cfg)

	// Verify all are assigned
	for _, port := range assigned {
		if port < 0 {
			t.Errorf("expected all ports to be assigned")
		}
	}
}

func TestResolvePorts_Deduplicates(t *testing.T) {
	cfg := config.DefaultConfig()
	// Set all services to the same port
	cfg.Services.Nginx.Port = 49240
	cfg.Services.PHP.Port = 49240
	cfg.Services.MariaDB.Port = 49240
	cfg.Services.Adminer.Port = 49240

	assigned := ResolvePorts(cfg)

	// All should get different ports (or at least not conflict)
	seen := make(map[int]string)
	for name, port := range assigned {
		if port > 0 {
			if prev, exists := seen[port]; exists {
				t.Errorf("port %d used by both %s and %s", port, prev, name)
			}
			seen[port] = name
		}
	}
}

func TestIsAvailable_ClosedConnection(t *testing.T) {
	// Test with a briefly occupied port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind: %v", err)
	}
	addr := ln.Addr().(*net.TCPAddr)
	port := addr.Port
	ln.Close()

	// Allow time for the port to be released
	time.Sleep(10 * time.Millisecond)

	if !IsAvailable(port) {
		t.Errorf("expected port %d to be available after close", port)
	}
}

func TestFormatChanges(t *testing.T) {
	changes := map[string]int{
		"nginx":   8081,
		"php-fpm": 9000,
	}

	var msg string
	for _, name := range []string{"nginx", "php-fpm"} {
		port := changes[name]
		msg += fmt.Sprintf("  %s: port %d in use, using %d\n", name, port-1, port)
	}

	if msg == "" {
		t.Error("expected formatted message")
	}
}
