package main

import (
	"flag"
	"fmt"
	"gpxtoolkit/cmd"
	"os"
)

var command = &cmd.CreateSVG{
	ZoomLevel:  16,
	TileWidth:  256,
	TileHeight: 256,
	Padding: struct {
		Top     int
		Left    int
		Botttom int
		Right   int
	}{
		Top:     0,
		Left:    0,
		Botttom: 0,
		Right:   0,
	},
}

func main() {
	flag.IntVar(&command.ZoomLevel, "z", command.ZoomLevel, "zoom level")
	flag.IntVar(&command.Padding.Top, "t", command.Padding.Top, "top padding")
	flag.IntVar(&command.Padding.Left, "l", command.Padding.Left, "left padding")
	flag.IntVar(&command.Padding.Botttom, "b", command.Padding.Botttom, "bottom padding")
	flag.IntVar(&command.Padding.Right, "r", command.Padding.Right, "right padding")
	flag.Parse()
	command.InputFile = flag.Arg(0)
	if command.InputFile == "" {
		fmt.Fprintf(os.Stderr, "Missing argument of input file\n")
		flag.Usage()
		os.Exit(1)
	}
	if command.ZoomLevel <= 0 {
		fmt.Fprintf(os.Stderr, "Please specify zoom level\n")
		flag.Usage()
		os.Exit(1)
	}
	err := command.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
