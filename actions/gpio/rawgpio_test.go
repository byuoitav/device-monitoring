package gpio

import (
	"fmt"
	"time"
)

func main() {
	// Example usage
	inputPin := NewGPIO(24)
	outputPin := NewGPIO(4)

	// Check initial state of all GPIO pins
	fmt.Println("Initial GPIO states:")
	if err := CheckAllGPIOStates(); err != nil {
		fmt.Println("Error checking GPIO states:", err)
		return
	}

	// Setup
	if err := inputPin.Export(); err != nil {
		fmt.Println("Error exporting input pin:", err)
		return
	}
	if err := outputPin.Export(); err != nil {
		fmt.Println("Error exporting output pin:", err)
		return
	}

	defer inputPin.Unexport()
	defer outputPin.Unexport()

	// Set direction
	if err := inputPin.SetDirection(IN); err != nil {
		fmt.Println("Error setting input pin direction:", err)
		return
	}

	if err := outputPin.SetDirection(OUT); err != nil {
		fmt.Println("Error setting output pin direction:", err)
		return
	}

	// Check state after setup
	fmt.Println("\nGPIO States after setup:")
	if err := CheckAllGPIOStates(); err != nil {
		fmt.Println("Error checking GPIO states:", err)
		return
	}

	// Loop and write to output pin
	for i := 0; i < 10; i++ {
		if err := outputPin.Write(i % 2); err != nil {
			fmt.Println("Error writing to output pin:", err)
			return
		}
		value, err := inputPin.Read()
		if err != nil {
			fmt.Println("Error reading from input pin:", err)
			return
		}
		fmt.Printf("Reading %d from GPIO %d\n", value, inputPin.pin)

		// Sleep to see the output
		time.Sleep(500 * time.Millisecond)
	}

	// Check state after loop
	fmt.Println("\nFinal GPIO States:")
	if err := CheckAllGPIOStates(); err != nil {
		fmt.Println("Error checking GPIO states:", err)
		return
	}
}
