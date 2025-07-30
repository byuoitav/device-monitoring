package model

import (
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
)

// Device mirrors the device in the database (only the fields we care about).
type Device struct {
	ID      string            `json:"_id"`
	Address string            `json:"address"`
	Type    DeviceType        `json:"type"`
	Proxy   map[string]string `json:"proxy,omitempty"` // optional proxy settings
}

// DeviceType only needs Commands for the health check.
type DeviceType struct {
	ID       string    `json:"_id"`
	Commands []Command `json:"commands"`
}

// Command contains onlyu the ID so we can check for support.
type Command struct {
	ID           string       `json:"_id"`
	Microservice Microservice `json:"microservice"`
	Endpoint     Endpoint     `json:"endpoint"`
}

// Microservice represents the microservice that handles the command.
type Microservice struct {
	Address string `json:"address"`
}

// Endpoint represents the endpoint for the command.
type Endpoint struct {
	Path string `json:"path"`
}

// HasCommand checks if this device type supports the given command.
func (d *Device) HasCommand(commandID string) bool {
	for i := range d.Type.Commands {
		if d.Type.Commands[i].ID == commandID {
			return true
		}
	}
	return false
}

// GetRoomID extracts the “room” prefix from the device ID.
// e.g. "BLDG1‑101‑PROJ01" → "BLDG1‑101"
func (d *Device) GetRoomID() string {
	parts := strings.Split(d.ID, "-")
	if len(parts) < 2 {
		return d.ID
	}
	return parts[0] + "-" + parts[1]
}

// validate the ID format
var deviceIDRegex = regexp.MustCompile(`^[A-Za-z0-9]{2,}-[A-Za-z0-9]+-`)

func (d Device) ValidID() bool {
	return deviceIDRegex.MatchString(d.ID)
}

func (d *Device) BuildCommandURL(commandID string) (string, error) {
	var cmd *Command
	for i := range d.Type.Commands {
		if d.Type.Commands[i].ID == commandID {
			cmd = &d.Type.Commands[i]
			break
		}
	}
	if cmd == nil {
		slog.Error("command not found in device type",
			slog.String("deviceID", d.ID),
			slog.String("commandID", commandID),
			slog.Any("commands", d.Type.Commands),
		)
		return "", fmt.Errorf("command %s not found in device type %s", commandID, d.Type.Commands)
	}

	// build and parse the base URL
	raw := fmt.Sprintf("%s%s", cmd.Microservice.Address, cmd.Endpoint.Path)
	u, err := url.Parse(raw)
	if err != nil {
		slog.Error("failed to parse command URL",
			slog.String("deviceID", d.ID),
			slog.String("rawURL", raw),
			slog.Any("error", err),
		)
		return "", fmt.Errorf("invalid command URL %s: %w", raw, err)
	}

	// apply proxy override if any regex matches
	for pattern, proxy := range d.Proxy {
		rx, err := regexp.Compile(pattern)
		if err != nil {
			slog.Warn("invalid proxy regex",
				slog.String("pattern", pattern),
				slog.Any("error", err),
			)
			continue
		}
		if rx.MatchString(commandID) {
			// take host:port from proxy, but preserve original URL port if needed
			origHostParts := strings.Split(u.Host, ":")
			proxyParts := strings.Split(proxy, ":")

			var newHost string
			switch len(proxyParts) {
			case 1:
				newHost = proxyParts[0]
				if len(origHostParts) > 1 {
					newHost += ":" + origHostParts[1] // preserve original port
				}
			case 2:
				newHost = proxyParts[0] + ":" + proxyParts[1]
			default:
				slog.Warn("invalid proxy value",
					slog.String("deviceID", d.ID),
					slog.String("proxy", proxy),
				)
				continue
			}
			u.Host = newHost
			break // use the first matching proxy
		}
	}
	return u.String(), nil
}

var IDValidation = regexp.MustCompile(`([A-z,0-9]{2,}-[A-z,0-9]+)-[A-z]+[0-9]+`)

func IsDeviceIDValid(id string) bool {
	vals := IDValidation.FindStringSubmatch(id)
	return len(vals) != 0
}
