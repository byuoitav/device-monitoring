package handlers

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/byuoitav/auth/wso2"
	"github.com/gin-gonic/gin"
)

var (
	WSO2Client        wso2.Client
	errDeviceNotFound = errors.New("unable to find specified device in the database")
	// ABC-123-AB2 style
	deviceHostnameRegex = regexp.MustCompile(`^[[:alnum:]]+-[[:alnum:]]+-[[:alnum:]]+$`)
)

// RefreshContainers (Float)
func RefreshContainers(ctx *gin.Context) {
	ip := ctx.ClientIP()
	slog.Info("Float attempt (incoming request)", slog.String("peer_ip", ip))

	// Require local caller (loopback only)
	if !isLoopback(ip) {
		msg := "refresh endpoint is local-only"
		slog.Warn(msg, slog.String("peer_ip", ip))
		ctx.JSON(http.StatusForbidden, gin.H{"error": msg})
		return
	}

	// Derive the device name from the local machine
	name, err := selfDeviceName()
	if err != nil {
		msg := fmt.Sprintf("failed to determine local device name: %v", err)
		slog.Error(msg)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	slog.Info("Resolved local hostname", slog.String("hostname", name))

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

func isLoopback(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && ip.IsLoopback()
}

// selfDeviceName returns the local device hostname in ABC-123-AB2 format
func selfDeviceName() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("error getting hostname: %w", err)
	}

	// Validate and format the hostname
	if !deviceHostnameRegex.MatchString(host) {
		return "", fmt.Errorf("invalid hostname format: %s", host)
	}
	return host, nil
}
