package gpxutil

import (
	"fmt"
	"gpxtoolkit/gpx"
	"time"

	"google.golang.org/protobuf/proto"
)

type ReTimestamp struct {
	DistanceFunc DistanceFunc
	Start        time.Time
	Speed        float64
}

func (c *ReTimestamp) Name() string {
	return fmt.Sprintf("Re-Timestamp from %v with Speed %f m/s", c.Start, c.Speed)
}

func (c *ReTimestamp) Run(tracklog *gpx.TrackLog) (int, error) {
	start := c.Start
	var err error
	n := 0
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			start, err = c.timestamp(seg.Points, start)
			if err != nil {
				return 0, err
			}
			n += len(seg.Points)
		}
	}
	return n, nil
}

func (c *ReTimestamp) timestamp(points []*gpx.Point, start time.Time) (time.Time, error) {
	lines := getLines(c.DistanceFunc, points)
	for i, line := range lines {
		line.a.NanoTime = proto.Int64(start.UnixNano())
		duration := time.Duration(line.dist/c.Speed) * time.Second
		start = start.Add(duration)
		if i == len(lines)-1 {
			line.b.NanoTime = proto.Int64(start.UnixNano())
		}
	}
	return start, nil
}
