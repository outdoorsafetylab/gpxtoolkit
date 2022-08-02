package router

import (
	"net/http"

	"gpxtoolkit/controller"
	"gpxtoolkit/elevation"
	"gpxtoolkit/middleware"

	"github.com/gorilla/mux"
)

func NewRouter(webroot string, service elevation.Service) http.Handler {
	r := mux.NewRouter()
	sub := r.PathPrefix("/cgi").Subrouter()
	sub.Use(middleware.Dump)
	sub.Use(middleware.NoCache)
	version := &controller.VersionController{}
	sub.HandleFunc("/version", version.Get).Methods("GET")
	milestone := &controller.MilestoneController{
		Service: service,
	}
	sub.HandleFunc("/milestones", milestone.Handler).Methods("POST")
	correct := &controller.CorrectController{
		Service: service,
	}
	sub.HandleFunc("/correct", correct.Handler).Methods("POST")
	r.NotFoundHandler = &weboortHandler{
		path: webroot,
	}
	return r
}
