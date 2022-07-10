package main

import (
	"flag"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"gpxtoolkit/router"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

var progname string
var service elevation.Service
var command = &Command{
	template: `{{printf "%.1f" .Kilometer}}K`,
	symbol:   "Milestone",
	distance: 100,
	reverse:  false,
	format:   "gpx",
}
var server = &Server{
	webroot: "./webroot",
	port:    8080,
}

func help(progname string) {
	fmt.Printf("Usage: %s <gpx>\n", progname)
}

func main() {
	progname = filepath.Base(os.Args[0])
	daemon := false
	env := os.Getenv("PORT")
	if env != "" {
		val, err := strconv.ParseInt(env, 10, 32)
		if err != nil {
			log.Printf("Using HTTP port: %s", env)
			server.port = int(val)
		}
	}
	env = os.Getenv("ELEVATION_URL")
	if env != "" {
		url, err := url.Parse(env)
		if err != nil || (url.Scheme != "http" && url.Scheme != "https") {
			log.Printf("Invalid URL of elevation service: '%s'", env)
			os.Exit(1)
			return
		}
		log.Printf("Using elevation service: %s", env)
		service = &elevation.OutdoorSafetyLab{
			Client: http.DefaultClient,
			URL:    env,
			Token:  os.Getenv("ELEVATION_TOKEN"),
		}
	}
	flag.StringVar(&command.template, "n", command.template, "template of milestone name")
	flag.StringVar(&command.symbol, "s", command.symbol, "GPX symbol of milestone")
	flag.Float64Var(&command.distance, "d", command.distance, "distance between milestone (in meter)")
	flag.BoolVar(&command.reverse, "r", command.reverse, "create milestone in reverse way")
	flag.StringVar(&command.format, "f", command.format, "output format, one of 'gpx', 'csv'")
	flag.StringVar(&command.output, "o", command.output, "path of output file")

	flag.BoolVar(&daemon, "D", daemon, "run as a HTTP server")
	flag.IntVar(&server.port, "p", server.port, "HTTP port")
	flag.StringVar(&server.webroot, "w", server.webroot, "web root dir")
	flag.Parse()
	var err error
	if daemon {
		err = server.Run()
	} else {
		err = command.Run()
	}
	if err != nil {
		os.Exit(1)
	}
}

type Command struct {
	template string
	distance float64
	symbol   string
	reverse  bool
	format   string
	output   string
}

func (c *Command) Run() error {
	file := flag.Arg(0)
	if file == "" {
		err := fmt.Errorf("Missing argument of input file")
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		help(progname)
		return err
	}
	gpxLog, err := gpx.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open GPX '%s': %s\n", file, err.Error())
		return err
	}
	marker := &gpxutil.Milestone{
		Distance: c.distance,
		MilestoneName: &gpxutil.MilestoneName{
			Template: c.template,
		},
		Reverse: c.reverse,
		Symbol:  c.symbol,
	}
	var output io.Writer
	if command.output == "" {
		log.SetOutput(ioutil.Discard)
		output = os.Stdout
	} else {
		output, err = os.Create(command.output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed create output file: %s\n", err.Error())
			return err
		}
	}
	switch c.format {
	case "gpx":
		_, err := marker.Run(gpxLog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark GPX: %s\n", err.Error())
			return err
		}
		writer := &gpx.Writer{
			Creator: progname,
			Writer:  output,
		}
		err = writer.Write(gpxLog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write GPX: %s\n", err.Error())
			return err
		}
		return nil
	case "csv":
		_, err := marker.Run(gpxLog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mark GPX: %s\n", err.Error())
			return err
		}
		csv := gpxutil.NewCSVWayPointWriter(output)
		_, err = csv.Run(gpxLog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write CSV: %s\n", err.Error())
			return err
		}
		return nil
	default:
		err := fmt.Errorf("Unknown format: %s", c.format)
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}
}

type Server struct {
	webroot string
	port    int
}

func (s *Server) Run() error {
	r := router.NewRouter(s.webroot, service)
	log.Printf("Listening port %d...", s.port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen %d: %s\n", s.port, err.Error())
	}
	return err
}
