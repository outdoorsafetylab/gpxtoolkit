package elevation

import "math"

type Service interface {
	Lookup(points []*LatLon) ([]*float64, error)
}

func Lookup(service Service, lat, lon float64) (*float64, error) {
	res, err := service.Lookup([]*LatLon{{Lat: lat, Lon: lon}})
	if err != nil {
		return nil, err
	}
	return res[0], nil
}

func IsValid(elev *float64) bool {
	return elev != nil && !math.IsNaN(*elev)
}
