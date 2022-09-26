package main

import (
	"flag"
	"fmt"
	"gpxtoolkit/cmd"
	"os"
)

var command = &cmd.CreateChart{
	Width:      1024,
	Height:     512,
	Background: "#40a040",
	Margin: struct {
		Top     float32
		Left    float32
		Botttom float32
		Right   float32
	}{
		Top:     0.1,
		Left:    0.1,
		Botttom: 0.1,
		Right:   0.1,
	},
}

func main() {
	flag.IntVar(&command.Width, "w", command.Width, "width")
	flag.IntVar(&command.Height, "h", command.Height, "height")
	flag.StringVar(&command.Background, "bg", command.Background, "background color")
	flag.Parse()
	command.InputFile = flag.Arg(0)
	if command.InputFile == "" {
		fmt.Fprintf(os.Stderr, "Missing argument of input file\n")
		flag.Usage()
		os.Exit(1)
	}
	err := command.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
