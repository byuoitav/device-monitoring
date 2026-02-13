package gpio

import (
	"fmt"
	"time"
)

func main() {
	in := NewGPIO(24) // <-- change if needed
	out := NewGPIO(4) // <-- change if needed

	// Open lines
	if err := in.OpenInput(); err != nil {
		panic(fmt.Errorf("open input: %w", err))
	}
	defer in.Close()

	if err := out.OpenOutput(LOW); err != nil {
		panic(fmt.Errorf("open output: %w", err))
	}
	defer out.Close()

	fmt.Println("Toggling output and reading input (10 cycles):")
	for i := 0; i < 10; i++ {
		val := i % 2
		if err := out.Write(val); err != nil {
			panic(fmt.Errorf("write: %w", err))
		}
		// small settle time
		time.Sleep(50 * time.Millisecond)

		r, err := in.Read()
		if err != nil {
			panic(fmt.Errorf("read: %w", err))
		}
		fmt.Printf("out=%d -> in=%d\n", val, r)

		time.Sleep(450 * time.Millisecond)
	}
	fmt.Println("Done.")
}
