package roomstate

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

// GetRoomState .
func GetRoomState(ctx context.Context, roomID string) (base.PublicRoom, *nerr.E) {
	log.L.Infof("Getting the status of %s", roomID)
	var state base.PublicRoom

	// build the request
	split := strings.Split(roomID, "-")
	if len(split) != 2 {
		return state, nerr.Createf("error", "failed to get room state - invalid room id %s", roomID)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/buildings/%v/rooms/%v", split[0], split[1]), nil)
	if err != nil {
		return state, nerr.Translate(err).Addf("failed to get room state")
	}

	// do the request
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return state, nerr.Translate(err).Addf("failed to get room state")
	}
	defer resp.Body.Close()

	// read the body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return state, nerr.Translate(err).Addf("failed to read API status response: %v", err)
	}

	if resp.StatusCode/100 != 2 {
		return state, nerr.Createf("error", "failed to get room state - %v response received from API. body: %s", resp.StatusCode, b)
	}

	err = json.Unmarshal(b, &state)
	if err != nil {
		return state, nerr.Translate(err).Addf("failed to unmarshal response")
	}

	return state, nil
}
