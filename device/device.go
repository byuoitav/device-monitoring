package device

import (
	"bytes"
	"errors"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/byuoitav/av-api/dbo"
)

func GetAddresses(building string, room string) (map[string]net.IP, error) {

	log.Printf("Getting room IP Addresses...")

	output := make(map[string]net.IP)

	devices, err := dbo.GetDevicesByRoom(building, room)
	if err != nil {
		log.Printf("Error getting devices from room: %s", err.Error())
		message := "Error getting devices from room: " + err.Error()
		return map[string]net.IP{}, errors.New(message)
	}

	for _, device := range devices {

		address := net.ParseIP(device.Address)

		if address == nil {

			log.Printf("Error parsing IP address of device: %s", device.Name)

		} else {

			log.Printf("Discovered address: %v", address)
			output[device.Name] = address

		}
	}

	return output, nil

}

func CheckDevices(addresses map[string]net.IP) error {

	log.Printf("Checking devices on network: %s", "monsters")

	for device, address := range addresses {

		names, err := net.Dial("tcp", address.String())
		if err != nil {

			log.Printf("Error dialing address: %s: %s", address.String(), err.Error())
			//publish event
			log.Printf("Firing event")

		} else {

			log.Printf("%s maps to %v", address.String(), names)
			log.Printf("Device: %s is responding normally", device)

		}

	}

	return nil
}

func ScanNetwork() (bool, error) {

	log.Printf("Scanning newtwork: %s", "strawberries")

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		message := "Problem scanning interface addresses: " + err.Error()
		return false, errors.New(message)
	}

	log.Printf("Addresses found: %v", addresses)

	for _, address := range addresses {

		ip, _, err := net.ParseCIDR(address.String())
		if err != nil {
			log.Printf(err.Error())
		}

		result, err := net.LookupAddr(ip.String())
		if err != nil {
			log.Printf("Problem looking up address: %v", err.Error())
			continue
		}

		log.Printf("%v maps to %v", address.String, result)

	}

	return true, nil
}

func Bash(addresses map[string]net.IP) error {

	log.Printf("Bashing Hacking")

	for device, address := range addresses {

		if address.String() == "0.0.0.0" {
			continue
		}

		log.Printf("Building command...")
		cmd := exec.Command("ping", address.String(), "-c 5", "-i 3")
		log.Printf("cmd: %v", cmd)

		cmd.Stdin = strings.NewReader(address.String())

		var out bytes.Buffer
		cmd.Stdout = &out

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		log.Printf("Running command...")
		err := cmd.Run()

		if err != nil && strings.Contains(stderr.String(), "0 packets received") {
			log.Print("Alert! No response from device %s at address %s", device, address.String())
		} else if strings.Contains(out.String(), "0.0%") {
			log.Printf("Device %s at address %s responding normally")
		} else {
			log.Printf("Houston, we have a problem")
		}

	}

	log.Printf("Done")

	return nil
}
