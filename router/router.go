package router

import (
	"gpxtoolkit/controller"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(webroot, gpxCreator string) http.Handler {
	r := mux.NewRouter()
	sub := r.PathPrefix("/cgi").Subrouter()
	milestone := &controller.MilestoneController{
		GPXCreator: gpxCreator,
	}
	sub.HandleFunc("/milestones", milestone.Handler).Methods("POST")
	r.NotFoundHandler = http.FileServer(http.Dir(webroot))
	return r
}
