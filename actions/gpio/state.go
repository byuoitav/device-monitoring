package gpio

import "sync"

var (
	mu   sync.RWMutex
	pins []Pin
)

// SetPins sets the canonical pin list (copy-on-write)
func SetPins(p []Pin) {
	mu.Lock()
	defer mu.Unlock()
	pins = make([]Pin, len(p))
	copy(pins, p)
}

// GetPins returns a COPY (safe snapshot) of the current pin list
func GetPins() []Pin {
	mu.RLock()
	defer mu.RUnlock()
	cp := make([]Pin, len(pins))
	copy(cp, pins)
	return cp
}

// StartAllMonitors launches Monitor() for each pin in the canonical slice.
func StartAllMonitors() {
	mu.RLock()
	defer mu.RUnlock()
	for i := range pins {
		go (&pins[i]).Monitor()
	}
}
