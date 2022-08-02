package main

import (
	"flag"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/log"
	"gpxtoolkit/router"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var service elevation.Service
var server = &Server{
	webroot: "./webroot/dist",
	port:    8080,
}

func main() {
	dev := false
	flag.BoolVar(&dev, "d", dev, "Development mode")
	flag.IntVar(&server.port, "p", server.port, "HTTP port")
	flag.StringVar(&server.webroot, "w", server.webroot, "web root dir")
	flag.Parse()
	err := log.Init(dev)
	if err != nil {
		os.Exit(1)
	}
	env := os.Getenv("PORT")
	if env != "" {
		val, err := strconv.ParseInt(env, 10, 32)
		if err == nil {
			log.Infof("Using HTTP port: %s", env)
			server.port = int(val)
		}
	}
	env = os.Getenv("ELEVATION_URL")
	if env != "" {
		url, err := url.Parse(env)
		if err != nil || (url.Scheme != "http" && url.Scheme != "https") {
			log.Errorf("Invalid URL of elevation service: '%s'", env)
			os.Exit(1)
			return
		}
		log.Infof("Using elevation service: %s", env)
		service = &elevation.OutdoorSafetyLab{
			Client: http.DefaultClient,
			URL:    env,
			Token:  os.Getenv("ELEVATION_TOKEN"),
		}
	}
	err = server.Run()
	if err != nil {
		os.Exit(1)
	}
}

type Server struct {
	webroot string
	port    int
}

func (s *Server) Run() error {
	r := router.NewRouter(s.webroot, service)
	log.Infof("Listening port %d...", s.port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen %d: %s\n", s.port, err.Error())
	}
	return err
}
