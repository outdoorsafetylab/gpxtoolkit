package main

import (
	"flag"
	"fmt"
	"gpxtoolkit/cmd"
	"gpxtoolkit/elevation"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var command = &cmd.CreateMarks{
	Template: `{{printf "%.1f" .Kilometer}}K`,
	Distance: 100,
	Reverse:  false,
	Format:   "gpx",
}

func main() {
	progname := filepath.Base(os.Args[0])
	env := os.Getenv("ELEVATION_URL")
	if env != "" {
		url, err := url.Parse(env)
		if err != nil || (url.Scheme != "http" && url.Scheme != "https") {
			log.Printf("Invalid URL of elevation service: '%s'", env)
			os.Exit(1)
			return
		}
		log.Printf("Using elevation service: %s", env)
		command.Service = &elevation.OutdoorSafetyLab{
			Client: http.DefaultClient,
			URL:    env,
			Token:  os.Getenv("ELEVATION_TOKEN"),
		}
	}
	command.Creator = progname
	flag.StringVar(&command.Template, "n", command.Template, "template of milestone name")
	flag.Float64Var(&command.Distance, "d", command.Distance, "distance between milestone (in meter)")
	flag.BoolVar(&command.Reverse, "r", command.Reverse, "create milestone in reverse way")
	flag.StringVar(&command.Format, "f", command.Format, "output format, one of 'gpx', 'csv'")

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
