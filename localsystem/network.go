package localsystem

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"strings"
)

const staticConfigPath = "/etc/network/static_config.json"

type StaticConfig struct {
	Addresses string `json:"addresses"`
	Gateway   string `json:"gateway"`
	DNS       string `json:"dns"`
}

// Hostname returns the hostname of the device
func Hostname() (string, error) {
	h, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	return h, nil
}

// MustHostname returns the hostname of the device, and panics if it fails
func MustHostname() string {
	h, err := Hostname()
	if err != nil {
		log.Fatalf("failed to get hostname: %v", err)
	}
	return h
}

// IPAddress gets the public ip address of the device
func IPAddress() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to list network interfaces: %w", err)
	}

	var ip net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			strings.Contains(addr.String(), "/24") {

			parsed, _, perr := net.ParseCIDR(addr.String())
			if perr != nil {
				return nil, fmt.Errorf("failed to parse IP %q: %w", addr.String(), perr)
			}
			ip = parsed
			break
		}
	}

	if ip == nil {
		return nil, fmt.Errorf("no non‑loopback /24 IP address found")
	}

	slog.Info("My IP address", slog.String("ip", ip.String()))
	return ip, nil
}

// IsConnectedToInternet returns true if the device can reach google's servers.
func IsConnectedToInternet() bool {
	conn, err := net.Dial("tcp", "google.com:80")
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// UsingDHCP returns true if the device is using DHCP, and false if it has a static ip set.
func UsingDHCP() (bool, error) {

	// Get the active connection name
	connName, err := getActiveConnectionName()
	if err != nil {
		return false, fmt.Errorf("failed to get active connection name: %w", err)
	}

	// use nmcli to check the ipv4 method of the connection
	cmd := exec.Command("nmcli", "-t", "-f", "ipv4.method", "connection", "show", connName)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to run nmcli command: %w", err)
	}

	method := strings.TrimSpace(out.String())
	switch method {
	case "auto":
		return true, nil
	case "manual", "disabled": // manual means static IP, disabled means no IP
		return false, nil
	default:
		return false, fmt.Errorf("unknown ipv4.method value: %s", method)
	}
}

// ToggleDHCP switches between DHCP and static mode. (using nmcli)
func ToggleDHCP() error {
	connName, err := getActiveConnectionName()
	if err != nil {
		return err
	}

	enabled, err := UsingDHCP()
	if err != nil {
		return fmt.Errorf("failed to determine current DHCP state: %w", err)
	}

	if enabled {
		// DHCP → Static (restore from saved config)
		cfg, err := loadStaticConfig()
		if err != nil {
			return fmt.Errorf("cannot restore static config: %w", err)
		}

		cmd := exec.Command("nmcli", "connection", "modify", connName,
			"ipv4.method", "manual",
			"ipv4.addresses", cfg.Addresses,
			"ipv4.gateway", cfg.Gateway,
			"ipv4.dns", cfg.DNS)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to apply static config: %w — %s", err, strings.TrimSpace(string(out)))
		}
	} else {
		// Static → DHCP (save current static config)
		cfg, err := readCurrentStaticConfig(connName)
		if err != nil {
			return fmt.Errorf("cannot save static config: %w", err)
		}
		if err := saveStaticConfig(cfg); err != nil {
			return err
		}

		cmd := exec.Command("nmcli", "connection", "modify", connName,
			"ipv4.method", "auto",
			"ipv4.addresses", "",
			"ipv4.gateway", "",
			"ipv4.dns", "")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to switch to DHCP: %w — %s", err, strings.TrimSpace(string(out)))
		}
	}

	// Apply changes
	if out, err := exec.Command("nmcli", "connection", "up", connName).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to activate connection: %w — %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// CanToggleDHCP returns nil if you can toggle DHCP, or an error if you can't
func CanToggleDHCP() error {
	connName, err := getActiveConnectionName()
	if err != nil {
		return fmt.Errorf("failed to get active connection name: %w", err)
	}

	// Check if the connection is modifiable (not locked)
	cmd := exec.Command("nmcli", "-t", "-f", "connection.permissions", "connection", "show", connName)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run nmcli command: %w", err)
	}

	permissions := strings.TrimSpace(out.String())
	if permissions != "" && permissions != "--" {
		return fmt.Errorf("connection %s is not modifiable, permissions: %s", connName, permissions)
	}

	return nil
}

func getActiveConnectionName() (string, error) {
	cmd := exec.Command("nmcli", "-t", "-f", "NAME", "connection", "show", "--active")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run nmcli command: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", fmt.Errorf("no active network connection found")
	}

	return lines[0], nil
}

func saveStaticConfig(cfg StaticConfig) error {
	f, err := os.Create(staticConfigPath)
	if err != nil {
		return fmt.Errorf("failed to save static config: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(cfg)
}

func loadStaticConfig() (StaticConfig, error) {
	var cfg StaticConfig
	f, err := os.Open(staticConfigPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to load static config: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("invalid static config file: %w", err)
	}
	return cfg, nil
}

func readCurrentStaticConfig(connName string) (StaticConfig, error) {
	cmd := exec.Command("nmcli", "-t", "-f", "ipv4.addresses,ipv4.gateway,ipv4.dns", "connection", "show", connName)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return StaticConfig{}, fmt.Errorf("failed to read current static config: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), ":")
	if len(lines) < 3 {
		return StaticConfig{}, fmt.Errorf("unexpected nmcli output: %s", out.String())
	}

	return StaticConfig{
		Addresses: lines[0],
		Gateway:   lines[1],
		DNS:       lines[2],
	}, nil
}
