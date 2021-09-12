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
	"strconv"
)

var command = &cmd.StartServer{
	GpxCreator: "https://github.com/outdoorsafetylab/gpxtoolkit",
	WebRoot:    "./webroot",
	Port:       8080,
}

func main() {
	env := os.Getenv("PORT")
	if env != "" {
		val, err := strconv.ParseInt(env, 10, 32)
		if err != nil {
			log.Printf("Using HTTP port: %s", env)
			command.Port = int(val)
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
		command.Service = &elevation.OutdoorSafetyLab{
			Client: http.DefaultClient,
			URL:    env,
			Token:  os.Getenv("ELEVATION_TOKEN"),
		}
	}
	flag.IntVar(&command.Port, "p", command.Port, "HTTP port")
	flag.StringVar(&command.WebRoot, "w", command.WebRoot, "web root dir")
	flag.Parse()
	err := command.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
