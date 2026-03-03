package messenger

import (
	"net/url"
	"strings"
	"sync"

	mess "github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
)

// Messenger .
type Messenger struct {
	*mess.Messenger

	registered   []chan events.Event
	registeredMu sync.Mutex
	once         sync.Once
}

// BuildMessenger .
func BuildMessenger(hubAddress, connectionType string, bufferSize int) (*Messenger, *nerr.E) {
	msgr, err := mess.BuildMessenger(normalizeHubAddress(hubAddress), connectionType, bufferSize)

	m := &Messenger{
		Messenger: msgr,
	}

	go func() {
		for {
			event := m.ReceiveEvent()

			m.registeredMu.Lock()

			// dump the event into each channel and skip ones that are full
			for i := range m.registered {
				select {
				case m.registered[i] <- event:
				default:
				}
			}

			m.registeredMu.Unlock()
		}
	}()

	return m, err
}

func normalizeHubAddress(hubAddress string) string {
	addr := strings.TrimSpace(hubAddress)
	addr = strings.TrimSuffix(addr, "/")

	if strings.HasPrefix(addr, "ws://") || strings.HasPrefix(addr, "wss://") {
		parsed, err := url.Parse(addr)
		if err == nil && parsed.Host != "" {
			return parsed.Host
		}

		addr = strings.TrimPrefix(addr, "ws://")
		addr = strings.TrimPrefix(addr, "wss://")
	}

	return addr
}

// Register .
func (m *Messenger) Register(ch chan events.Event) {
	m.registeredMu.Lock()
	defer m.registeredMu.Unlock()

	m.registered = append(m.registered, ch)
}

// Deregister .
func (m *Messenger) Deregister(ch chan events.Event) {
	m.registeredMu.Lock()
	defer m.registeredMu.Unlock()

	i := 0
	for i = range m.registered {
		if m.registered[i] == ch {
			break
		}
	}

	m.registered[i] = m.registered[len(m.registered)-1]
	m.registered[len(m.registered)-1] = nil
	m.registered = m.registered[:len(m.registered)-1]
}
