package elevation

import (
	"context"
	"fmt"
	"time"

	"gpxtoolkit/log"

	"googlemaps.github.io/maps"
)

type Google struct {
	APIKey string
}

func (s *Google) Lookup(latLons []*LatLon) ([]*float64, error) {
	result := make([]*float64, 0)
	for len(latLons) > 0 {
		n := len(latLons)
		if n > 500 {
			n = 500
		}
		req := latLons[0:n]
		latLons = latLons[n:]
		res, err := s.lookup(req)
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}
	return result, nil
}

func (s *Google) lookup(latLons []*LatLon) ([]*float64, error) {
	c, err := maps.NewClient(maps.WithAPIKey(s.APIKey))
	if err != nil {
		return nil, err
	}
	r := &maps.ElevationRequest{
		Locations: make([]maps.LatLng, len(latLons)),
	}
	for i, coords := range latLons {
		r.Locations[i] = maps.LatLng{
			Lat: coords.Lat,
			Lng: coords.Lon,
		}
	}
	start := time.Now()
	defer func() {
		log.Debugf("Looked up %d points in %v", len(latLons), time.Since(start))
	}()
	res, err := c.Elevation(context.Background(), r)
	if err != nil {
		return nil, err
	}
	if len(res) != len(latLons) {
		err = fmt.Errorf("unepxected number of result: expect=%d, was=%d", len(latLons), len(res))
		log.Errorf("%s", err.Error())
		return nil, err
	}
	points := make([]*float64, len(res))
	for i, r := range res {
		elev := r.Elevation
		points[i] = &elev
	}
	return points, nil
}
