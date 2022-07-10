package gpxutil

import (
	"fmt"
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"log"
	"math"
	"time"

	"google.golang.org/protobuf/proto"
)

type Interpolate struct {
	Service           elevation.Service
	Distance          float64
	ByTerrainDistance bool
	distanceFunc      DistanceFunc
}

func (c *Interpolate) Name() string {
	return fmt.Sprintf("Interpolate by Distance %f", c.Distance)
}

func (c *Interpolate) Run(tracklog *gpx.TrackLog) (int, error) {
	if c.ByTerrainDistance {
		c.distanceFunc = terrainDistance
	} else {
		c.distanceFunc = horizontalDistance
	}
	n := 0
	for _, t := range tracklog.Tracks {
		for _, seg := range t.Segments {
			num := len(seg.Points)
			points, err := c.interpolate(seg.Points)
			if err != nil {
				return 0, err
			}
			if c.ByTerrainDistance && c.Service != nil {
				_, err := correctPoints(c.Service, seg.Points)
				if err != nil {
					return 0, err
				}
			}
			n += (len(points) - num)
			seg.Points = points
		}
	}
	return n, nil
}

func (c *Interpolate) interpolate(points []*gpx.Point) ([]*gpx.Point, error) {
	interpolated := make([]*gpx.Point, 0)
	lines := getLines(c.distanceFunc, points)
	res := make([]*gpx.Point, 0)
	for _, line := range lines {
		res = append(res, line.a)
		num := int(math.Round(line.dist / c.Distance))
		if num < 0 {
			continue
		}
		for i := 1; i < num; i++ {
			p := interpolate(line.a, line.b, float64(i)/float64(num))
			interpolated = append(interpolated, p)
			res = append(res, p)
		}
	}
	if c.Service != nil {
		_, err := correctPoints(c.Service, interpolated)
		if err != nil {
			return nil, err
		}
	}
	log.Printf("Interpolated %d points", len(interpolated))
	res = append(res, lines[len(lines)-1].b)
	return res, nil
}

func interpolate(a, b *gpx.Point, ratio float64) *gpx.Point {
	lat1 := a.GetLatitude()
	lat2 := b.GetLatitude()
	lon1 := a.GetLongitude()
	lon2 := b.GetLongitude()
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	t1 := a.Time()
	t2 := b.Time()
	dt := t2.Sub(t1)
	lat := lat1 + dlat*ratio
	lon := lon1 + dlon*ratio
	p := &gpx.Point{
		Latitude:  proto.Float64(lat),
		Longitude: proto.Float64(lon),
	}
	if a.Elevation != nil && b.Elevation != nil {
		ele1 := a.GetElevation()
		dele := b.GetElevation() - ele1
		p.Elevation = proto.Float64(ele1 + dele*ratio)
	}
	if a.NanoTime != nil && b.NanoTime != nil {
		p.NanoTime = proto.Int64(t1.Add(dt * time.Duration(ratio)).UnixNano())
	}
	return p
}
