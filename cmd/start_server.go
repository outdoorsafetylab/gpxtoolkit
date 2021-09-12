package cmd

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/router"
	"log"
	"net/http"
)

type StartServer struct {
	WebRoot    string
	Service    elevation.Service
	GpxCreator string
	Port       int
}

func (s *StartServer) Run() error {
	r := router.NewRouter(s.WebRoot, s.GpxCreator, s.Service)
	log.Printf("Listening port %d...", s.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.Port), r)
	if err != nil {
		return fmt.Errorf("Failed to listen %d: %s", s.Port, err.Error())
	}
	return err
}
