package ping

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/device-monitoring/localsystem"
	"go.uber.org/zap"
)

// Result .
type Result struct {
	Err string `json:"error,omitempty"`

	IP               net.IP        `json:"ip,omitempty"`
	PacketsSent      int           `json:"packets-sent,omitempty"`
	PacketsReceived  int           `json:"packets-received,omitempty"`
	PacketsLost      int           `json:"packets-lost,omitempty"`
	AverageRoundTrip time.Duration `json:"average-round-trip,omitempty"`
}

// Room pings the room and returns the results
func Room(ctx context.Context, log *zap.SugaredLogger) (map[string]*Result, *nerr.E) {
	// get devices from db
	devices, err := db.GetDB().GetDevicesByRoom(localsystem.MustRoomID())
	if err != nil {
		return map[string]*Result{}, nerr.Translate(err).Addf("unable to get devices in room %v", localsystem.MustRoomID())
	}

	hosts := []string{}
	for i := range devices {
		if len(devices[i].Address) == 0 || strings.EqualFold(devices[i].Address, "0.0.0.0") {
			continue
		}
		hosts = append(hosts, devices[i].Address)
	}

	log.Infof("Pinging %v devices in room", len(hosts))

	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pinger, err := NewPinger()
	if err != nil {
		return map[string]*Result{}, nerr.Translate(err).Addf("failed to ping devices")
	}
	defer pinger.Close()

	results := pinger.Ping(c, 3, hosts...)
	return results, nil
}

// Ping .
func (p *Pinger) Ping(ctx context.Context, count int, addrs ...string) map[string]*Result {
	// TODO Payload size?
	results := make(map[string]*Result)
	resultsMu := sync.Mutex{}
	wg := sync.WaitGroup{}

	// create a host struct for each host
	for _, addr := range addrs {
		wg.Add(1)

		// make a result struct for each addr
		ips, err := p.resolver.LookupIPAddr(ctx, addr)
		if err != nil {
			resultsMu.Lock()
			results[addr] = &Result{
				Err: fmt.Sprintf("failed to resolve ip address: %s", err),
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
			results[addr] = &Result{
				Err: fmt.Sprintf("ip ipv4 address found"),
			}
			resultsMu.Unlock()

			wg.Done()
			continue
		}

		h := &host{
			host:    addr,
			ip:      ip,
			replies: make(chan reply, 5),
		}

		p.hostsMu.Lock()
		p.hosts[ip.String()] = h
		p.hostsMu.Unlock()

		go func(hh *host) {
			result := p.ping(ctx, hh, count)

			resultsMu.Lock()
			results[hh.host] = result
			resultsMu.Unlock()

			wg.Done()
		}(h)
	}

	wg.Wait()
	return results
}
