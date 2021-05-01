package main

import (
	"flag"
	"fmt"
	"gpxtoolkit/gpx"
	"gpxtoolkit/milestone"
	"os"
)

func help(progname string) {
	fmt.Printf("Usage: %s <gpx>\n", progname)
}

func main() {
	progname := os.Args[0]
	flag.Parse()
	file := flag.Arg(0)
	if file == "" {
		help(progname)
		os.Exit(1)
	}
	log, err := gpx.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open GPX '%s': %s\n", file, err.Error())
		os.Exit(1)
	}
	marker := &milestone.Marker{
		Distance:     100,
		NameTemplate: `5{{printf "%02d" .Index}}/{{printf "%.1f" .Kilometer}}K`,
	}
	err = marker.Mark(log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to mark GPX: %s\n", err.Error())
		os.Exit(1)
	}
	writer := &gpx.Writer{
		Creator: progname,
		Writer:  os.Stdout,
	}
	err = writer.Write(log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write GPX: %s\n", err.Error())
		os.Exit(1)
	}
}
