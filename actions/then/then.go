package then

import (
	"github.com/byuoitav/shipwright/actions/then"
)

func init() {
	then.Add("ping-devices", PingDevices)
	then.Add("send-event", SendEvent)
}
