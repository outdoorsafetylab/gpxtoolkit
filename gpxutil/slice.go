package gpxutil

import (
	"fmt"
	"gpxtoolkit/gpx"

	"google.golang.org/protobuf/proto"
)

type Slice struct {
	Start, End *gpx.WayPoint
	Points     []*gpx.Point
}

type SliceByWaypoints struct {
	DistanceFunc DistanceFunc
	Threshold    float64
	Waypoints    []*gpx.WayPoint
}

func (c *SliceByWaypoints) Name() string {
	return fmt.Sprintf("Slice by Waypoints with Threshold %fm", c.Threshold)
}

func (c *SliceByWaypoints) Run(tracklog *gpx.TrackLog) (int, error) {
	num := 0
	tracks := make([]*gpx.Track, 0)
	for _, t := range tracklog.Tracks {
		for _, s := range t.Segments {
			points := s.Points
			slices, err := c.slice(points)
			if err != nil {
				return 0, err
			}
			for _, slice := range slices {
				track := &gpx.Track{
					Segments: []*gpx.Segment{
						{Points: slice.Points},
					},
				}
				if slice.Start != nil {
					if slice.End != nil {
						track.Name = proto.String(fmt.Sprintf("%s→%s", slice.Start.GetName(), slice.End.GetName()))
					} else {
						track.Name = proto.String(fmt.Sprintf("%s→", slice.Start.GetName()))
					}
					//tracklog.WayPoints = append(tracklog.WayPoints, slice.Start)
				} else if slice.End != nil {
					track.Name = proto.String(fmt.Sprintf("→%s", slice.End.GetName()))
					//tracklog.WayPoints = append(tracklog.WayPoints, slice.Start)
				}
				tracks = append(tracks, track)
			}
			num += len(points)
		}
	}
	tracklog.Tracks = tracks
	return num, nil
}

func (c *SliceByWaypoints) slice(points []*gpx.Point) ([]*Slice, error) {
	segments, err := sliceByWaypoints(c.DistanceFunc, points, c.Waypoints, c.Threshold)
	if err != nil {
		return nil, err
	}
	slices := make([]*Slice, len(segments))
	for i, seg := range segments {
		slice := &Slice{
			Points: seg.points,
		}
		if seg.a.waypoint != nil {
			slice.Start = &gpx.WayPoint{
				Name:      seg.a.waypoint.Name,
				Latitude:  seg.a.point.Latitude,
				Longitude: seg.a.point.Longitude,
				Elevation: seg.a.point.Elevation,
			}
		}
		if seg.b.waypoint != nil {
			slice.End = &gpx.WayPoint{
				Name:      seg.b.waypoint.Name,
				Latitude:  seg.b.point.Latitude,
				Longitude: seg.b.point.Longitude,
				Elevation: seg.b.point.Elevation,
			}
		}
		slices[i] = slice
	}
	return slices, nil
}
