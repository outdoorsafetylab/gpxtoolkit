package controller

import (
	"encoding/csv"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/milestone"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"text/template"
)

type MilestoneController struct {
	GPXCreator string
	Service    elevation.Service
}

func (c *MilestoneController) Handler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 * 1048576)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	distance, err := strconv.ParseFloat(r.FormValue("distance"), 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid distance: %s", err.Error()), 400)
		return
	}
	f, h, err := r.FormFile("gpx-file")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()
	log, err := gpx.Parse(f)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	tmpl, err := template.New("").Parse(r.FormValue("name-template"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %s", err.Error()), 400)
		return
	}
	marker := &milestone.Marker{
		Distance:     distance,
		NameTemplate: tmpl,
		Service:      c.Service,
	}
	if _, ok := r.Form["reverse"]; ok {
		marker.Reverse = true
	}
	format := r.FormValue("format")
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
		extname := filepath.Ext(h.Filename)
		basename := h.Filename[0 : len(h.Filename)-len(extname)]
		filename := fmt.Sprintf("%s%s%s", basename, r.FormValue("filename-suffix"), extname)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename*=UTF-8''%s`, url.PathEscape(filename)))
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
		extname := filepath.Ext(h.Filename)
		basename := h.Filename[0 : len(h.Filename)-len(extname)]
		filename := fmt.Sprintf("%s%s%s", basename, r.FormValue("filename-suffix"), ".csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename*=UTF-8''%s`, url.PathEscape(filename)))
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
