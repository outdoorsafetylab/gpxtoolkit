package gpxutil

import (
	"gpxtoolkit/gpx"
	"math"
	"time"
)

type DistanceFunc func(a, b *gpx.Point) float64

func HaversinDistance(a, b *gpx.Point) float64 {
	return gpx.GeoDistance(a.GetLatitude(), a.GetLongitude(), b.GetLatitude(), b.GetLongitude())
}

func TerrainDistance(a, b *gpx.Point) float64 {
	// FIXME: might be another better solution...
	h := HaversinDistance(a, b)
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

func (l *line) closestPoint(c *gpx.Point) *gpx.Point {
	xa, ya := toWebMercator(l.a.GetLatitude(), l.a.GetLongitude())
	xb, yb := toWebMercator(l.b.GetLatitude(), l.b.GetLongitude())
	xc, yc := toWebMercator(c.GetLatitude(), c.GetLongitude())
	ab := &vector{
		x: xb - xa,
		y: yb - ya,
	}
	ac := &vector{
		x: xc - xa,
		y: yc - ya,
	}
	dotProduct := ab.x*ac.x + ab.y*ac.y
	abLengthSquare := ab.x*ab.x + ab.y*ab.y
	if abLengthSquare == 0 {
		return l.a
	}
	ratio := dotProduct / abLengthSquare
	if ratio <= 0 {
		return l.a
	} else if ratio >= 1 {
		return l.b
	}
	return interpolate(l.a, l.b, ratio)
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
				// } else {
				// log.Printf("Line[%d]: zero duration", i)
			}
			// } else {
			// log.Printf("Line[%d]: missing time", i)
		}
		lines[i] = line
	}
	return lines
}

func joinLines(lines []*line) []*gpx.Point {
	points := make([]*gpx.Point, 0)
	for _, line := range lines {
		points = append(points, line.a)
	}
	last := lines[len(lines)-1]
	points = append(points, last.b)
	return points
}
