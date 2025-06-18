package messenger

import (
	"log/slog"
	"os"
	"sync"

	"github.com/byuoitav/central-event-system/hub/base"
)

var (
	once sync.Once
	m    *Messenger
)

// Get .
func Get() *Messenger {
	once.Do(func() {
		var err error

		m, err = BuildMessenger(os.Getenv("HUB_ADDRESS"), base.Messenger, 5000)
		if err != nil {
			slog.Warn("failed to build messenger: %s", slog.String("error", err.Error()))
		}
	})

	return m
}
