package gpxutil

import (
	"gpxtoolkit/gpx"
	"time"

	"google.golang.org/protobuf/proto"
)

func Interpolate(a, b *gpx.Point, ratio float64) *gpx.Point {
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
	lat := lat1 + dlat*ratio
	lon := lon1 + dlon*ratio
	res := &gpx.Point{
		Latitude:  proto.Float64(lat),
		Longitude: proto.Float64(lon),
	}
	if a.Elevation != nil && b.Elevation != nil {
		res.Elevation = proto.Float64(ele1 + dele*ratio)
	}
	if a.NanoTime != nil && b.NanoTime != nil {
		res.NanoTime = proto.Int64(t1.Add(dt * time.Duration(ratio)).UnixNano())
	}
	return res
}
