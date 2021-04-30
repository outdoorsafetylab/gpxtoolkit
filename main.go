package main

import (
	"flag"
	"fmt"
	"gpxtoolkit/gpx"
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
	r, err := os.Open(file)
	if err != nil {
		fmt.Printf("Failed to open file '%s': %s\n", file, err.Error())
		os.Exit(1)
	}
	parser := &gpx.Parser{}
	log, err := parser.Parse(r)
	if err != nil {
		fmt.Printf("Failed to read GPX '%s': %s\n", file, err.Error())
		os.Exit(1)
	}
	writer := &gpx.Writer{
		Creator: progname,
		Writer:  os.Stdout,
	}
	err = writer.Write(log)
	if err != nil {
		fmt.Printf("Failed to write GPX: %s\n", err.Error())
		os.Exit(1)
	}
}
