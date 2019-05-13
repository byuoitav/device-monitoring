package handlers

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/common/v2/events"
	"github.com/labstack/echo"
)

// SendEvent injects an event into the event mesh
func SendEvent(ctx echo.Context) error {
	event := events.Event{}

	err := ctx.Bind(&event)
	if err != nil {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("unable to send event: %s", err))
	}

	// TODO fix
	// jobs.Messenger().SendEvent(event)
	return ctx.String(http.StatusOK, "success")
}
