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
	alpha := 0.7
	stats := tracklog.Stat(alpha)

	query := r.URL.Query()
	commands := &gpxutil.ChainedCommands{
		Commands: []gpxutil.Command{
			// gpxutil.RemoveDuplicated(),
			// gpxutil.RemoveOutlierBySpeed(),
			// &gpxutil.RemoveOutlierByEIF{Threshold: 0.7},
			&gpxutil.Interpolate{
				Service:  c.Service,
				Distance: 100,
			},
			// &gpxutil.Simplify{
			// 	Service: c.Service,
			// 	Epsilon: 35,
			// 	First:   true,
			// },
			// &gpxutil.CorrectElevation{
			// 	Service: c.Service,
			// },
		},
	}
	_, err = commands.Run(tracklog)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Printf("Before %v", stats)
	log.Printf("After %v", tracklog.Stat(alpha))

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
