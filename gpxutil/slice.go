package gpxutil

import "gpxtoolkit/gpx"

type Slice struct {
	Start, End *gpx.WayPoint
	Points     []*gpx.Point
}

func SliceByWaypoints(distanceFunc DistanceFunc, points []*gpx.Point, waypoints []*gpx.WayPoint, threshold float64) ([]*Slice, error) {
	projections, err := projectWaypoints(distanceFunc, points, waypoints, threshold)
	if err != nil {
		return nil, err
	}
	segments := projections.slice(points)
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
