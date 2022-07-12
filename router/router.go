package router

import (
	"gpxtoolkit/controller"
	"gpxtoolkit/elevation"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(commit, tag, webroot string, service elevation.Service) http.Handler {
	r := mux.NewRouter()
	sub := r.PathPrefix("/cgi").Subrouter()
	version := &controller.VersionController{
		Commit: commit,
		Tag:    tag,
	}
	sub.HandleFunc("/version", version.Handler).Methods("GET")
	milestone := &controller.MilestoneController{
		Service: service,
	}
	sub.HandleFunc("/milestones", milestone.Handler).Methods("POST")
	correct := &controller.CorrectController{
		Service: service,
	}
	sub.HandleFunc("/correct", correct.Handler).Methods("POST")
	r.NotFoundHandler = NoCache(http.FileServer(http.Dir(webroot)))
	return r
}
