package controller

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"net/http"
	"strconv"
	"text/template"
)

type MilestoneController struct {
	GPXCreator string
	Service    elevation.Service
}

func (c *MilestoneController) Handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	distance, err := strconv.ParseFloat(query.Get("distance"), 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid distance: %s", err.Error()), 400)
		return
	}
	tracklog, err := gpx.Parse(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	tmpl, err := template.New("").Parse(query.Get("name-template"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %s", err.Error()), 400)
		return
	}
	commands := &gpxutil.ChainedCommands{
		Commands: []gpxutil.Command{
			&gpxutil.ProjectWaypoints{Threshold: 100},
			&gpxutil.Milestone{
				Distance:     distance,
				NameTemplate: tmpl,
				Reverse:      query.Get("reverse") == "true",
				Symbol:       "Milestone",
				FitWaypoints: query.Get("fit-waypoints") == "true",
			},
			&gpxutil.CorrectElevation{
				Service: c.Service,
			},
		},
	}
	_, err = commands.Run(tracklog)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	format := query.Get("format")
	switch format {
	case "gpx":
		writer := &gpx.Writer{
			Creator: c.GPXCreator,
			Writer:  w,
		}
		w.Header().Set("Content-Type", "application/gpx+xml")
		err = writer.Write(tracklog)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to write GPX: %s", err.Error()), 500)
			return
		}
		return
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		csv := gpxutil.NewCSVWayPointWriter(w)
		_, err = csv.Run(tracklog)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to write CSV: %s", err.Error()), 500)
			return
		}
		return
	default:
		http.Error(w, fmt.Sprintf("Unknown format: %s", format), 400)
		return
	}
}
