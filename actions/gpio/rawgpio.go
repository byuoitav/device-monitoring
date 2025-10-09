package gpio

import (
	"fmt"

	gpiocdev "github.com/warthog618/go-gpiocdev"
)

const (
	IN  = 0
	OUT = 1

	LOW  = 0
	HIGH = 1
)

type GPIO struct {
	chip string
	pin  int
	line *gpiocdev.Line
}

func NewGPIO(pin int) *GPIO {
	return &GPIO{chip: "gpiochip0", pin: pin}
}

// OpenInput requests the line as input and keeps it open.
func (g *GPIO) OpenInput() error {
	l, err := gpiocdev.RequestLine(g.chip, g.pin, gpiocdev.AsInput)
	if err != nil {
		return fmt.Errorf("request input %s/%d: %w", g.chip, g.pin, err)
	}
	g.line = l
	return nil
}

// OpenOutput requests the line as output with an initial value (0/1) and keeps it open.
func (g *GPIO) OpenOutput(initial int) error {
	init := 0
	if initial != 0 {
		init = 1
	}
	l, err := gpiocdev.RequestLine(g.chip, g.pin, gpiocdev.AsOutput(init))
	if err != nil {
		return fmt.Errorf("request output %s/%d: %w", g.chip, g.pin, err)
	}
	g.line = l
	return nil
}

func (g *GPIO) Close() error {
	if g.line != nil {
		err := g.line.Close()
		g.line = nil
		return err
	}
	return nil
}

// Read returns 0/1 from an already-open INPUT line.
// If not yet open, it will try to open as input first.
func (g *GPIO) Read() (int, error) {
	if g.line == nil {
		if err := g.OpenInput(); err != nil {
			return 0, err
		}
	}
	v, err := g.line.Value()
	if err != nil {
		return 0, fmt.Errorf("read value: %w", err)
	}
	if v != 0 {
		return 1, nil
	}
	return 0, nil
}

// Write sets 0/1 on an already-open OUTPUT line.
// If not yet open, it will try to open as output with provided value.
func (g *GPIO) Write(value int) error {
	if g.line == nil {
		return g.OpenOutput(value)
	}
	v := 0
	if value != 0 {
		v = 1
	}
	if err := g.line.SetValue(v); err != nil {
		return fmt.Errorf("set value: %w", err)
	}
	return nil
}

// --- Legacy sysfs shims kept as no-ops so higher-level code compiles ---
func (g *GPIO) Export() error            { return nil }
func (g *GPIO) Unexport() error          { return g.Close() }
func (g *GPIO) SetDirection(_ int) error { return nil }
