package gpxutil

import (
	"gpxtoolkit/elevation"
	"gpxtoolkit/gpx"
	"math"

	"google.golang.org/protobuf/proto"
)

type CorrectElevation struct {
	Service elevation.Service
}

func (c *CorrectElevation) Name() string {
	return "Correct Elevation"
}

func (c *CorrectElevation) Run(tracklog *gpx.TrackLog) (int, error) {
	points := make([]*elevation.LatLon, 0)
	for _, p := range tracklog.WayPoints {
		points = append(points, &elevation.LatLon{Lat: p.GetLatitude(), Lon: p.GetLongitude()})
	}
	elevations, err := c.Service.Lookup(points)
	if err != nil {
		return 0, err
	}
	n := 0
	for i, p := range tracklog.WayPoints {
		elev := elevations[i]
		if elev == nil || math.IsNaN(*elev) {
			continue
		}
		p.Elevation = proto.Float64(math.Round(*elev))
		n++
	}
	for _, t := range tracklog.Tracks {
		for _, s := range t.Segments {
			points := make([]*elevation.LatLon, 0)
			for _, p := range s.Points {
				points = append(points, &elevation.LatLon{Lat: p.GetLatitude(), Lon: p.GetLongitude()})
			}
			elevations, err = c.Service.Lookup(points)
			if err != nil {
				return 0, err
			}
			for i, p := range s.Points {
				elev := elevations[i]
				if elev == nil || math.IsNaN(*elev) {
					continue
				}
				p.Elevation = proto.Float64(math.Round(*elev))
				n++
			}
		}
	}
	return n, nil
}
