package handlers

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/byuoitav/auth/wso2"
	"github.com/gin-gonic/gin"
)

var (
	WSO2Client        wso2.Client
	errDeviceNotFound = errors.New("unable to find specified device in the database")
)

// RefreshContainers (Float)
func RefreshContainers(ctx *gin.Context) {
	ip := ctx.ClientIP()
	slog.Info("Starting float attempt", slog.String("ip", ip))

	// reverse DNS lookup
	names, err := net.LookupAddr(ip)
	if err != nil {
		msg := fmt.Sprintf("error resolving hostname for ip: %s", ip)
		slog.Error(msg, slog.Any("error", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	slog.Info("Reverse DNS result", slog.Any("names", names))

	// find a valid device hostname (ABC-123-AB2 format)
	name := ""
	deviceHostnameRegex := regexp.MustCompile("^[[:alnum:]]+-[[:alnum:]]+-[[:alnum:]]+$")
	for _, n := range names {
		short := strings.SplitN(n, ".", 2)[0]
		if deviceHostnameRegex.MatchString(short) {
			name = short
			break
		}
	}

	if name == "" {
		msg := fmt.Sprintf("no valid device hostname found for IP: %s", ip)
		slog.Error(msg)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	slog.Info("Resolved hostname", slog.String("hostname", name))

	// try prd, then stg
	// try prd, then stg
	if err := floatDevice(&WSO2Client, name, "prd"); err != nil {
		if errors.Is(err, errDeviceNotFound) {
			slog.Info("Device not found in prd, trying stg", slog.String("hostname", name))
			if err2 := floatDevice(&WSO2Client, name, "stg"); err2 != nil {
				handleFloatError(ctx, err2, name)
				return
			}
		} else {
			handleFloatError(ctx, err, name)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Successfully Floated %s", name)})
}

func floatDevice(client *wso2.Client, name, env string) error {
	url := fmt.Sprintf("https://api.byu.edu/domains/av/flight-deck/v2/refloat/%s", name)
	slog.Info("Float device request", slog.String("url", url), slog.String("env", env))

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("error building request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		if strings.Contains(string(body), "failed to get device") {
			return errDeviceNotFound
		}
		return fmt.Errorf("unknown failure from flight-deck: %s", string(body))
	}
	return nil
}

func handleFloatError(ctx *gin.Context, err error, name string) {
	if errors.Is(err, errDeviceNotFound) {
		msg := fmt.Sprintf("Device %s not found in any flight-deck environment", name)
		slog.Warn(msg)
		ctx.JSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
