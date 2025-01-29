package gpio

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	IN  = 0
	OUT = 1

	LOW  = 0
	HIGH = 1
)

type GPIO struct {
	pin int
}

func NewGPIO(pin int) *GPIO {
	return &GPIO{pin: pin}
}

// Export exports the pin to userspace.
func (g *GPIO) Export() error {
	return writeFile("/sys/class/gpio/export", strconv.Itoa(g.pin))
}

// Unexport unexports the pin from userspace.
func (g *GPIO) Unexport() error {
	return writeFile("/sys/class/gpio/unexport", strconv.Itoa(g.pin))
}

func (g *GPIO) SetDirection(direction int) error {
	var dir string
	if direction == IN {
		dir = "in"
	} else {
		dir = "out"
	}

	return writeFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", g.pin), dir)
}

func (g *GPIO) Read() (int, error) {
	data, err := os.ReadFile(fmt.Sprintf("/sys/class/gpio/gpio%d", g.pin))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data[:len(data)-1]))
}

func (g *GPIO) Write(value int) error {
	var val string
	if value == LOW {
		val = "0"
	} else {
		val = "1"
	}
	return writeFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", g.pin), val)
}

func writeFile(filepath, data string) error {
	return os.WriteFile(filepath, []byte(data), 0644)
}

func (g *GPIO) CheckState() (bool, string, error) {
	// Check if the pin is exported
	_, err := os.Stat(fmt.Sprintf("/sys/class/gpio/gpio%d", g.pin))
	if os.IsNotExist(err) {
		return false, "", nil // Pin is not exported
	}
	if err != nil {
		return false, "", err
	}

	// Pin is exported, check direction
	dirData, err := os.ReadFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", g.pin))
	if err != nil {
		return true, "", fmt.Errorf("error reading direction: %v", err)
	}
	direction := strings.TrimSpace(string(dirData))

	return true, direction, nil
}

func CheckAllGPIOStates() error {
	files, err := filepath.Glob("/sys/class/gpio/gpio*")
	if err != nil {
		return fmt.Errorf("error listing GPIO pins: %v", err)
	}

	for _, file := range files {
		pinStr := strings.TrimPrefix(filepath.Base(file), "gpio")
		pin, err := strconv.Atoi(pinStr)
		if err != nil {
			fmt.Printf("error converting pin number %q to int: %v\n", pinStr, err)
			continue
		}

		g := NewGPIO(pin)
		exported, direction, err := g.CheckState()
		if err != nil {
			fmt.Printf("error checking state of pin %d: %v\n", pin, err)
			continue
		}

		if exported {
			fmt.Printf("Pin %d is exported and set to %q\n", pin, direction)
		} else {
			fmt.Printf("Pin %d is not exported\n", pin)
		}
	}

	return nil
}
