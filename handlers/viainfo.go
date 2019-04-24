package handlers

/*
// ViaData .
type ViaData struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// ViaInfo .
func ViaInfo(ctx echo.Context) error {
	// get all of the via's out of couch
	devices, err := db.GetDB().GetDevicesByRoomAndType(localsystem.MustRoomID(), "via-connect-pro")
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	ret := []ViaData{}

	for i := range devices {
		data := ViaData{
			Name:    devices[i].ID,
			Address: devices[i].Address,
		}

		ret = append(ret, data)
	}

	return ctx.JSON(http.StatusOK, ret)
}
*/
