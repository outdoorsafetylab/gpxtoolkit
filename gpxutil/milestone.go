package gpxutil

import (
	"bytes"
	"gpxtoolkit/gpx"
	"log"
	"math"
	"text/template"

	"google.golang.org/protobuf/proto"
)

type Milestone struct {
	Distance     float64
	NameTemplate *template.Template
	Symbol       string
	Reverse      bool
	FitWaypoints bool
}

func (c *Milestone) Name() string {
	return "Create Milestones"
}

func (c *Milestone) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			points := seg.Points
			if c.Reverse {
				points = make([]*gpx.Point, len(seg.Points))
				for i, p := range seg.Points {
					points[len(seg.Points)-1-i] = p
				}
			}
			waypoints := tracklog.WayPoints
			if !c.FitWaypoints {
				waypoints = nil
			}
			milestones, err := c.milestone(points, waypoints)
			if err != nil {
				return 0, err
			}
			n += len(milestones)
			log.Printf("Appending %d milestones", len(milestones))
			tracklog.WayPoints = append(tracklog.WayPoints, milestones...)
		}
	}
	return n, nil
}

type milestone struct {
	name     string
	distance float64
}

func (c *Milestone) milestone(points []*gpx.Point, waypoints []*gpx.WayPoint) ([]*gpx.WayPoint, error) {
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
	milestones := make([]*milestone, int(math.Floor(total/c.Distance)))
	for i := range milestones {
		payload := &struct {
			Index     int
			Number    int
			Meter     float64
			Kilometer float64
		}{
			Index: i,
		}
		payload.Number = payload.Index + 1
		payload.Meter = float64(payload.Number) * c.Distance
		payload.Kilometer = payload.Meter / 1000
		var buf bytes.Buffer
		err := c.NameTemplate.Execute(&buf, payload)
		if err != nil {
			return nil, err
		}
		milestones[i] = &milestone{
			name:     buf.String(),
			distance: float64(i+1) * c.Distance,
		}
		log.Printf("Milestone %d: %s @ %f m", i+1, milestones[i].name, milestones[i].distance)
	}
	if waypoints != nil {
		projections, err := projectWaypoints(points, waypoints, c.Distance/2)
		if err != nil {
			return nil, err
		}
		segments := projections.slice(points)
		log.Printf("Sliced %d points to %d segments", len(points), len(segments))
		n := 0
		for i, segment := range segments {
			log.Printf("Segment %d: %d points", i, len(segment))
			n += len(segment)
		}
		log.Printf("Total %d points", n)
		markers := make([]*gpx.WayPoint, 0)
		end := 0.0
		for i, segment := range segments {
			start := end
			distances := make([]float64, len(segment)-1)
			for j, b := range segment[1:] {
				a := segment[j]
				dist := a.DistanceTo(b)
				distances[j] = dist
				log.Printf("Distance %d: %f", j, dist)
				end += dist
			}
			num := int(math.Round(end/c.Distance) - math.Round(start/c.Distance))
			length := (end - start)
			log.Printf("Segment %d: %f meters (from %f to %f) with %d milestones", i, length, start, end, num)
			step := length / float64(num)
			milestones := make([]*milestone, num)
			for j := range milestones {
				payload := &struct {
					Index     int
					Number    int
					Meter     float64
					Kilometer float64
				}{
					Index: int(math.Round(start/c.Distance)) + j,
				}
				payload.Number = payload.Index + 1
				payload.Meter = float64(payload.Number) * c.Distance
				payload.Kilometer = payload.Meter / 1000
				var buf bytes.Buffer
				err := c.NameTemplate.Execute(&buf, payload)
				if err != nil {
					return nil, err
				}
				milestones[j] = &milestone{
					name:     buf.String(),
					distance: float64(j+1) * step,
				}
			}
			m, err := c.create(segment, milestones, distances)
			if err != nil {
				return nil, err
			}
			markers = append(markers, m...)
		}
		return markers, nil
	} else {
		return c.create(points, milestones, distances)
	}
}

func (c *Milestone) create(points []*gpx.Point, milestones []*milestone, distances []float64) ([]*gpx.WayPoint, error) {
	log.Printf("Creating %d milestones from %d points ", len(milestones), len(points))
	markers := make([]*gpx.WayPoint, 0)
	start := 0.0
	for i, b := range points[1:] {
		a := points[i]
		var dist float64
		if distances != nil {
			dist = distances[i]
		} else {
			dist = a.DistanceTo(b)
		}
		end := start + dist
		// log.Printf("Current distance: %f", end)
		for _, ms := range milestones {
			// log.Printf("Milestone %d: %s @ %f", j, ms.name, ms.distance)
			if int(start*1000) >= int(ms.distance*1000) || int(end*1000) < int(ms.distance*1000) {
				continue
			}
			log.Printf("Hit milestone %s: %f", ms.name, ms.distance)
			p := Interpolate(a, b, (ms.distance-start)/dist)
			wpt := &gpx.WayPoint{
				Name:      proto.String(ms.name),
				Latitude:  p.Latitude,
				Longitude: p.Longitude,
				NanoTime:  p.NanoTime,
				Elevation: p.Elevation,
			}
			if c.Symbol != "" {
				wpt.Symbol = proto.String(c.Symbol)
			}
			markers = append(markers, wpt)
		}
		start += dist
	}
	return markers, nil
}