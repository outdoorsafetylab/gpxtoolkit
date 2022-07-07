package milestone

import (
	"bytes"
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"log"
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
			points := s.Points
			if m.Reverse {
				log.Printf("Reverse processing trk[%d]/trkseg[%d]", i, j)
				n := len(s.Points)
				points = make([]*gpx.Point, n)
				for i, p := range s.Points {
					points[n-1-i] = p
				}
			} else {
				log.Printf("Forward processing trk[%d]/trkseg[%d]", i, j)
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
	if len(points) <= 0 {
		return []*gpx.WayPoint{}, nil
	}
	distances := make([]float64, len(points)-1)
	total := 0.0
	for i, b := range points[1:] {
		a := points[i]
		dist := a.DistanceTo(b)
		distances[i] = dist
		total += dist
	}
	milestones := make([]*milestone, int(math.Floor(total/m.Distance)))
	for i, _ := range milestones {
		payload := &templatePayload{
			Index:  i,
			Number: i + 1,
		}
		payload.Number = payload.Index + 1
		payload.Meter = float64(payload.Number) * m.Distance
		payload.Kilometer = payload.Meter / 1000
		var buf bytes.Buffer
		err := m.NameTemplate.Execute(&buf, payload)
		if err != nil {
			return nil, err
		}
		milestones[i] = &milestone{
			name:     buf.String(),
			distance: float64(i) * m.Distance,
		}
	}
	markers := make([]*gpx.WayPoint, 0)
	start := 0.0
	for i, b := range points[1:] {
		a := points[i]
		dist := distances[i]
		end := start + dist
		for _, ms := range milestones {
			if start > ms.distance || end <= ms.distance {
				continue
			}
			p := interpolate(a, b, (ms.distance-start)/dist)
			wpt := &gpx.WayPoint{
				Name:      proto.String(ms.name),
				Latitude:  p.Latitude,
				Longitude: p.Longitude,
				NanoTime:  p.NanoTime,
				Elevation: p.Elevation,
			}
			if m.Symbol != "" {
				wpt.Symbol = proto.String(m.Symbol)
			}
			markers = append(markers, wpt)
		}
		start += dist
	}
	return markers, nil
}

func interpolate(a, b *gpx.Point, ratio float64) *gpx.Point {
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
	lat := lat1 + dlat*ratio
	lon := lon1 + dlon*ratio
	res := &gpx.Point{
		Latitude:  proto.Float64(lat),
		Longitude: proto.Float64(lon),
	}
	if a.Elevation != nil && b.Elevation != nil {
		res.Elevation = proto.Float64(ele1 + dele*ratio)
	}
	if a.NanoTime != nil && b.NanoTime != nil {
		res.NanoTime = proto.Int64(t1.Add(dt * time.Duration(ratio)).UnixNano())
	}
	return res
}

func (m *Marker) marks2(points []*gpx.Point) ([]*gpx.WayPoint, error) {
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
