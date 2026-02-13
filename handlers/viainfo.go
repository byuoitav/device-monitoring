package handlers

import (
	"net/http"

	"github.com/byuoitav/device-monitoring/couchdb"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/gin-gonic/gin"
)

// ViaData .
type ViaData struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// ViaInfo .
func ViaInfo(c *gin.Context) {
	// get all of the via's out of couch
	devices, err := couchdb.GetDevicesByRoomAndType(c, localsystem.MustRoomID(), "via-connect-pro")
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to get via devices: %v", err)
		return
	}

	ret := []ViaData{}

	for i := range devices {
		data := ViaData{
			Name:    devices[i].ID,
			Address: devices[i].Address,
		}

		ret = append(ret, data)
	}

	c.JSON(http.StatusOK, ret)
}
