package cmd

import (
	"encoding/csv"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"gpxtoolkit/milestone"
	"os"
	"text/template"
)

type CreateMarks struct {
	InputFile string
	Service   elevation.Service
	Creator   string
	Template  string
	Distance  float64
	Reverse   bool
	Format    string
}

func (c *CreateMarks) Run() error {
	log, err := gpx.Open(c.InputFile)
	if err != nil {
		return fmt.Errorf("Failed to open GPX '%s': %s", c.InputFile, err.Error())
	}
	tmpl, err := template.New("").Parse(c.Template)
	if err != nil {
		return fmt.Errorf("Failed to parse template: %s", err.Error())
	}
	marker := &milestone.Marker{
		Distance:     c.Distance,
		NameTemplate: tmpl,
		Reverse:      c.Reverse,
		Service:      c.Service,
	}
	switch c.Format {
	case "gpx":
		err := marker.MarkToGPX(log)
		if err != nil {
			return fmt.Errorf("Failed to mark GPX: %s", err.Error())
		}
		writer := &gpx.Writer{
			Creator: c.Creator,
			Writer:  os.Stdout,
		}
		err = writer.Write(log)
		if err != nil {
			return fmt.Errorf("Failed to write GPX: %s", err.Error())
		}
		return nil
	case "csv":
		records := [][]string{
			{"Name", "Latitude", "Longitude", "Elevation"},
		}
		records, err = marker.MarkToCSV(records, log)
		if err != nil {
			return fmt.Errorf("Failed to create CSV: %s", err.Error())
		}
		writer := csv.NewWriter(os.Stdout)
		for _, record := range records {
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("Failed to write CSV: %s", err.Error())
			}
		}
		writer.Flush()
		return nil
	default:
		return fmt.Errorf("Unknown format: %s", c.Format)
	}
}
