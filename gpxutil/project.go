package gpxutil

import (
	"fmt"
	"gpxtoolkit/gpx"
	"log"
)

type ProjectWaypoints struct {
	Threshold struct {
		Inclusive float64
		Exclusive float64
	}
}

func (c *ProjectWaypoints) Name() string {
	return fmt.Sprintf("Project Waypoints with Threshold [%f,%f] m", c.Threshold.Inclusive, c.Threshold.Exclusive)
}

func (c *ProjectWaypoints) Run(tracklog *gpx.TrackLog) (int, error) {
	points := make([]*gpx.Point, 0)
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			points = append(points, seg.Points...)
		}
	}
	projections, err := projectWaypoints(points, tracklog.WayPoints, c.Threshold.Inclusive, c.Threshold.Exclusive)
	if err != nil {
		return 0, err
	}
	n := 0
	for i, p := range projections {
		wpt := tracklog.WayPoints[i]
		if p.point == nil {
			log.Printf("No projection: %s", wpt.GetName())
			continue
		}
		wpt.Latitude = p.point.Latitude
		wpt.Longitude = p.point.Longitude
		wpt.Elevation = p.point.Elevation
		wpt.NanoTime = p.point.NanoTime
		n++
	}
	return n, nil
}

type projection struct {
	waypoint *gpx.WayPoint
	point    *gpx.Point
	line     *line
	distance float64
}

type projections []*projection

type segment struct {
	a, b struct {
		waypoint *gpx.WayPoint
		point    *gpx.Point
	}
	points []*gpx.Point
}

func (projections projections) slice(points []*gpx.Point) []*segment {
	segments := make([]*segment, 0)
	seg := &segment{
		points: make([]*gpx.Point, 0),
	}
	for i, b := range points[1:] {
		a := points[i]
		seg.points = append(seg.points, a)
		for _, prj := range projections {
			if prj.point == nil {
				continue
			}
			if prj.line.a == a && prj.line.b == b {
				seg.points = append(seg.points, prj.point)
				seg.b = struct {
					waypoint *gpx.WayPoint
					point    *gpx.Point
				}{
					waypoint: prj.waypoint,
					point:    prj.point,
				}
				segments = append(segments, seg)
				seg = &segment{
					a: struct {
						waypoint *gpx.WayPoint
						point    *gpx.Point
					}{
						waypoint: prj.waypoint,
						point:    prj.point,
					},
					points: make([]*gpx.Point, 0),
				}
				seg.points = append(seg.points, prj.point)
			}
		}
	}
	seg.points = append(seg.points, points[len(points)-1])
	segments = append(segments, seg)
	log.Printf("Sliced %d segments", len(segments))
	for i, seg := range segments {
		a := "start"
		if seg.a.waypoint != nil {
			a = seg.a.waypoint.GetName()
		}
		b := "end"
		if seg.b.waypoint != nil {
			b = seg.b.waypoint.GetName()
		}
		log.Printf("Segment %d: from %s to %s => %d points", i, a, b, len(seg.points))
	}
	return segments
}

func projectWaypoints(points []*gpx.Point, waypoints []*gpx.WayPoint, inclusive, exclusive float64) (projections, error) {
	lines := getLines(points)
	projections := make(projections, len(waypoints))
	for i, w := range waypoints {
		prj := &projection{}
		projections[i] = prj
		p := w.GetPoint()
		for j, l := range lines {
			pp := l.project(p)
			var dist float64
			if pp == nil {
				d1 := p.DistanceTo(l.a)
				d2 := p.DistanceTo(l.b)
				if d1 <= inclusive {
					if d2 < d1 {
						pp = l.b
						dist = d2
					} else {
						pp = l.a
						dist = d1
					}
				} else if d2 <= inclusive {
					pp = l.b
					dist = d2
				} else {
					continue
				}
			} else {
				dist = p.DistanceTo(pp)
			}
			if dist > exclusive {
				continue
			}
			// d1 := p.DistanceTo(l.a)
			// d2 := p.DistanceTo(l.b)
			lat1 := l.a.GetLatitude()
			lat2 := l.b.GetLatitude()
			lon1 := l.a.GetLongitude()
			lon2 := l.b.GetLongitude()
			// // log.Printf("Distance from '%s' to (%f,%f):(%f:%f): %f", w.GetName(), lat1, lon1, lat2, lon2, dist)
			if prj.point == nil || dist < prj.distance {
				log.Printf("Shortest distance from '%s' to line[%d] (%f,%f):(%f:%f): %f", w.GetName(), j, lat1, lon1, lat2, lon2, dist)
				prj.point = pp
				prj.waypoint = w
				prj.line = l
				prj.distance = dist
			}
		}
	}
	return projections, nil
}
