package ping

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type reply struct {
	body icmp.MessageBody
	at   time.Time
}

type host struct {
	Host
	ip      net.IP
	seq     int
	replies chan reply
}

func (p *Pinger) ping(ctx context.Context, host *host, config Config) *Result {
	result := &Result{
		IP: host.ip,
	}

	var avgrtt time.Duration

	for host.seq < config.Count {
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
			result.Error = fmt.Sprintf("failed to marshal ping message: %s", err)
			break
		}

		// write the message
		tSent := time.Now()
		n, err := p.conn.WriteTo(b, &net.IPAddr{
			IP: host.ip,
		})
		if err != nil {
			result.Error = fmt.Sprintf("failed to send ping: %s", err)
			break
		} else if n != len(b) {
			result.Error = fmt.Sprintf("sending ping failed: wrote %v bytes, expected %v", n, len(b))
			break
		}

		result.PacketsSent++

		// wait for a response
		select {
		case <-time.After(config.Delay):
			// count this as a lost packet
			slog.Info("lost packet", "seq", host.seq, "address", host.Addr)
			result.PacketsLost++
			host.seq++
		case reply := <-host.replies:
			// discard the reply if it's old
			if body, ok := reply.body.(*icmp.Echo); ok {
				if body.Seq == host.seq {
					slog.Debug("received a reply", "host", host.Addr, "time", reply.at, "seq", body.Seq)
				} else {
					slog.Debug("received a reply", "host", host.Addr, "time", reply.at, "seq", body.Seq, "expected_seq", host.seq)
				}

				host.seq++
				result.PacketsReceived++
				avgrtt += reply.at.Sub(tSent)
				time.Sleep(config.Delay)
			} else {
				slog.Warn("received a reply", "host", host.Addr, "time", reply.at, "unknown_type", fmt.Sprintf("%#v", reply.body))
			}
		case <-ctx.Done():
			result.Error = fmt.Sprintf("timed out waiting for a response from %s", host.Addr)
		}

		if len(result.Error) > 0 {
			break
		}
	}

	// calculate info in result
	if result.PacketsSent != 0 {
		avgrtt /= time.Duration(result.PacketsSent)
		if avgrtt != 0 {
			result.AverageRoundTrip = avgrtt.String()
		}
	} else {
		if result.Error == "" {
			result.Error = "no packets were sent"
		}
	}

	return result
}
