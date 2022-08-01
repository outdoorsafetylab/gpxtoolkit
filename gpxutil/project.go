package gpxutil

import (
	"fmt"
	"gpxtoolkit/gpx"
	"log"
)

type ProjectWaypoints struct {
	DistanceFunc DistanceFunc
	Threshold    float64
}

func (c *ProjectWaypoints) Name() string {
	return fmt.Sprintf("Project Waypoints with Threshold %fm", c.Threshold)
}

func (c *ProjectWaypoints) Run(tracklog *gpx.TrackLog) (int, error) {
	points := make([]*gpx.Point, 0)
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			points = append(points, seg.Points...)
		}
	}
	projections, err := projectWaypoints(c.DistanceFunc, points, tracklog.WayPoints, c.Threshold)
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
	return segments
}

func projectWaypoints(distanceFunc DistanceFunc, points []*gpx.Point, waypoints []*gpx.WayPoint, threshold float64) (projections, error) {
	lines := getLines(distanceFunc, points)
	projections := make(projections, len(waypoints))
	for i, w := range waypoints {
		prj := &projection{}
		projections[i] = prj
		p := w.GetPoint()
		for _, l := range lines {
			pp := l.closestPoint(distanceFunc, p)
			dist := HaversinDistance(p, pp)
			if dist > threshold {
				continue
			}
			if prj.point == nil || dist < prj.distance {
				prj.point = pp
				prj.waypoint = w
				prj.line = l
				prj.distance = dist
			}
		}
	}
	return projections, nil
}
