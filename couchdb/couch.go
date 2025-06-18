package couchdb

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/byuoitav/device-monitoring/model"
	"github.com/go-kivik/kivik/v4"
)

var (
	client    *kivik.Client
	clientErr error
	once      sync.Once
	dbName    string
)

func initClient() {
	addr := os.Getenv("COUCHDB_ADDR") // e.g. http://localhost:5984
	if addr == "" {
		clientErr = fmt.Errorf("COUCHDB_ADDR environment variable not set")
		return
	}

	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}

	client, clientErr = kivik.New("couch", addr)
	dbName = os.Getenv("COUCHDB_DB")
	if dbName == "" {
		dbName = "devices"
	}
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
