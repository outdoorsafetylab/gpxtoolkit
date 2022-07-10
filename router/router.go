package router

import (
	"gpxtoolkit/controller"
	"gpxtoolkit/elevation"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(webroot string, service elevation.Service) http.Handler {
	r := mux.NewRouter()
	sub := r.PathPrefix("/cgi").Subrouter()
	milestone := &controller.MilestoneController{
		Service: service,
	}
	sub.HandleFunc("/milestones", milestone.Handler).Methods("POST")
	correct := &controller.CorrectController{
		Service: service,
	}
	sub.HandleFunc("/correct", correct.Handler).Methods("POST")
	r.NotFoundHandler = http.FileServer(http.Dir(webroot))
	return r
}
