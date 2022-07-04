package milestone

import (
	"bytes"
	"gpxtoolkit/gpx"
	"log"
	"math"
	"text/template"
	"time"

	"google.golang.org/protobuf/proto"
)

// func (m *Marker) expand(points []*gpx.Point, waypoints []*gpx.WayPoint) []*gpx.Point {
// 	lines := make([]*line, len(points)-1)
// 	for i, p := range points {
// 		if i < len(lines) {
// 			lines[i] = &line{
// 				p1: p,
// 				p2: points[i+1],
// 			}
// 		}
// 	}
// 	threshold := m.Distance / 2
// 	closestLines := make([]struct {
// 		waypoint *gpx.WayPoint
// 		line     *line
// 		dist1    float64
// 		dist2    float64
// 	}, len(waypoints))
// 	for i, w := range waypoints {
// 		closestLines[i].waypoint = w
// 		p := w.GetPoint()
// 		minDist := 0.0
// 		for _, l := range lines {
// 			d1 := p.DistanceTo(l.p1)
// 			d2 := p.DistanceTo(l.p2)
// 			if d1 > threshold && d2 > threshold {
// 				continue
// 			}
// 			dist := l.virtualDistanceTo(p)
// 			if minDist <= 0 || dist < minDist {
// 				minDist = dist
// 				closestLines[i].line = l
// 			}
// 		}
// 	}
// 	res := make([]*gpx.Point, 0)
// 	for i, l := range lines {
// 		res = append(res, l.p1)
// 		for _, cl := range closestLines {
// 			if l != cl.line {
// 				continue
// 			}
// 			w := cl.waypoint
// 			log.Printf("Expanding with waypoint: %s", w.GetName())
// 			res = append(res, w.GetPoint())
// 		}
// 		if i == len(lines)-1 {
// 			res = append(res, l.p2)
// 		}
// 	}
// 	return res
// }

type gpxPoints []*gpx.Point

func (points gpxPoints) toLines() []*line {
	lines := make([]*line, 0)
	for i, a := range points[:len(points)-1] {
		b := points[i+1]
		dist := a.DistanceTo(b)
		if dist <= 0 {
			continue
		}
		lines = append(lines, &line{
			p1: a,
			p2: b,
		})
	}
	log.Printf("Converted %d points to %d lines.", len(points), len(lines))
	return lines
}

func (points gpxPoints) projectWaypoints(waypoints []*gpx.WayPoint) []*waypointProjection {
	lines := points.toLines()
	projections := make([]*waypointProjection, 0)
	log.Printf("Projecting %d waypoints to %d lines...", len(waypoints), len(lines))
	for _, w := range waypoints {
		prj := &waypointProjection{
			waypoint: w,
		}
		p := w.GetPoint()
		for j, l := range lines {
			pp := l.project(p)
			if pp == nil {
				continue
			}
			dist := p.DistanceTo(pp)
			// d1 := p.DistanceTo(l.p1)
			// d2 := p.DistanceTo(l.p2)
			lat1 := l.p1.GetLatitude()
			lat2 := l.p2.GetLatitude()
			lon1 := l.p1.GetLongitude()
			lon2 := l.p2.GetLongitude()
			// // log.Printf("Distance from '%s' to (%f,%f):(%f:%f): %f", w.GetName(), lat1, lon1, lat2, lon2, dist)
			if prj.projection == nil || dist < prj.distance {
				// log.Printf("New shortest distance: %f", dist)
				log.Printf("Shortest distance from '%s' to line[%d] (%f,%f):(%f:%f): %f", w.GetName(), j, lat1, lon1, lat2, lon2, dist)
				prj.start = l.p1
				prj.end = l.p2
				prj.projection = pp
				prj.distance = dist
			}
		}
		if prj.projection == nil {
			continue
		}
		projections = append(projections, prj)
	}
	return projections
}

func (points gpxPoints) expand(projections []*waypointProjection) gpxPoints {
	res := make(gpxPoints, 0)
	for _, p := range points {
		res = append(res, p)
		for _, prj := range projections {
			if p == prj.start {
				res = append(res, prj.projection)
			}
		}
	}
	return res
}

type segment struct {
	points   gpxPoints
	total    float64
	distance float64
	waypoint *gpx.WayPoint
}

func (points gpxPoints) segments(projections []*waypointProjection, distance float64) []*segment {
	segments := make([]*segment, 0)
	seg := &segment{
		points: make(gpxPoints, 0),
	}
	expaned := points.expand(projections)
	total := float64(0)
	for i, b := range expaned[1:] {
		a := expaned[i]
		total += a.DistanceTo(b)
		seg.points = append(seg.points, a)
		if total < distance {
			continue
		}
		for _, prj := range projections {
			if b != prj.projection {
				continue
			}
			seg.total = total
			seg.distance = total / math.Round(total/distance)
			log.Printf("Hit waypoint[%d]: %s: distance=%f", i, prj.waypoint.GetName(), seg.distance)
			seg.waypoint = prj.waypoint
			segments = append(segments, seg)
			seg = &segment{
				points: make(gpxPoints, 0),
			}
			total = 0
			break
		}
	}
	seg.distance = total / math.Round(total/distance)
	segments = append(segments, seg)
	return segments
}

func (points gpxPoints) marks(baseDistance, realDistance, targetDistance float64, nameTemplate *template.Template, symbol string) ([]*gpx.WayPoint, error) {
	res := make([]*gpx.WayPoint, 0)
	remainder := float64(0)
	for i, b := range points[1:] {
		a := points[i]
		dist := b.DistanceTo(a)
		if (remainder + dist) >= realDistance {
			first := (realDistance - remainder)
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
				payload.Meter = baseDistance + float64(payload.Number)*targetDistance
				payload.Kilometer = payload.Meter / 1000
				var buf bytes.Buffer
				err := nameTemplate.Execute(&buf, payload)
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
				if symbol != "" {
					wpt.Symbol = proto.String(symbol)
				}
				res = append(res, wpt)
				pos += realDistance
			}
			remainder = remainder + dist - realDistance
		} else {
			remainder += dist
		}
	}
	return res, nil
}

type templatePayload struct {
	Index     int
	Number    int
	Meter     float64
	Kilometer float64
}

type line struct {
	p1 *gpx.Point
	p2 *gpx.Point
}

// https://stackoverflow.com/a/6853926
func (l *line) project(p *gpx.Point) *gpx.Point {
	x := p.GetLatitude()
	y := p.GetLongitude()
	x1 := l.p1.GetLatitude()
	y1 := l.p1.GetLongitude()
	x2 := l.p2.GetLatitude()
	y2 := l.p2.GetLongitude()
	A := x - x1
	B := y - y1
	C := x2 - x1
	D := y2 - y1

	dot := A*C + B*D
	len_sq := C*C + D*D
	param := -1.0
	if len_sq != 0 { //in case of 0 length line
		param = dot / len_sq
	}

	if param > 0 && param <= 1 {
		t1 := l.p1.GetNanoTime()
		t2 := l.p2.GetNanoTime()
		dt := t2 - t1

		lat1 := l.p1.GetLatitude()
		lat2 := l.p2.GetLatitude()
		dlat := lat2 - lat1

		lon1 := l.p1.GetLongitude()
		lon2 := l.p2.GetLongitude()
		dlon := lon2 - lon1

		ele1 := l.p1.GetElevation()
		ele2 := l.p2.GetElevation()
		dele := ele2 - ele1

		dist1 := p.DistanceTo(l.p1)
		dist2 := p.DistanceTo(l.p2)
		dist := dist1 + dist2
		interpolation := dist1 / dist
		return &gpx.Point{
			NanoTime:  proto.Int64(t1 + int64(float64(dt)*interpolation)),
			Latitude:  proto.Float64(lat1 + (dlat * interpolation)),
			Longitude: proto.Float64(lon1 + (dlon * interpolation)),
			Elevation: proto.Float64(ele1 + (dele * interpolation)),
		}
	} else {
		return nil
	}
}

type waypointProjection struct {
	waypoint   *gpx.WayPoint
	start      *gpx.Point
	end        *gpx.Point
	projection *gpx.Point
	distance   float64
}
