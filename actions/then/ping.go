package then

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/device-monitoring/localsystem"
	"go.uber.org/zap"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	// ICMPProtocol .
	ICMPProtocol = 1

	// ICMP6Protocol .
	ICMP6Protocol = 58
)

// Pinger .
type Pinger struct {
	resolver net.Resolver
	id       uint16
	conn     net.PacketConn

	hosts   map[string]*host
	hostsMu sync.RWMutex
}

// Result .
type Result struct {
	Err error `json:"error,omitempty"`

	IP               net.IP        `json:"ip,omitempty"`
	PacketsSent      int           `json:"packets-sent,omitempty"`
	PacketsReceived  int           `json:"packets-received,omitempty"`
	PacketsLost      int           `json:"packets-lost,omitempty"`
	AverageRoundTrip time.Duration `json:"average-round-trip,omitempty"`
}

type reply struct {
	body icmp.MessageBody
	at   time.Time
}

type host struct {
	host    string
	ip      net.IP
	seq     int
	replies chan reply

	// stats go here
}

// PingDevices .
func PingDevices(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
	// get devices from db
	devices, err := db.GetDB().GetDevicesByRoom(localsystem.MustRoomID())
	if err != nil {
		return nerr.Translate(err).Addf("unable to get devices in room %v", localsystem.MustRoomID())
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
		return nerr.Translate(err).Addf("failed to ping devices")
	}
	defer pinger.Close()

	results := pinger.Ping(c, hosts...)
	log.Infof("results: %v", results)

	return nil
}

// NewPinger .
func NewPinger() (*Pinger, error) {
	// check os permissions
	if os.Getuid() != 0 {
		return nil, fmt.Errorf("insufficient permissions to ping; must run program as root user")
	}

	p := &Pinger{
		resolver: net.Resolver{},
		hosts:    make(map[string]*host),
		id:       uint16(os.Getpid()),
	}

	return p, p.listen()
}

func (p *Pinger) listen() error {
	// start listening for icmp packets
	var err error
	p.conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("failed to bind to icmp socket: %s", err)
	}

	go func() {
		resp := make([]byte, 2048)
		for {
			n, peer, err := p.conn.ReadFrom(resp)
			if err != nil {
				if netErr, ok := err.(net.Error); !ok || !netErr.Temporary() {
					break
				}
			} else {
				p.receive(peer.(*net.IPAddr).IP, resp[:n], time.Now())
			}
		}
	}()

	return nil
}

func (p *Pinger) receive(source net.IP, bytes []byte, at time.Time) {
	// parse message
	m, err := icmp.ParseMessage(ICMPProtocol, bytes) // only send icmp (not icmp6) packets rn
	if err != nil {
		return
	}

	switch m.Type {
	case ipv4.ICMPTypeEchoReply:
		p.process(source, m.Body, at)
	case ipv4.ICMPTypeDestinationUnreachable:
		// pull out body
	}
}

func (p *Pinger) process(source net.IP, body icmp.MessageBody, at time.Time) {
	echo, ok := body.(*icmp.Echo)
	if !ok || echo == nil {
		log.L.Infof("expected *icmp.Echo, got %#v", body)
		return
	}

	if uint16(echo.ID) != p.id {
		return
	}

	p.hostsMu.RLock()
	host := p.hosts[source.String()]
	if host != nil {
		host.replies <- reply{
			body: body,
			at:   at,
		}
	}
	p.hostsMu.RUnlock()
}

func (p *Pinger) ping(ctx context.Context, host *host) *Result {
	result := &Result{}

	// format the message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(p.id),
			Seq:  host.seq,
			Data: make([]byte, 32),
		},
	}

	b, err := msg.Marshal(nil)
	if err != nil {
		result.Err = fmt.Errorf("failed to marshal ping message: %s", err)
		return result
	}

	// write the message
	n, err := p.conn.WriteTo(b, &net.IPAddr{
		IP: host.ip,
	})
	if err != nil {
		result.Err = fmt.Errorf("failed to send ping: %s", err)
		return result
	} else if n != len(b) {
		result.Err = fmt.Errorf("sending ping failed: wrote %v bytes, expected %v", n, len(b))
		return result
	}

	result.PacketsSent++

	// wait for a response
	select {
	case reply := <-host.replies:
		log.L.Infof("received a reply from %s at %s", host.host, reply.at)
		result.PacketsReceived++
	case <-ctx.Done():
		result.Err = fmt.Errorf("timed out waiting for a response from %s", host.host)
	}

	return result
}

// Ping .
func (p *Pinger) Ping(ctx context.Context, addrs ...string) map[string]*Result {
	results := make(map[string]*Result)
	resultsMu := sync.Mutex{}
	wg := sync.WaitGroup{}

	// create a host struct for each host
	for _, addr := range addrs {
		wg.Add(1)

		// make a result struct for each addr
		ips, err := p.resolver.LookupIPAddr(ctx, addr)
		if err != nil {
			results[addr] = &Result{
				Err: fmt.Errorf("failed to resolve ip address: %s", err),
			}
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
			results[addr] = &Result{
				Err: fmt.Errorf("ip ipv4 address found"),
			}
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
			result := p.ping(ctx, hh)

			resultsMu.Lock()
			results[hh.host] = result
			resultsMu.Unlock()

			wg.Done()
		}(h)
	}

	wg.Wait()
	return results
}

// Close .
func (p *Pinger) Close() {
	defer p.conn.Close()
	// any other cleanup
}
