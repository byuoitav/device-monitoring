package eventinfrastructure

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/fatih/color"
)

const ContextRouter = "router"
const ContextEventNode = "eventnode"

func GetIP() string {
	defer color.Unset()
	var ip net.IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err.Error()
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && strings.Contains(address.String(), "/24") {
			ip, _, err = net.ParseCIDR(address.String())
			if err != nil {
				color.Set(color.FgHiRed)
				log.Fatalf("[error] %s", err.Error())
			}
		}
	}

	if ip == nil {
		color.Set(color.FgRed)
		log.Printf("[error] failed to find an non-loopback IP Address. Using PI_HOSTNAME/DEVELOPMENT_HOSTNAME as IP.")

		devhn := os.Getenv("DEVELOPMENT_HOSTNAME")
		if len(devhn) != 0 {
			color.Set(color.FgYellow)
			log.Printf("Development machine. Using hostname %s", devhn)
			return devhn
		}

		pihn := os.Getenv("PI_HOSTNAME")
		if len(pihn) == 0 {
			color.Set(color.FgRed)
			log.Fatalf("[error] PI_HOSTNAME is not set.")
		}
		return pihn
	}

	color.Set(color.FgHiGreen)
	log.Printf("My IP address is %v", ip.String())
	return ip.String()
}
