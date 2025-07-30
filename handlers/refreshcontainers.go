package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// RefreshContainers (Float)
func RefreshContainers(ctx *gin.Context) {
	var req *http.Request
	var err error

	// get the pi's address
	piHostname, err := os.Hostname()
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to get hostname: %v", err))
		return
	}

	newCtx, cancel := context.WithTimeout(ctx.Request.Context(), 30*time.Second)
	defer cancel()

	req, err = http.NewRequestWithContext(newCtx, http.MethodPost, fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/v2/refloat/%v", piHostname), nil)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to create refloat request: %v", err))
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to refloat: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to read body: %v response from flight-deck", err))
			return
		}

		ctx.String(http.StatusInternalServerError, fmt.Sprintf("%v response from flight-deck", string(body)))
		return
	}

	ctx.String(http.StatusOK, "Refreshed Containers")
}
