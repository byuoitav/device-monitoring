package then

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/device-monitoring/actions/activesignal"
	"github.com/byuoitav/device-monitoring/actions/ping"
	"github.com/byuoitav/device-monitoring/localsystem"
	"github.com/byuoitav/device-monitoring/messenger"
	"github.com/byuoitav/shipwright/actions/then"
	"go.uber.org/zap"
)

func init() {
	then.Add("ping-devices", pingDevices)
	then.Add("hardware-info", hardwareInfo)
	then.Add("active-signal", activeSignal)
}

func pingDevices(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
	results, err := ping.Room(ctx, log)
	if err != nil {
		return err
	}

	// TODO push up results
	log.Infof("results: %+v", results)
	return err
}

func activeSignal(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
	systemID, err := localsystem.SystemID()
	if err != nil {
		return err.Addf("unable to get active signal")
	}

	active, err := activesignal.GetMap(ctx)
	if err != nil {
		return err.Addf("unable to get active signal")
	}

	// key is deviceID, value is true/false
	for k, v := range active {
		deviceInfo := events.GenerateBasicDeviceInfo(k)

		messenger.Get().SendEvent(events.Event{
			GeneratingSystem: systemID,
			Timestamp:        time.Now(),
			EventTags: []string{
				events.DetailState,
			},
			TargetDevice: deviceInfo,
			AffectedRoom: deviceInfo.BasicRoomInfo,
			Key:          "active-signal",
			Value:        fmt.Sprintf("%v", v),
		})
	}

	return nil
}

// CheckServices .
func CheckServices(ctx context.Context, with []byte, log *zap.SugaredLogger) *nerr.E {
	/*
		var configs []health.ServiceCheckConfig

		err := then.FillStructFromTemplate(ctx, string(with), log)
		if err != nil {
		}

	*/
	return nil
}
