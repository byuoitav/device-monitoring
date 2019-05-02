package messenger

import (
	"os"
	"sync"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

var (
	once sync.Once
	m    *Messenger
)

// Get .
func Get() *Messenger {
	once.Do(func() {
		var err *nerr.E

		m, err = BuildMessenger(os.Getenv("HUB_ADDRESS"), base.Messenger, 5000)
		if err != nil {
			log.L.Fatalf("failed to build messenger: %s", err.Error())
		}
	})

	return m
}
