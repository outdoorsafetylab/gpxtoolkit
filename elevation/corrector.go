package elevation

import (
	"gpxtoolkit/gpx"
	"math"
)

type Corrector struct {
	Service Service
}

func (c *Corrector) Correct(log *gpx.TrackLog) (int, error) {
	points := make([]*LatLon, 0)
	for _, p := range log.WayPoints {
		points = append(points, &LatLon{p.GetLongitude(), p.GetLatitude()})
	}
	for _, t := range log.Tracks {
		for _, s := range t.Segments {
			for _, p := range s.Points {
				points = append(points, &LatLon{p.GetLongitude(), p.GetLatitude()})
			}
		}
	}
	alts, err := c.Service.Lookup(points)
	if err != nil {
		return 0, err
	}
	i := 0
	for _, p := range log.WayPoints {
		alt := alts[i]
		i++
		if alt == nil || math.IsNaN(*alt) {
			continue
		}
		p.Elevation = alt
	}
	for _, t := range log.Tracks {
		for _, s := range t.Segments {
			for _, p := range s.Points {
				alt := alts[i]
				i++
				if alt == nil || math.IsNaN(*alt) {
					continue
				}
				p.Elevation = alt
			}
		}
	}
	return len(alts), nil
}
