package gpxutil

import (
	"fmt"
	"gpxtoolkit/gpx"
	"gpxtoolkit/simpleline"
	"time"

	"google.golang.org/protobuf/proto"
)

type Simplify struct {
	Epsilon float64
	First   bool
}

func (s *Simplify) Name() string {
	return fmt.Sprintf("Simplify with Epsilon %f", s.Epsilon)
}

func (s *Simplify) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			num := len(seg.Points)
			points, err := s.simplify(seg.Points)
			if err != nil {
				return 0, err
			}
			n += (num - len(points))
			seg.Points = points
		}
	}
	return n, nil
}

func (s *Simplify) simplify(points []*gpx.Point) ([]*gpx.Point, error) {
	dataPoints := make([]simpleline.Point, len(points))
	for i, p := range points {
		dataPoints[i] = &simpleline.Point3d{
			X: p.GetLatitude(),
			Y: p.GetLongitude(),
			Z: float64(p.Time().Unix()),
		}
	}
	simplified, err := simpleline.RDP(dataPoints, s.Epsilon, func(p1, p2 simpleline.Point) float64 {
		return gpx.GeoDistance(p1.Vector()[0], p1.Vector()[1], p2.Vector()[0], p2.Vector()[1])
	}, s.First)
	if err != nil {
		return nil, err
	}
	res := make([]*gpx.Point, len(simplified))
	for i, p := range simplified {
		res[i] = &gpx.Point{
			Latitude:  proto.Float64(p.Vector()[0]),
			Longitude: proto.Float64(p.Vector()[1]),
			NanoTime:  proto.Int64(time.Unix(int64(p.Vector()[2]), 0).UnixNano()),
		}
	}
	return res, nil
}
