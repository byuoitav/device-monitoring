package couchdb

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/byuoitav/device-monitoring/model"
	kivik "github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb" // The CouchDB driver
)

var (
	client    *kivik.Client
	clientErr error
	once      sync.Once
	dbName    string
	systemID  string
)

func initClient() {
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	address := os.Getenv("DB_ADDRESS") // should be something like "http://localhost:5984"
	dbName = os.Getenv("COUCHDB_DB")
	systemID = os.Getenv("SYSTEM_ID")

	if dbName == "" {
		dbName = "devices"
	}

	if username == "" || password == "" || address == "" {
		clientErr = fmt.Errorf("missing DB_USERNAME, DB_PASSWORD, or DB_ADDRESS")
		return
	}

	password = url.QueryEscape(password)

	// Trim possible scheme prefix from address for later parsing
	address = strings.TrimPrefix(address, "http://")
	fullURL := fmt.Sprintf("http://%s:%s@%s", username, password, address)

	// Mask password for logging
	maskedURL := fmt.Sprintf("http://%s:*****@%s", username, address)
	slog.Info("Initializing CouchDB client", slog.String("addr", maskedURL), slog.String("db", dbName))

	// Create the Kivik client
	client, clientErr = kivik.New("couch", fullURL)
}

func getClient() (*kivik.Client, error) {
	once.Do(initClient)
	return client, clientErr
}

// GetDevicesByRoom returns all devices whose room_id == roomID
func GetDevicesByRoom(ctx context.Context, roomID string) ([]model.Device, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	dbase := client.DB(dbName)
	if err := dbase.Err(); err != nil {
		return nil, err
	}

	// Use a Mango $regex selector on "_id"
	// e.g. ^BLDG1-101-
	pattern := fmt.Sprintf("^%s-", regexp.QuoteMeta(roomID))
	query := map[string]any{
		"selector": map[string]any{
			"_id": map[string]any{"$regex": pattern},
		},
	}

	rows := dbase.Find(ctx, query)
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	var out []model.Device
	for rows.Next() {
		var d model.Device
		if err := rows.ScanDoc(&d); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	// populate the device Types with their commands found in the device_types database
	for i := range out {
		if out[i].Type.ID == "" {
			slog.Warn("Device has no type ID, skipping", slog.String("device_id", out[i].ID))
			continue
		}
		typeID := out[i].Type.ID
		typeDoc := client.DB("device_types").Get(ctx, typeID)
		if typeDoc.Err() != nil {
			slog.Error("Failed to get device type", slog.String("type_id", typeID), slog.Any("error", typeDoc.Err()))
			continue
		}
		var deviceType model.DeviceType
		if err := typeDoc.ScanDoc(&deviceType); err != nil {
			slog.Error("Failed to scan device type", slog.String("type_id", typeID), slog.Any("error", err))
			continue
		}
		out[i].Type = deviceType
	}
	return out, nil
}

func GetDevicesByRoomAndType(ctx context.Context, roomID, typeID string) ([]model.Device, error) {
	// get all devices in the room
	devices, err := GetDevicesByRoom(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices by room: %w", err)
	}
	// filter by type
	var filtered []model.Device
	for _, d := range devices {
		if strings.EqualFold(d.Type.ID, typeID) {
			filtered = append(filtered, d)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no devices found in room %s with type %s", roomID, typeID)
	}
	if len(filtered) < len(devices) {
		fmt.Printf("Filtered %d devices in room %s with type %s\n", len(filtered), roomID, typeID)
	} else {
		fmt.Printf("No filtering applied, all %d devices in room %s with type %s\n", len(filtered), roomID, typeID)
	}
	return filtered, nil
}

// ValidateConnection tries to connect to CouchDB and ping the DB.
func ValidateConnection(ctx context.Context) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("couchdb client error: %w", err)
	}

	// verify connectivity to CouchDB server
	ping, err := client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("couchdb server ping failed: %w", err)
	}
	if !ping {
		return fmt.Errorf("couchdb server is not reachable")
	}
	// verify the DB is usable
	db := client.DB(dbName)
	if err := db.Err(); err != nil {
		return fmt.Errorf("couchdb db error: %w", err)
	}

	return nil
}

func GetMonitoringConfig(ctx context.Context) (map[string]any, error) {
	client, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get couchdb client: %w", err)
	}

	if systemID == "" {
		return nil, fmt.Errorf("SYSTEM_ID environment variable is not set")
	}

	db := client.DB("device-monitoring")
	if err := db.Err(); err != nil {
		return nil, fmt.Errorf("failed to open device-monitoring database: %w", err)
	}

	row := db.Get(ctx, systemID)
	if row.Err() != nil {
		return nil, fmt.Errorf("failed to get monitoring config for system %s: %w", systemID, row.Err())
	}

	var cfg map[string]any
	if err := row.ScanDoc(&cfg); err != nil {
		return nil, fmt.Errorf("failed to scan monitoring config for system %s: %w", systemID, err)
	}

	return cfg, nil
}
