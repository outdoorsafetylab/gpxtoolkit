package controller

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"log"
	"net/http"
)

type CorrectController struct {
	Service elevation.Service
}

func (c *CorrectController) Handler(w http.ResponseWriter, r *http.Request) {
	tracklog, err := gpx.Parse(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	before := tracklog.Stat()
	excluded := 0

	query := r.URL.Query()
	commands := &gpxutil.ChainedCommands{
		Commands: []gpxutil.Command{
			gpxutil.RemoveDistanceLessThan(0.1),
			gpxutil.RemoveOutlierBySpeed(),
			&gpxutil.RemoveOutlierByEIF{Threshold: 0.7},
			&gpxutil.Simplify{
				Epsilon: 10,
				First:   true,
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

	after := tracklog.Stat()
	log.Printf("Before: %v", before)
	log.Printf("After: %v", after)
	log.Printf("Excluded: %v", excluded)

	switch queryGetString(query, "format", "gpx") {
	case "gpx":
		writer := &gpx.Writer{
			Creator: r.Host,
			Writer:  w,
		}
		w.Header().Set("Content-Type", "application/gpx+xml")
		err = writer.Write(tracklog)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to write GPX: %s", err.Error()), 500)
			return
		}
	case "csv":
		writer := gpxutil.NewCSVPointWriter(w)
		w.Header().Set("Content-Type", "text/csv")
		_, err = writer.Run(tracklog)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to write CSV: %s", err.Error()), 500)
			return
		}
	}
}
