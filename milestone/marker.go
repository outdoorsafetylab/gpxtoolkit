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
				dele := float64(-1)
				if a.Elevation != nil && b.Elevation != nil {
					dele = ele2 - ele1
				}
				t1 := a.Time()
				t2 := b.Time()
				dt := time.Duration(-1)
				if a.NanoTime != nil && b.NanoTime != nil {
					dt = t2.Sub(t1)
				}
				pos := first
				for pos < dist {
					ratio := pos / dist
					lat := lat1 + dlat*ratio
					lon := lon1 + dlon*ratio
					payload := &templatePayload{
						Index:  len(res),
						Number: len(res) + 1,
						Meter:  float64(len(res)+1) * m.Distance,
					}
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
					if dele > 0 {
						wpt.Elevation = proto.Float64(ele1 + dele*ratio)
					}
					if dt > 0 {
						wpt.NanoTime = proto.Int64(t1.Add(dt * time.Duration(ratio)).UnixNano())
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
	Index  int
	Number int
	Meter  float64
}
