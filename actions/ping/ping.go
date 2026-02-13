package ping

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/device-monitoring/couchdb"
	"github.com/byuoitav/device-monitoring/localsystem"
)

// Config .
type Config struct {
	Count int           // the number of pings to send
	Delay time.Duration // the delay after sending a ping before sending the next
}

// Host .
type Host struct {
	ID   string
	Addr string
}

// Result .
type Result struct {
	Error string `json:"error,omitempty"`

	IP               net.IP `json:"ip,omitempty"`
	PacketsSent      int    `json:"packets-sent,omitempty"`
	PacketsReceived  int    `json:"packets-received,omitempty"`
	PacketsLost      int    `json:"packets-lost,omitempty"`
	AverageRoundTrip string `json:"average-round-trip,omitempty"`
}

// Room pings the room and returns the results
func Room(
	ctx context.Context,
	roomID string,
	config Config,
	logger *slog.Logger,
) (map[string]*Result, error) {
	// get devices from db
	devices, err := couchdb.GetDevicesByRoom(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("unable to list devices in room %q: %w", localsystem.MustRoomID(), err)
	}

	// build the host list, skipping devices with no address
	hosts := make([]Host, 0, len(devices))
	for _, d := range devices {
		if d.Address == "" || strings.EqualFold(d.Address, "0.0.0.0") {
			continue
		}
		hosts = append(hosts, Host{ID: d.ID, Addr: d.Address})
	}

	logger.Info("Pinging devices in room",
		slog.String("room_id", roomID),
		slog.Int("host_count", len(hosts)),
	)

	pinger, err := NewPinger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pinger: %w", err)
	}
	defer pinger.Close()

	results := pinger.Ping(ctx, config, hosts...)
	return results, nil
}

// Ping .
func (p *Pinger) Ping(ctx context.Context, config Config, hosts ...Host) map[string]*Result {
	// TODO Payload size?
	results := make(map[string]*Result)
	resultsMu := sync.Mutex{}
	wg := sync.WaitGroup{}

	// create a host struct for each host
	for i := range hosts {
		wg.Add(1)

		// make a result struct for each addr
		ips, err := p.resolver.LookupIPAddr(ctx, hosts[i].Addr)
		if err != nil {
			resultsMu.Lock()
			results[hosts[i].ID] = &Result{
				Error: fmt.Sprintf("failed to resolve ip address: %s", err),
			}
			resultsMu.Unlock()

			wg.Done()
			continue
		}

		var ip net.IP
		for _, i := range ips {
			if ip = i.IP.To4(); ip != nil {
				break
			}
		}

		if ip == nil {
			resultsMu.Lock()
			results[hosts[i].ID] = &Result{
				Error: "no ipv4 address found",
			}
			resultsMu.Unlock()

			wg.Done()
			continue
		}

		h := &host{
			Host: Host{
				ID:   hosts[i].ID,
				Addr: hosts[i].Addr,
			},
			ip:      ip,
			replies: make(chan reply, 10),
		}

		p.hostsMu.Lock()
		p.hosts[ip.String()] = h
		p.hostsMu.Unlock()

		go func(hh *host) {
			result := p.ping(ctx, hh, config)

			resultsMu.Lock()
			results[hh.ID] = result
			resultsMu.Unlock()

			wg.Done()
		}(h)
	}

	wg.Wait()
	return results
}
