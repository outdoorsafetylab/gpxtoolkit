package gpxutil

import (
	"fmt"
	"gpxtoolkit/gpx"
	"time"

	"google.golang.org/protobuf/proto"
)

type TimeShift struct {
	Duration time.Duration
}

func (c *TimeShift) Name() string {
	return fmt.Sprintf("Shift timestamp of waypoints and track points by %v", c.Duration)
}

func (c *TimeShift) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			for _, p := range seg.Points {
				if p.NanoTime == nil {
					continue
				}
				p.NanoTime = proto.Int64(p.Time().Add(c.Duration).UnixNano())
				n++
			}
		}
	}
	for _, p := range tracklog.WayPoints {
		if p.NanoTime == nil {
			continue
		}
		p.NanoTime = proto.Int64(p.Time().Add(c.Duration).UnixNano())
		n++
	}
	return n, nil
}
