package milestone

import (
	"bytes"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"math"
	"text/template"
	"time"

	"google.golang.org/protobuf/proto"
)

type Marker struct {
	Distance     float64
	NameTemplate *template.Template
	Symbol       string
	Reverse      bool
	Service      elevation.Service
}

func (m *Marker) MarkToGPX(log *gpx.TrackLog) error {
	marks, err := m.Marks(log)
	if err != nil {
		return err
	}
	log.WayPoints = append(log.WayPoints, marks...)
	return nil
}

func (m *Marker) MarkToCSV(csv [][]string, log *gpx.TrackLog) ([][]string, error) {
	marks, err := m.Marks(log)
	if err != nil {
		return nil, err
	}
	for _, m := range marks {
		csv = append(csv, []string{m.GetName(), fmt.Sprintf("%f", m.GetLatitude()), fmt.Sprintf("%f", m.GetLongitude()), fmt.Sprintf("%f", m.GetElevation())})
	}
	return csv, nil
}

func (m *Marker) Marks(log *gpx.TrackLog) ([]*gpx.WayPoint, error) {
	allMarks := make([]*gpx.WayPoint, 0)
	for _, t := range log.Tracks {
		for _, s := range t.Segments {
			points := s.Points
			if m.Reverse {
				n := len(s.Points)
				points = make([]*gpx.Point, n)
				for i, p := range s.Points {
					points[n-1-i] = p
				}
			}
			marks, err := m.marks(points)
			if err != nil {
				return nil, err
			}
			allMarks = append(allMarks, marks...)
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

func (m *Marker) marks(points []*gpx.Point) ([]*gpx.WayPoint, error) {
	res := make([]*gpx.WayPoint, 0)
	remainder := float64(0)
	var a *gpx.Point
	for _, b := range points {
		if a != nil {
			dist := b.DistanceTo(a)
			if (remainder + dist) >= m.Distance {
				first := (m.Distance - remainder)
				lat1 := a.GetLatitude()
				lat2 := b.GetLatitude()
				lon1 := a.GetLongitude()
				lon2 := b.GetLongitude()
				dlat := lat2 - lat1
				dlon := lon2 - lon1
				ele1 := a.GetElevation()
				ele2 := b.GetElevation()
				dele := ele2 - ele1
				t1 := a.Time()
				t2 := b.Time()
				dt := t2.Sub(t1)
				pos := first
				for pos < dist {
					ratio := pos / dist
					lat := lat1 + dlat*ratio
					lon := lon1 + dlon*ratio
					payload := &templatePayload{
						Index: len(res),
					}
					payload.Number = payload.Index + 1
					payload.Meter = float64(payload.Number) * m.Distance
					payload.Kilometer = payload.Meter / 1000
					var buf bytes.Buffer
					err := m.NameTemplate.Execute(&buf, payload)
					if err != nil {
						return nil, err
					}
					wpt := &gpx.WayPoint{
						Latitude:  proto.Float64(lat),
						Longitude: proto.Float64(lon),
						Name:      proto.String(buf.String()),
					}
					if a.Elevation != nil && b.Elevation != nil {
						wpt.Elevation = proto.Float64(ele1 + dele*ratio)
					}
					if a.NanoTime != nil && b.NanoTime != nil {
						wpt.NanoTime = proto.Int64(t1.Add(dt * time.Duration(ratio)).UnixNano())
					}
					if m.Symbol != "" {
						wpt.Symbol = proto.String(m.Symbol)
					}
					res = append(res, wpt)
					pos += m.Distance
				}
				remainder = math.Mod(remainder+dist, m.Distance)
			} else {
				remainder += dist
			}
		}
		a = b
	}
	return res, nil
}

type templatePayload struct {
	Index     int
	Number    int
	Meter     float64
	Kilometer float64
}
