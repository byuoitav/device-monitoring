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
		slog.Info("Messenger initialized", slog.String("address", os.Getenv("HUB_ADDRESS")), slog.String("type", base.Messenger))
		slog.Error("Error initializing messenger", slog.Any("error", err))
	})

	return m
}
