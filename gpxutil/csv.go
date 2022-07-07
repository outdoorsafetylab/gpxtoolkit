package gpxutil

import (
	"encoding/csv"
	"fmt"
	"gpxtoolkit/gpx"
	"io"
)

type CSVPointColumn struct {
	Name  string
	Value func(trackIndex int, track *gpx.Track, segmentIndex int, segment *gpx.Segment, pointIndex int, point *gpx.Point) string
}

type CSVPointWriter struct {
	Writer  *csv.Writer
	Columns []CSVPointColumn
}

func (c *CSVPointWriter) Name() string {
	return "Export Points to CSV"
}

func NewCSVPointWriter(writer io.Writer) *CSVPointWriter {
	return &CSVPointWriter{
		Writer: csv.NewWriter(writer),
		Columns: []CSVPointColumn{
			{
				Name: "Track",
				Value: func(trackIndex int, track *gpx.Track, segmentIndex int, segment *gpx.Segment, pointIndex int, point *gpx.Point) string {
					if track.Name != nil {
						return track.GetName()
					}
					return fmt.Sprintf("Track[%d]", trackIndex)
				},
			},
			{
				Name: "Segment",
				Value: func(trackIndex int, track *gpx.Track, segmentIndex int, segment *gpx.Segment, pointIndex int, point *gpx.Point) string {
					return fmt.Sprintf("Segment[%d]", segmentIndex)
				},
			},
			{
				Name: "Time",
				Value: func(trackIndex int, track *gpx.Track, segmentIndex int, segment *gpx.Segment, pointIndex int, point *gpx.Point) string {
					return fmt.Sprintf("%v", point.Time())
				},
			},
			{
				Name: "Latitude",
				Value: func(trackIndex int, track *gpx.Track, segmentIndex int, segment *gpx.Segment, pointIndex int, point *gpx.Point) string {
					return fmt.Sprintf("%v", point.GetLatitude())
				},
			},
			{
				Name: "Longitude",
				Value: func(trackIndex int, track *gpx.Track, segmentIndex int, segment *gpx.Segment, pointIndex int, point *gpx.Point) string {
					return fmt.Sprintf("%v", point.GetLongitude())
				},
			},
			{
				Name: "Elevation",
				Value: func(trackIndex int, track *gpx.Track, segmentIndex int, segment *gpx.Segment, pointIndex int, point *gpx.Point) string {
					return fmt.Sprintf("%v", point.GetElevation())
				},
			},
		},
	}
}

func (w *CSVPointWriter) Run(tracklog *gpx.TrackLog) (int, error) {
	names := make([]string, len(w.Columns))
	for i, c := range w.Columns {
		names[i] = c.Name
	}
	if err := w.Writer.Write(names); err != nil {
		return 0, err
	}
	n := 0
	for i, t := range tracklog.Tracks {
		for j, s := range t.Segments {
			for k, p := range s.Points {
				values := make([]string, len(w.Columns))
				for l, c := range w.Columns {
					values[l] = c.Value(i, t, j, s, k, p)
				}
				if err := w.Writer.Write(values); err != nil {
					return 0, err
				}
				n++
			}
		}
	}
	w.Writer.Flush()
	return n, nil
}
