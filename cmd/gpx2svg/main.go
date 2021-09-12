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
	Background: "#40a040",
	TilePadding: struct {
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
	flag.StringVar(&command.Background, "bg", command.Background, "background color")
	flag.IntVar(&command.TilePadding.Top, "tp", command.TilePadding.Top, "top padding")
	flag.IntVar(&command.TilePadding.Left, "lp", command.TilePadding.Left, "left padding")
	flag.IntVar(&command.TilePadding.Botttom, "bp", command.TilePadding.Botttom, "bottom padding")
	flag.IntVar(&command.TilePadding.Right, "rp", command.TilePadding.Right, "right padding")
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
