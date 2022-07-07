package gpxutil

import (
	"gpxtoolkit/gpx"
	"time"
)

type line struct {
	p1       *gpx.Point
	p2       *gpx.Point
	dist     float64
	duration *time.Duration
	speed    *float64
}

func getLines(points []*gpx.Point) []*line {
	if len(points) <= 1 {
		return []*line{}
	}
	lines := make([]*line, len(points)-1)
	for i, p := range points[1:] {
		line := &line{
			p1: points[i],
			p2: p,
		}
		line.dist = line.p1.DistanceTo(line.p2)
		// log.Printf("Line[%d]: dist=%f", i, line.dist)
		if line.p1.NanoTime != nil && line.p2.NanoTime != nil {
			line.duration = new(time.Duration)
			*line.duration = line.p2.Time().Sub(line.p1.Time())

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
		if i == 0 || points[len(points)-1] != line.p1 {
			points = append(points, line.p1)
		}
		if points[len(points)-1] != line.p2 {
			points = append(points, line.p2)
		}
	}
	return points
}
