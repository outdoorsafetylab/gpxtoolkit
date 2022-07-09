package gpxutil

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"math"
	"time"

	"google.golang.org/protobuf/proto"
)

type Interpolate struct {
	Service      elevation.Service
	DistanceFunc DistanceFunc
	Distance     float64
}

func (c *Interpolate) Name() string {
	return fmt.Sprintf("Interpolate by Distance %f", c.Distance)
}

func (c *Interpolate) Run(tracklog *gpx.TrackLog) (int, error) {
	n := 0
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			num := len(seg.Points)
			points, err := c.interpolate(seg.Points)
			if err != nil {
				return 0, err
			}
			n += (len(points) - num)
			seg.Points = points
		}
	}
	return n, nil
}

func (c *Interpolate) interpolate(points []*gpx.Point) ([]*gpx.Point, error) {
	lines := getLines(c.DistanceFunc, points)
	res := make([]*gpx.Point, 0)
	for _, line := range lines {
		res = append(res, line.a)
		num := int(math.Round(line.dist / c.Distance))
		if num < 0 {
			continue
		}
		for i := 1; i < num; i++ {
			p := interpolate(line.a, line.b, float64(i)/float64(num), c.Service)
			res = append(res, p)
		}
	}
	res = append(res, lines[len(lines)-1].b)
	return res, nil
}

func interpolate(a, b *gpx.Point, ratio float64, service elevation.Service) *gpx.Point {
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
	p := &gpx.Point{
		Latitude:  proto.Float64(lat),
		Longitude: proto.Float64(lon),
	}
	if service != nil {
		elev, err := elevation.Lookup(service, p.GetLatitude(), p.GetLongitude())
		if err != nil && elevation.IsValid(elev) {
			p.Elevation = proto.Float64(*elev)
		}
	}
	if p.Elevation == nil && a.Elevation != nil && b.Elevation != nil {
		p.Elevation = proto.Float64(ele1 + dele*ratio)
	}
	if a.NanoTime != nil && b.NanoTime != nil {
		p.NanoTime = proto.Int64(t1.Add(dt * time.Duration(ratio)).UnixNano())
	}
	return p
}
