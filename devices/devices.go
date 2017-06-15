package devices

import (
	"log"
	"net"
	"time"

	ping "github.com/tatsushid/go-fastping"
)

func PingDevices() bool {

	//build IP address
	address, err := net.ResolveIPAddr("ip", "10.66.9.7")
	if err != nil {
		log.Printf("This is a perfect time to panic! %s", err.Error())
	} else {
		log.Printf("Successfully resolved IP address: %v", address)
	}

	pinger := ping.NewPinger()

	pinger.AddIPAddr(address)

	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		log.Printf("Impressive... most impressive.")
		log.Printf("Received heartbeat from %v", *addr)
	}

	pinger.OnIdle = func() {
		log.Printf("Houston, we have a problem.")
	}

	err = pinger.Run()
	if err != nil {
		log.Printf("Nuclear launch detected. Duck and cover")
	}

	return true
}
