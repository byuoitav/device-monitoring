package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/byuoitav/auth/wso2"
	"github.com/gin-gonic/gin"
)

// Per-environment config for Flight-Deck + WSO2
type FDEnvConfig struct {
	APIBase      string // e.g. https://api.byu.edu/domains/av/flight-deck/v2
	GatewayURL   string // e.g. https://api.byu.edu (WSO2 issuer/gateway base)
	ClientID     string
	ClientSecret string
	Scopes       []string // optional, if your wso2.Client supports explicit scopes
}

// Aggregate config (PRD required; STG optional fallback)
type FDConfig struct {
	PRD FDEnvConfig
	STG FDEnvConfig
}

// ===== Legacy global for backward compatibility =====
// Some other handlers in your codebase still reference this.
// server.go sets it with: handlers.WSO2Client = *wso2.New(...)
var WSO2Client wso2.Client

// Set from main() via InitFlightDeck
var fdCfg FDConfig

// One WSO2 client per environment (built on first use or in InitFlightDeck)
var wso2PRD *wso2.Client
var wso2STG *wso2.Client

var (
	errDeviceNotFound   = errors.New("unable to find specified device in the database")
	deviceHostnameRegex = regexp.MustCompile(`^[[:alnum:]]+-[[:alnum:]]+-[[:alnum:]]+$`)
)

// Optional interface to inject a custom http.Client if the library supports it
type httpClientSetter interface {
	SetHTTPClient(*http.Client)
}

// InitFlightDeck wires the global env config and (optionally) prebuilds WSO2 clients.
func InitFlightDeck(cfg FDConfig) error {
	fdCfg = cfg

	// PRD is required to be complete
	if !isComplete(cfg.PRD) {
		return fmt.Errorf("PRD Flight-Deck/WSO2 config is incomplete: need APIBase, GatewayURL, ClientID, ClientSecret")
	}
	cPRD, err := buildWSO2From(cfg.PRD)
	if err != nil {
		return fmt.Errorf("build PRD WSO2: %w", err)
	}
	wso2PRD = cPRD

	// STG is optional; if any field is set, require all
	if !isEmpty(cfg.STG) {
		if !isComplete(cfg.STG) {
			return fmt.Errorf("STG config partially set; provide all of APIBase, GatewayURL, ClientID, ClientSecret or leave all empty")
		}
		cSTG, err := buildWSO2From(cfg.STG)
		if err != nil {
			return fmt.Errorf("build STG WSO2: %w", err)
		}
		wso2STG = cSTG
	}

	return nil
}

// Helpers for config completeness
func isComplete(e FDEnvConfig) bool {
	return e.APIBase != "" && e.GatewayURL != "" && e.ClientID != "" && e.ClientSecret != ""
}
func isEmpty(e FDEnvConfig) bool {
	return e.APIBase == "" && e.GatewayURL == "" && e.ClientID == "" && e.ClientSecret == ""
}

// buildWSO2From constructs a WSO2 client using the library constructor.
// If the client type supports SetHTTPClient(*http.Client), we inject a proxy-free client.
func buildWSO2From(e FDEnvConfig) (*wso2.Client, error) {
	if e.GatewayURL == "" || e.ClientID == "" || e.ClientSecret == "" {
		return nil, fmt.Errorf("missing gateway or client credentials")
	}

	// if the last character of e.GatewayURL is not a slash, add it
	if !strings.HasSuffix(e.GatewayURL, "/") {
		e.GatewayURL += "/"
	}

	// Create via constructor (fields are unexported)
	cli := wso2.New(e.ClientID, e.ClientSecret, e.GatewayURL, "device-monitoring")

	// Prepare a hardened http.Client for the auth/gateway calls
	custom := &http.Client{
		Timeout: 12 * time.Second,
		Transport: &http.Transport{
			Proxy:                 nil, // ignore HTTP(S)_PROXY/no_proxy for these calls
			IdleConnTimeout:       30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
		},
	}
	// Inject the client iff the library exposes a setter
	if setter, ok := any(cli).(httpClientSetter); ok {
		setter.SetHTTPClient(custom)
	} else {
		slog.Debug("wso2.Client has no SetHTTPClient; using library default http.Client")
	}

	// If your wso2 library supports scopes via a method, set them here (example):
	// if scoper, ok := any(cli).(interface{ SetScopes([]string) }); ok && len(e.Scopes) > 0 {
	// 	scoper.SetScopes(e.Scopes)
	// }

	return cli, nil
}

// RefreshContainers (old "refloat"): loopback-only; PRD first (401 auto-refresh), then STG fallback.
func RefreshContainers(ctx *gin.Context) {
	ip := ctx.ClientIP()
	slog.Info("Float attempt (incoming request)", slog.String("peer_ip", ip))

	// Local-only guard
	if !isLoopback(ip) {
		msg := "refresh endpoint is local-only"
		slog.Warn(msg, slog.String("peer_ip", ip))
		ctx.JSON(http.StatusForbidden, gin.H{"error": msg})
		return
	}

	// Resolve local device name (ABC-123-AB2 style)
	name, err := selfDeviceName()
	if err != nil {
		msg := fmt.Sprintf("failed to determine local device name: %v", err)
		slog.Error(msg)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	slog.Info("Resolved local hostname", slog.String("hostname", name))

	// Try PRD first (with 401 token refresh), then STG if auth/not-found issues
	if err := floatDeviceWithRetry(name, "prd"); err != nil {
		if isAuthError(err) || errors.Is(err, errDeviceNotFound) {
			slog.Info("PRD float failed; trying STG", slog.String("hostname", name), slog.String("reason", err.Error()))
			if err2 := floatDeviceWithRetry(name, "stg"); err2 != nil {
				handleFloatError(ctx, err2, name)
				return
			}
		} else {
			handleFloatError(ctx, err, name)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Successfully Floated %s. Will take a couple minutes to complete", name)})
}

// floatDeviceWithRetry performs a refloat once, then on 401 rebuilds the client and retries once.
// It maps common statuses to meaningful errors. 2xx and 409 are treated as success.
func floatDeviceWithRetry(name, env string) error {
	var envCfg FDEnvConfig
	var cli **wso2.Client
	switch env {
	case "prd":
		envCfg = fdCfg.PRD
		cli = &wso2PRD
	case "stg":
		envCfg = fdCfg.STG
		cli = &wso2STG
	default:
		return fmt.Errorf("unknown env %q", env)
	}
	if envCfg.APIBase == "" {
		return fmt.Errorf("API base not configured for %s", env)
	}
	if *cli == nil {
		c, err := buildWSO2From(envCfg)
		if err != nil {
			return err
		}
		*cli = c
	}

	// First attempt
	status, body, err := floatOnce(*cli, envCfg.APIBase, name, env)
	if err == nil && status/100 == 2 {
		return nil
	}
	if err == nil && status == http.StatusNotFound && strings.Contains(body, "failed to get device") {
		return errDeviceNotFound
	}

	// 401 → force fresh token by rebuilding client, then retry once
	if status == http.StatusUnauthorized {
		slog.Warn("401 from Flight-Deck; refreshing token and retrying", slog.String("env", env))
		c, berr := buildWSO2From(envCfg)
		if berr != nil {
			return fmt.Errorf("token refresh failed: %v (original 401)", berr)
		}
		*cli = c
		status, body, err = floatOnce(*cli, envCfg.APIBase, name, env)
	}

	// Interpret result
	if err != nil {
		return err
	}
	if status/100 == 2 {
		return nil
	}
	switch status {
	case http.StatusUnauthorized:
		return fmt.Errorf("401 unauthorized (check client id/secret, scopes, subscription)")
	case http.StatusForbidden:
		return fmt.Errorf("403 forbidden (scope/ACL missing for %s)", env)
	case http.StatusNotFound:
		if strings.Contains(body, "failed to get device") {
			return errDeviceNotFound
		}
		return fmt.Errorf("404 from flight-deck: %s", body)
	case http.StatusTooManyRequests:
		return fmt.Errorf("429 rate limited by flight-deck: %s", body)
	case http.StatusConflict:
		// Often indicates an operation already in progress—treat as OK
		slog.Warn("409 conflict (already floating?)", slog.String("env", env), slog.String("body", body))
		return nil
	default:
		return fmt.Errorf("flight-deck error %d: %s", status, body)
	}
}

// floatOnce issues POST {API_BASE}/refloat/{hostname} with a short, bounded context (not tied to the client).
// It returns (status, smallBody, err) and logs WWW-Authenticate on non-2xx for diagnostics.
func floatOnce(client *wso2.Client, apiBase, name, env string) (int, string, error) {
	url := strings.TrimRight(apiBase, "/") + "/refloat/" + name
	slog.Info("Float device request", slog.String("url", url), slog.String("env", env))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("error building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(res.Body, 4<<10)) // cap to 4KB in logs
	body := string(bodyBytes)

	if res.StatusCode/100 != 2 {
		if www := res.Header.Get("WWW-Authenticate"); www != "" {
			slog.Warn("non-2xx with WWW-Authenticate", slog.Int("status", res.StatusCode), slog.String("www_authenticate", www))
		}
		slog.Warn("flight-deck non-2xx", slog.Int("status", res.StatusCode), slog.String("body", body))
	}
	return res.StatusCode, body, nil
}

// isAuthError returns true when the error indicates an authorization/permission problem.
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "401") ||
		strings.Contains(msg, "unauthor") ||
		strings.Contains(msg, "403") ||
		strings.Contains(msg, "forbidden")
}

// handleFloatError shapes the HTTP response for the caller and logs appropriately.
func handleFloatError(ctx *gin.Context, err error, name string) {
	switch {
	case errors.Is(err, errDeviceNotFound):
		msg := fmt.Sprintf("Device %s not found in any flight-deck environment", name)
		slog.Warn(msg)
		ctx.JSON(http.StatusNotFound, gin.H{"error": msg})
	case strings.Contains(err.Error(), "401"):
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case strings.Contains(err.Error(), "403"):
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case strings.Contains(err.Error(), "429"):
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
	default:
		slog.Error("float failed", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// isLoopback restricts access to local callers.
func isLoopback(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && ip.IsLoopback()
}

// selfDeviceName returns the local device hostname in ABC-123-AB2 format.
// It uppercases the host to satisfy the regex if the OS returns lowercase.
func selfDeviceName() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("error getting hostname: %w", err)
	}
	host = strings.ToUpper(host)
	if !deviceHostnameRegex.MatchString(host) {
		return "", fmt.Errorf("invalid hostname format: %s", host)
	}
	return host, nil
}
