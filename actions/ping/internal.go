package ping

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/byuoitav/common/log"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type reply struct {
	body icmp.MessageBody
	at   time.Time
}

type host struct {
	host    string
	ip      net.IP
	seq     int
	replies chan reply
}

func (p *Pinger) ping(ctx context.Context, host *host, count int) *Result {
	result := &Result{
		IP: host.ip,
	}

	for host.seq < count {
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
			result.Err = fmt.Sprintf("failed to marshal ping message: %s", err)
			break
		}

		// write the message
		tSent := time.Now()
		n, err := p.conn.WriteTo(b, &net.IPAddr{
			IP: host.ip,
		})
		if err != nil {
			result.Err = fmt.Sprintf("failed to send ping: %s", err)
			break
		} else if n != len(b) {
			result.Err = fmt.Sprintf("sending ping failed: wrote %v bytes, expected %v", n, len(b))
			break
		}

		result.PacketsSent++

		// wait for a response
		select {
		case reply := <-host.replies:
			log.L.Infof("received a reply from %s at %s (seq: %d)", host.host, reply.at, reply.body.(*icmp.Echo).Seq)
			result.PacketsReceived++
			result.AverageRoundTrip += reply.at.Sub(tSent)
			host.seq++
		case <-ctx.Done():
			result.Err = fmt.Sprintf("timed out waiting for a response from %s", host.host)
		}

		if len(result.Err) > 0 {
			break
		}
	}

	// calculate info in result
	result.AverageRoundTrip /= time.Duration(result.PacketsSent)

	return result
}
