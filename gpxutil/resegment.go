package gpxutil

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"

	"google.golang.org/protobuf/proto"
)

type ReSegment struct {
	Service      elevation.Service
	DistanceFunc DistanceFunc
	Threshold    float64
}

func (c *ReSegment) Name() string {
	return fmt.Sprintf("Re-Segment by Waypoints with Threshold %fm", c.Threshold)
}

func (c *ReSegment) Run(tracklog *gpx.TrackLog) (int, error) {
	points := make([]*gpx.Point, 0)
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			points = append(points, seg.Points...)
		}
	}
	projections, err := projectWaypoints(c.DistanceFunc, points, tracklog.WayPoints, c.Threshold, c.Service)
	if err != nil {
		return 0, err
	}
	segments := projections.slice(points)
	tracklog.Tracks = make([]*gpx.Track, len(segments))
	for i, seg := range segments {
		name := ""
		if seg.a.waypoint != nil {
			name += seg.a.waypoint.GetName()
		}
		name += "→"
		if seg.b.waypoint != nil {
			name += seg.b.waypoint.GetName()
		}
		track := &gpx.Track{
			Name: proto.String(name),
			Segments: []*gpx.Segment{
				{Points: seg.points},
			},
		}
		tracklog.Tracks[i] = track
	}
	return len(points), nil
}
