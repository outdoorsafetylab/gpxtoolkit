package gpxutil

import (
	"fmt"
	"gpxtoolkit/gpx"
	"log"
)

type ProjectWaypoints struct {
	Threshold float64
}

func (c *ProjectWaypoints) Name() string {
	return fmt.Sprintf("Project Waypoints with Threshold %f m", c.Threshold)
}

func (c *ProjectWaypoints) Run(tracklog *gpx.TrackLog) (int, error) {
	points := make([]*gpx.Point, 0)
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			points = append(points, seg.Points...)
		}
	}
	projections, err := projectWaypoints(points, tracklog.WayPoints, c.Threshold)
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
		lat1 := wpt.GetLatitude()
		lat2 := wpt.GetLatitude()
		lon1 := p.point.GetLongitude()
		lon2 := p.point.GetLongitude()
		log.Printf("Projecting %s from (%f,%f) to (%f:%f) => %f m", wpt.GetName(), lat1, lon1, lat2, lon2, p.distance)
		wpt.Latitude = p.point.Latitude
		wpt.Longitude = p.point.Longitude
		wpt.Elevation = p.point.Elevation
		wpt.NanoTime = p.point.NanoTime
		n++
	}
	return n, nil
}

type projection struct {
	point    *gpx.Point
	line     *line
	distance float64
}

type projections []*projection

func (projections projections) slice(points []*gpx.Point) [][]*gpx.Point {
	slices := make([][]*gpx.Point, 0)
	slice := make([]*gpx.Point, 0)
	for i, b := range points[1:] {
		a := points[i]
		for _, prj := range projections {
			if prj.point == nil {
				continue
			}
			if prj.line.p1 == a && prj.line.p2 == b {
				slice = append(slice, prj.point)
				slices = append(slices, slice)
				slice = make([]*gpx.Point, 0)
			}
		}
		slice = append(slice, a)
	}
	slice = append(slice, points[len(points)-1])
	slices = append(slices, slice)
	return slices
}

func projectWaypoints(points []*gpx.Point, waypoints []*gpx.WayPoint, threshold float64) (projections, error) {
	lines := getLines(points)
	projections := make(projections, len(waypoints))
	for i, w := range waypoints {
		prj := &projection{}
		projections[i] = prj
		p := w.GetPoint()
		for _, l := range lines {
			pp := l.project(p)
			if pp == nil {
				continue
			}
			dist := p.DistanceTo(pp)
			if dist > threshold {
				continue
			}
			// d1 := p.DistanceTo(l.p1)
			// d2 := p.DistanceTo(l.p2)
			// lat1 := l.p1.GetLatitude()
			// lat2 := l.p2.GetLatitude()
			// lon1 := l.p1.GetLongitude()
			// lon2 := l.p2.GetLongitude()
			// // log.Printf("Distance from '%s' to (%f,%f):(%f:%f): %f", w.GetName(), lat1, lon1, lat2, lon2, dist)
			if prj.point == nil || dist < prj.distance {
				// log.Printf("New shortest distance: %f", dist)
				// log.Printf("Shortest distance from '%s' to line[%d] (%f,%f):(%f:%f): %f", w.GetName(), j, lat1, lon1, lat2, lon2, dist)
				prj.point = pp
				prj.line = l
				prj.distance = dist
			}
		}
	}
	return projections, nil
}