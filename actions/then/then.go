package then

import (
	"context"

	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/device-monitoring/actions/ping"
	"github.com/byuoitav/shipwright/actions/then"
	"go.uber.org/zap"
)

func init() {
	then.Add("ping-devices", PingDevices)
}

// PingDevices .
func PingDevices(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
	results, err := ping.Room(ctx, log)
	if err != nil {
		return err
	}

	// push up results
	log.Infof("results: %+v", results)
	return err
}
