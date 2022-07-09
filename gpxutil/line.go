package gpxutil

import (
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"math"
	"time"
)

type DistanceFunc func(a, b *gpx.Point) float64

var HorizontalDistance DistanceFunc = func(a, b *gpx.Point) float64 {
	return gpx.GeoDistance(a.GetLatitude(), a.GetLongitude(), b.GetLatitude(), b.GetLongitude())
}

var HypotenuseDistance DistanceFunc = func(a, b *gpx.Point) float64 {
	h := HorizontalDistance(a, b)
	v := a.GetElevation() - b.GetElevation()
	return math.Sqrt(h*h + v*v)
}

type line struct {
	a        *gpx.Point
	b        *gpx.Point
	dist     float64
	duration *time.Duration
	speed    *float64
}

// https://stackoverflow.com/a/6853926
func (l *line) closestPoint(distanceFunc DistanceFunc, p *gpx.Point, service elevation.Service) *gpx.Point {
	pp := &gpx.Point{
		Latitude:  l.a.Latitude,
		Longitude: l.b.Longitude,
	}
	A := distanceFunc(p, l.a)
	B := distanceFunc(p, l.b)
	C := distanceFunc(pp, l.b)
	D := distanceFunc(pp, l.a)

	dot := A*C + B*D
	len_sq := C*C + D*D
	param := -1.0
	if len_sq != 0 { //in case of 0 length line
		param = dot / len_sq
	}
	if param < 0 {
		return l.a
	} else if param > 0 {
		return l.b
	} else {
		dist1 := distanceFunc(p, l.a)
		dist2 := distanceFunc(p, l.b)
		dist := dist1 + dist2
		return interpolate(l.a, l.b, dist1/dist, service)
	}
	// x := p.GetLatitude()
	// y := p.GetLongitude()
	// x1 := l.a.GetLatitude()
	// y1 := l.a.GetLongitude()
	// x2 := l.b.GetLatitude()
	// y2 := l.b.GetLongitude()
	// A := x - x1
	// B := y - y1
	// C := x2 - x1
	// D := y2 - y1

	// dot := A*C + B*D
	// len_sq := C*C + D*D
	// param := -1.0
	// if len_sq != 0 { //in case of 0 length line
	// 	param = dot / len_sq
	// }

	// if param > 0 && param <= 1 {
	// 	dist1 := distanceFunc(p, l.a)
	// 	dist2 := distanceFunc(p, l.b)
	// 	dist := dist1 + dist2
	// 	return Interpolate(l.a, l.b, dist1/dist)
	// } else {
	// 	return nil
	// }
}

func getLines(distanceFunc DistanceFunc, points []*gpx.Point) []*line {
	if len(points) <= 1 {
		return []*line{}
	}
	lines := make([]*line, len(points)-1)
	for i, b := range points[1:] {
		a := points[i]
		line := &line{
			a: a,
			b: b,
		}
		line.dist = distanceFunc(line.a, line.b)
		// log.Printf("Line[%d]: dist=%f", i, line.dist)
		if line.a.NanoTime != nil && line.b.NanoTime != nil {
			line.duration = new(time.Duration)
			*line.duration = line.b.Time().Sub(line.a.Time())

			line.speed = new(float64)
			if *line.duration != 0 {
				*line.speed = line.dist / (*line.duration).Seconds()
				// log.Printf("Line[%d]: speed=%f", i, *line.speed)
			}
		}
		lines[i] = line
	}
	return lines
}

func joinLines(lines []*line) []*gpx.Point {
	points := make([]*gpx.Point, 0)
	for i, line := range lines {
		if i == 0 || points[len(points)-1] != line.a {
			points = append(points, line.a)
		}
		if points[len(points)-1] != line.b {
			points = append(points, line.b)
		}
	}
	return points
}
