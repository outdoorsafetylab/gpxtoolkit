package cmd

import (
	"fmt"
	"io"
	"os"
)

type CreateChart struct {
	InputFile  string
	Width      int
	Height     int
	Background string
	Margin     struct {
		Top     float32
		Left    float32
		Botttom float32
		Right   float32
	}
}

func (c *CreateChart) Run() error {
	if c.Width <= 0 {
		return fmt.Errorf("Invalid width: %d", c.Width)
	}
	if c.Height <= 0 {
		return fmt.Errorf("Invalid height: %d", c.Height)
	}
	// log, err := gpx.Open(c.InputFile)
	// if err != nil {
	// 	return fmt.Errorf("Failed to open GPX '%s': %s", c.InputFile, err.Error())
	// }
	w := newIndenter(os.Stdout, "  ")
	w.printf("<svg width=\"%d\" height=\"%d\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\">\n", c.Width, c.Height)
	w.increase()
	w.printf("<g id=\"background\">\n")
	w.increase()
	w.printf("<rect width=\"%d\" height=\"%d\" fill=\"%s\"/>\n", c.Width, c.Height, c.Background)
	w.decrease()
	w.printf("</g>\n")
	w.decrease()
	w.printf("</svg>")
	return nil
}

type indenter struct {
	writer      io.Writer
	level       int
	indentation string
}

func newIndenter(writer io.Writer, indentation string) *indenter {
	return &indenter{writer: writer, indentation: indentation}
}

func (w *indenter) increase() {
	w.level++
}

func (w *indenter) decrease() {
	w.level--
}

func (w *indenter) printf(format string, args ...interface{}) {
	for i := 0; i < w.level; i++ {
		fmt.Fprintf(w.writer, w.indentation)
	}
	fmt.Fprintf(w.writer, format, args...)
}
