package gpxutil

import (
	"fmt"
	"os"
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
	projections, err := projectWaypoints(c.DistanceFunc, points, tracklog.WayPoints, c.Threshold)
	if err != nil {
		return 0, err
	}
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
	log.Debugf("Sliced %d segments", len(segments))
	return segments
}

func projectWaypoints(distanceFunc DistanceFunc, points []*gpx.Point, waypoints []*gpx.WayPoint, threshold float64) (projections, error) {
	lines := getLines(distanceFunc, points)
	projections := make(projections, len(waypoints))
	for i, w := range waypoints {
		prj := &projection{}
		projections[i] = prj
		p := w.GetPoint()
		mileage := 0.0
		for _, l := range lines {
			mileage += l.dist
			pp := l.closestPoint(p)
			dist := HaversinDistance(p, pp)
			if threshold > 0 && dist > threshold {
				continue
			}
			if prj.point == nil || dist < prj.distanceToLine {
				prj.point = pp
				prj.waypoint = w
				prj.line = l
				prj.distanceToLine = dist
				prj.mileage = mileage + distanceFunc(l.a, prj.point)
			}
		}
	}
	return projections, nil
}

func sliceByWaypoints(distanceFunc DistanceFunc, points []*gpx.Point, waypoints []*gpx.WayPoint, threshold float64) ([]*segment, error) {
	lines := getLines(distanceFunc, points)
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
	for _, prj := range projections {
		if prj.point == nil {
			continue
		}
		fmt.Fprintf(os.Stderr, "%s: %f\n", prj.waypoint.GetName(), prj.mileage)
	}
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
				lines = lines[i:]
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
	fmt.Fprintf(os.Stderr, "Sliced %d points to %d segments\n", len(points), len(segments))
	for i, s := range segments {
		fmt.Fprintf(os.Stderr, "Segment %d: %d points\n", i, len(s.points))
	}
	return segments, nil
}
