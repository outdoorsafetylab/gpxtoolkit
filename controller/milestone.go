package controller

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/gpxutil"
	"net/http"
)

type MilestoneController struct {
	Service elevation.Service
}

func (c *MilestoneController) Handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	tracklog, err := gpx.Parse(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	name := &gpxutil.MilestoneName{
		Template: query.Get("template"),
	}
	_, err = name.Eval(&gpxutil.MilestoneNameVariables{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid template: %s", err.Error()), 400)
		return
	}
	distance := queryGetFloat64(query, "distance", 100)
	commands := &gpxutil.ChainedCommands{
		Commands: []gpxutil.Command{
			gpxutil.RemoveDistanceLessThan(0.1),
			// gpxutil.RemoveOutlierBySpeed(),
			// &gpxutil.RemoveOutlierByEIF{Threshold: 0.7},
			// &gpxutil.Simplify{
			// 	Epsilon: 10,
			// 	First:   true,
			// },
			// &gpxutil.ReTimestamp{
			// 	Start: time.Unix(0, 0),
			// 	Speed: 0.5,
			// },
			// &gpxutil.ReSegment{
			// 	Threshold: 500,
			// },
			// &gpxutil.CorrectElevation{
			// 	Service: c.Service,
			// },
			&gpxutil.Milestone{
				Service:           c.Service,
				Distance:          distance,
				MilestoneName:     name,
				Reverse:           queryGetBool(query, "reverse", false),
				Symbol:            queryGetString(query, "symbol", "Milestone"),
				FitWaypoints:      queryGetBool(query, "fits", false),
				ByTerrainDistance: queryGetBool(query, "terrainDistance", false),
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
			Creator: r.Host,
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
