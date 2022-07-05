package controller

import (
	"encoding/csv"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/milestone"
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
	log, err := gpx.Parse(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	tmpl, err := template.New("").Parse(query.Get("name-template"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %s", err.Error()), 400)
		return
	}
	marker := &milestone.Marker{
		Distance:     distance,
		NameTemplate: tmpl,
		Service:      c.Service,
		Symbol:       "Milestone",
	}
	if query.Get("reverse") == "true" {
		marker.Reverse = true
	}
	format := query.Get("format")
	switch format {
	case "gpx":
		err = marker.MarkToGPX(log)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to mark GPX: %s", err.Error()), 500)
			return
		}
		writer := &gpx.Writer{
			Creator: c.GPXCreator,
			Writer:  w,
		}
		w.Header().Set("Content-Type", "application/gpx+xml")
		err = writer.Write(log)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to write GPX: %s", err.Error()), 500)
			return
		}
		return
	case "csv":
		records := [][]string{
			{r.FormValue("csv-name"), r.FormValue("csv-latitude"), r.FormValue("csv-longitude"), r.FormValue("csv-elevation")},
		}
		records, err = marker.MarkToCSV(records, log)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create CSV: %s", err.Error()), 500)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		writer := csv.NewWriter(w)
		for _, record := range records {
			if err := writer.Write(record); err != nil {
				http.Error(w, fmt.Sprintf("Failed to write CSV: %s", err.Error()), 500)
				return
			}
		}
		writer.Flush()
		return
	default:
		http.Error(w, fmt.Sprintf("Unknown format: %s", format), 400)
		return
	}
}
