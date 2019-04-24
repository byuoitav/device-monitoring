package ping

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/byuoitav/common/log"
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

// Close .
func (p *Pinger) Close() {
	defer p.conn.Close()
	// any other cleanup
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
		body := m.Body.(*icmp.DstUnreach)
		if body == nil {
			return
		}

		hdr, err := ipv4.ParseHeader(body.Data)
		if err != nil {
			return
		}

		data := body.Data[hdr.Len:]
		msg, err := icmp.ParseMessage(ICMPProtocol, data)
		if err != nil {
			return
		}

		p.process(source, msg.Body, at)
	default:
		return
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
