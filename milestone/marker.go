package milestone

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"log"
	"text/template"
)

type Marker struct {
	Distance     float64
	NameTemplate *template.Template
	Symbol       string
	Reverse      bool
	Service      elevation.Service
}

func (m *Marker) MarkToGPX(tracklog *gpx.TrackLog) error {
	marks, err := m.Marks(tracklog)
	if err != nil {
		return err
	}
	tracklog.WayPoints = append(tracklog.WayPoints, marks...)
	return nil
}

func (m *Marker) MarkToCSV(csv [][]string, tracklog *gpx.TrackLog) ([][]string, error) {
	marks, err := m.Marks(tracklog)
	if err != nil {
		return nil, err
	}
	for _, m := range marks {
		csv = append(csv, []string{m.GetName(), fmt.Sprintf("%f", m.GetLatitude()), fmt.Sprintf("%f", m.GetLongitude()), fmt.Sprintf("%f", m.GetElevation())})
	}
	return csv, nil
}

func (m *Marker) Marks(tracklog *gpx.TrackLog) ([]*gpx.WayPoint, error) {
	allMarks := make([]*gpx.WayPoint, 0)
	for i, t := range tracklog.Tracks {
		log.Printf("Processing trk[%d]: %s", i, t.GetName())
		for j, s := range t.Segments {
			var points gpxPoints = s.Points
			if m.Reverse {
				log.Printf("Reverse processing trk[%d]/trkseg[%d]", i, j)
				n := len(s.Points)
				points = make(gpxPoints, n)
				for i, p := range s.Points {
					points[n-1-i] = p
				}
			} else {
				log.Printf("Forward processing trk[%d]/trkseg[%d]", i, j)
			}
			// marks, err := points.marks(0, m.Distance, m.Distance, m.NameTemplate, m.Symbol)
			// if err != nil {
			// 	return nil, err
			// }
			// allMarks = append(allMarks, marks...)
			projections := points.projectWaypoints(tracklog.WayPoints)
			segments := points.segments(projections, m.Distance)
			total := 0.0
			for i, seg := range segments {
				log.Printf("segment[%d]: %d: %f: %s", i, len(seg.points), seg.distance, seg.waypoint.GetName())
				marks, err := seg.points.marks(total, seg.distance, m.Distance, m.NameTemplate, m.Symbol)
				if err != nil {
					return nil, err
				}
				allMarks = append(allMarks, marks...)
				total += m.Distance * float64(len(marks)+1)
				log.Printf("total: %f", total)
			}
			// for _, p := range projections {
			// 	prj := p.projection
			// 	allMarks = append(allMarks, &gpx.WayPoint{
			// 		NanoTime:  prj.NanoTime,
			// 		Latitude:  prj.Latitude,
			// 		Longitude: prj.Longitude,
			// 		Elevation: prj.Elevation,
			// 		Symbol:    proto.String("Milestone"),
			// 	})
			// }
		}
	}
	if m.Service != nil {
		points := make([]*elevation.LatLon, len(allMarks))
		for i, m := range allMarks {
			points[i] = &elevation.LatLon{Lat: m.GetLatitude(), Lon: m.GetLongitude()}
		}
		eles, err := m.Service.Lookup(points)
		if err != nil {
			return nil, err
		}
		if len(eles) != len(allMarks) {
			return nil, fmt.Errorf("unexpected length of elevation result: expect=%d, actual=%d", len(points), len(eles))
		}
		for i, m := range allMarks {
			m.Elevation = eles[i]
		}
	}
	return allMarks, nil
}
