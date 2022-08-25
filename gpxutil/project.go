package gpxutil

import (
	"fmt"
	"sort"

	"gpxtoolkit/gpx"
	"gpxtoolkit/log"

	"google.golang.org/protobuf/proto"
)

type ProjectWaypoints struct {
	DistanceFunc DistanceFunc
	Threshold    float64
	KeepOriginal bool
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
	lines := getLines(c.DistanceFunc, points)
	projections := projectWaypoints(c.DistanceFunc, lines, tracklog.WayPoints, c.Threshold)
	n := 0
	for i, p := range projections {
		wpt := tracklog.WayPoints[i]
		if p.point == nil {
			log.Infof("No projection: %s", wpt.GetName())
			continue
		}
		if c.KeepOriginal {
			tracklog.WayPoints = append(tracklog.WayPoints, &gpx.WayPoint{
				Name:      proto.String(fmt.Sprintf("%s'", wpt.GetName())),
				Latitude:  p.point.Latitude,
				Longitude: p.point.Longitude,
				Elevation: p.point.Elevation,
			})
		} else {
			wpt.Latitude = p.point.Latitude
			wpt.Longitude = p.point.Longitude
			wpt.Elevation = p.point.Elevation
			wpt.NanoTime = p.point.NanoTime
		}
		n++
	}
	return n, nil
}

type projection struct {
	waypoint       *gpx.WayPoint
	point          *gpx.Point
	line           *line
	distanceToLine float64
	mileage        float64
}

type projections []*projection

type segment struct {
	a, b struct {
		waypoint *gpx.WayPoint
		point    *gpx.Point
	}
	points []*gpx.Point
}

func projectWaypoints(distanceFunc DistanceFunc, lines []*line, waypoints []*gpx.WayPoint, threshold float64) projections {
	projections := make(projections, 0)
	for _, w := range waypoints {
		p := w.GetPoint()
		mileage := 0.0
		var closest *projection
		for _, l := range lines {
			mileage += l.dist
			pp := l.closestPoint(p)
			dist := HaversinDistance(p, pp)
			if threshold > 0 && dist > threshold {
				if closest != nil {
					projections = append(projections, closest)
					closest = nil
				}
				continue
			}
			if closest == nil || dist < closest.distanceToLine {
				prj := &projection{}
				prj.point = pp
				prj.waypoint = w
				prj.line = l
				prj.distanceToLine = dist
				prj.mileage = mileage + distanceFunc(l.a, prj.point)
				closest = prj
			}
		}
		if closest != nil {
			projections = append(projections, closest)
		}
	}
	sort.Slice(projections, func(i, j int) bool {
		return projections[i].mileage < projections[j].mileage
	})
	return projections
}

func sliceByWaypoints(distanceFunc DistanceFunc, points []*gpx.Point, waypoints []*gpx.WayPoint, threshold float64) ([]*segment, error) {
	lines := getLines(distanceFunc, points)
	projections := projectWaypoints(distanceFunc, lines, waypoints, threshold)
	segments := make([]*segment, 0)
	seg := &segment{
		points: make([]*gpx.Point, 0),
	}
	for _, prj := range projections {
		for i, l := range lines {
			seg.points = append(seg.points, l.a)
			if prj.line == l {
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
				seg.points = append(seg.points, l.b)
				lines = lines[i+1:]
				break
			}
		}
	}
	for i, l := range lines {
		seg.points = append(seg.points, l.a)
		if i == len(lines)-1 {
			seg.points = append(seg.points, l.b)
		}
	}
	segments = append(segments, seg)
	log.Debugf("Sliced %d segments", len(segments))
	return segments, nil
}
