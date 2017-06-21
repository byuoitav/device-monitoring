package device

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/device-monitoring-microservice/logstash"
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

func PingAddresses(building, room string, addresses map[string]string) error {

	log.Printf("Pinging with bash commands...")

	for device, address := range addresses {

		if address != "0.0.0.0" {

			go Ping(building, room, device, address)

		}

	}

	log.Printf("Done")

	return nil
}

func Ping(building, room, device, address string) {

	log.Printf("Pinging %s", address)

	log.Printf("Building command...")
	cmd := exec.Command("ping", address, "-c 5", "-i 3")
	log.Printf("cmd: %v", cmd)

	cmd.Stdin = strings.NewReader(address)

	var out bytes.Buffer
	cmd.Stdout = &out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Printf("Running command...")
	timestamp := string(time.Now().Format(time.RFC3339))
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running command: %s", err.Error())
		return
	}

	log.Printf("Command output: %s", out.String())

	if strings.Contains(out.String(), "Request timeout") {

		log.Printf("Alert! No response from device %s at address %s", device, address)
		err = logstash.SendEvent(building, room, timestamp, device, "Not responding")

	} else {

		log.Printf("Device %s at address %s responding normally", device, address)
		err = logstash.SendEvent(building, room, timestamp, device, "Responding")

	}

	if err != nil {
		log.Printf("Error sending event: %s", err.Error())
	}

}
