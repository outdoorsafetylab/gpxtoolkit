package server

import (
	"fmt"
	"github.com/outdoorsafetylab/gpxtoolkit/elevation"
	"github.com/outdoorsafetylab/gpxtoolkit/log"
	"net/http"
	"os"
)

type Server struct {
	Webroot string
	Port    int16
}

func (s *Server) Run(service elevation.Service) error {
	r := NewRouter(s.Webroot, service)
	log.Infof("Listening port %d...", s.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.Port), r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen %d: %s\n", s.Port, err.Error())
	}
	return err
}
