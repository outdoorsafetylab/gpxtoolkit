package milestone

import (
	"bytes"
	"gpxtoolkit/gpx"
	"html/template"
	"math"
	"time"

	"google.golang.org/protobuf/proto"
)

type Marker struct {
	Distance     float64
	NameTemplate string
	Symbol       string
}

func (m *Marker) Mark(log *gpx.TrackLog) error {
	tmpl, err := template.New("").Parse(m.NameTemplate)
	if err != nil {
		return err
	}
	for _, t := range log.Tracks {
		for _, s := range t.Segments {
			markers, err := m.mark(tmpl, s.Points)
			if err != nil {
				return err
			}
			log.WayPoints = append(log.WayPoints, markers...)
		}
	}
	return nil
}

func (m *Marker) mark(tmpl *template.Template, points []*gpx.Point) ([]*gpx.WayPoint, error) {
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
					err := tmpl.Execute(&buf, payload)
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
