package roomstate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/base"
)

// Get .
func Get(ctx context.Context, roomID string) (base.PublicRoom, error) {
	slog.Info("Getting room state %s", slog.String("roomID", roomID))
	var state base.PublicRoom

	// build the request
	split := strings.Split(roomID, "-")
	if len(split) != 2 {
		slog.Error("Failed to get room state - invalid room id",
			slog.String("roomID", roomID),
			slog.String("error", "invalid room id format, expected 'buildingID-roomID'"),
		)
		return state, fmt.Errorf("failed to get room state - invalid room id %s", roomID)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/buildings/%v/rooms/%v", split[0], split[1]), nil)
	if err != nil {
		slog.Error("Failed to create request to get room state",
			slog.String("roomID", roomID),
			slog.String("error", err.Error()),
		)
		return state, fmt.Errorf("failed to get room state - %v", err)
	}

	// do the request
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to get room state",
			slog.String("roomID", roomID),
			slog.String("error", err.Error()),
		)
		return state, fmt.Errorf("failed to get room state - %v", err)
	}
	defer resp.Body.Close()

	// read the body
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read AV-API response",
			slog.String("roomID", roomID),
			slog.Int("status_code", resp.StatusCode),
			slog.String("response_body", string(b)),
		)
		return state, fmt.Errorf("failed to read AV-API response - %v", err)
	}

	if resp.StatusCode/100 != 2 {
		slog.Error("Failed to get room state",
			slog.String("roomID", roomID),
			slog.Int("status_code", resp.StatusCode),
			slog.String("response_body", string(b)),
		)
		return state, fmt.Errorf("failed to get room state - %v response received from API. body: %s", resp.StatusCode, b)
	}

	err = json.Unmarshal(b, &state)
	if err != nil {
		slog.Error("Failed to unmarshal room state",
			slog.String("roomID", roomID),
			slog.String("response_body", string(b)),
		)
		return state, fmt.Errorf("failed to unmarshal room state - %v", err)
	}

	return state, nil
}
