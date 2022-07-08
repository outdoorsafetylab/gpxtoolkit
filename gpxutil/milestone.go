package gpxutil

import (
	"bytes"
	"fmt"
	"gpxtoolkit/gpx"
	"log"
	"math"
	"text/template"

	"google.golang.org/protobuf/proto"
)

type Milestone struct {
	Distance     float64
	Template     *template.Template
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
	waypoint *gpx.WayPoint
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
		name, err := c.milestoneName(i, float64(i-1)*c.Distance)
		if err != nil {
			return nil, err
		}
		milestones[i] = &milestone{
			name:     name,
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
			log.Printf("Segment %d: %d points", i, len(segment.points))
			n += len(segment.points)
		}
		log.Printf("Total %d points", n)
		markers := make([]*gpx.WayPoint, 0)
		end := 0.0
		for i, segment := range segments {
			start := end
			distances := make([]float64, len(segment.points)-1)
			for j, b := range segment.points[1:] {
				a := segment.points[j]
				dist := a.DistanceTo(b)
				distances[j] = dist
				// log.Printf("Distance %d: %f", j, dist)
				end += dist
			}
			num := int(math.Round(end/c.Distance) - math.Round(start/c.Distance))
			length := (end - start)
			log.Printf("Segment %d: %f meters (from %f to %f) with %d milestones", i, length, start, end, num)
			step := length / float64(num)
			milestones := make([]*milestone, num)
			for j := range milestones {
				index := int(math.Round(start/c.Distance)) + j
				name, err := c.milestoneName(index, float64(index+1)*c.Distance)
				if err != nil {
					return nil, err
				}
				milestones[j] = &milestone{
					name:     name,
					distance: float64(j+1) * step,
				}
			}
			if num > 1 {
				milestones[num-1].waypoint = segment.b.waypoint
			}
			m, err := c.create(segment.points, milestones, distances)
			if err != nil {
				return nil, err
			}
			markers = append(markers, m...)
			if i == 7 {
				break
			}
		}
		return markers, nil
	} else {
		return c.create(points, milestones, distances)
	}
}

func (c *Milestone) milestoneName(index int, meter float64) (string, error) {
	data := &struct {
		Index     int
		Number    int
		Meter     float64
		Kilometer float64
	}{
		Index: index,
	}
	data.Number = data.Index + 1
	data.Meter = meter
	data.Kilometer = data.Meter / 1000
	var buf bytes.Buffer
	err := c.Template.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
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
			if ms.waypoint != nil {
				if ms.waypoint.Name != nil {
					ms.waypoint.Name = proto.String(fmt.Sprintf("%s/%s", ms.waypoint.GetName(), ms.name))
				} else {
					ms.waypoint.Name = proto.String(ms.name)
				}
			} else {
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
		}
		start += dist
	}
	return markers, nil
}
