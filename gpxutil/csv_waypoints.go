package gpxutil

import (
	"encoding/csv"
	"fmt"
	"gpxtoolkit/gpx"
	"gpxtoolkit/twd97"
	"io"
)

type CSVWayPointColumn struct {
	Name  string
	Value func(index int, waypoint *gpx.WayPoint) string
}

type CSVWayPointWriter struct {
	Writer  *csv.Writer
	Columns []CSVWayPointColumn
}

func (c *CSVWayPointWriter) Name() string {
	return "Export Points to CSV"
}

func NewCSVWayPointWriter(writer io.Writer) *CSVWayPointWriter {
	return &CSVWayPointWriter{
		Writer: csv.NewWriter(writer),
		Columns: []CSVWayPointColumn{
			{
				Name: "名稱",
				Value: func(index int, point *gpx.WayPoint) string {
					if point.Name != nil {
						return point.GetName()
					}
					return fmt.Sprintf("Waypoint[%d]", index)
				},
			},
			{
				Name: "時間",
				Value: func(index int, point *gpx.WayPoint) string {
					return fmt.Sprintf("%v", point.Time())
				},
			},
			{
				Name: "緯度",
				Value: func(index int, point *gpx.WayPoint) string {
					return fmt.Sprintf("%.6f", point.GetLatitude())
				},
			},
			{
				Name: "經度",
				Value: func(index int, point *gpx.WayPoint) string {
					return fmt.Sprintf("%.6f", point.GetLongitude())
				},
			},
			{
				Name: "TWD97東距",
				Value: func(index int, point *gpx.WayPoint) string {
					x, _ := twd97.FromWGS84(point.GetLongitude(), point.GetLatitude(), false)
					return fmt.Sprintf("%.0f", x)
				},
			},
			{
				Name: "TWD97北距",
				Value: func(index int, point *gpx.WayPoint) string {
					_, y := twd97.FromWGS84(point.GetLongitude(), point.GetLatitude(), false)
					return fmt.Sprintf("%.0f", y)
				},
			},
			{
				Name: "高程",
				Value: func(index int, point *gpx.WayPoint) string {
					return fmt.Sprintf("%.1f", point.GetElevation())
				},
			},
		},
	}
}

func (w *CSVWayPointWriter) Run(tracklog *gpx.TrackLog) (int, error) {
	names := make([]string, len(w.Columns))
	for i, c := range w.Columns {
		names[i] = c.Name
	}
	if err := w.Writer.Write(names); err != nil {
		return 0, err
	}
	for i, p := range tracklog.WayPoints {
		values := make([]string, len(w.Columns))
		for l, c := range w.Columns {
			values[l] = c.Value(i, p)
		}
		if err := w.Writer.Write(values); err != nil {
			return 0, err
		}
	}
	w.Writer.Flush()
	return len(tracklog.WayPoints), nil
}
