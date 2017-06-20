package device

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"strings"

	"github.com/byuoitav/av-api/dbo"
)

func GetAddresses(building string, room string) (map[string]string, error) {

	log.Printf("Getting room IP Addresses...")

	output := make(map[string]string)

	devices, err := dbo.GetDevicesByRoom(building, room)
	if err != nil {
		log.Printf("Error getting devices from room: %s", err.Error())
		message := "Error getting devices from room: " + err.Error()
		return map[string]string{}, errors.New(message)
	}

	for _, device := range devices {

		log.Printf("Discovered address: %v", device.Address)
		output[device.Name] = device.Address

	}

	return output, nil

}

func PingAddresses(addresses map[string]string) error {

	log.Printf("Pinging with bash commands...")

	for device, address := range addresses {

		if address == "0.0.0.0" {
			continue
		}

		log.Printf("Building command...")
		cmd := exec.Command("ping", address, "-c 5", "-i 3")
		log.Printf("cmd: %v", cmd)

		cmd.Stdin = strings.NewReader(address)

		var out bytes.Buffer
		cmd.Stdout = &out

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		log.Printf("Running command...")
		cmd.Run()

		log.Printf("Command output: %s", out.String())

		if strings.Contains(out.String(), "Request timeout") {
			log.Printf("Alert! No response from device %s at address %s", device, address)
		} else if strings.Contains(out.String(), "0.0%") {
			log.Printf("Device %s at address %s responding normally", device, address)
		} else {
			log.Printf("Houston, we have a problem")
		}

	}

	log.Printf("Done")

	return nil
}
