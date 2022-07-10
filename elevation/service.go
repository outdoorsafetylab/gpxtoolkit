package elevation

import "math"

type Service interface {
	Lookup(points []*LatLon) ([]*float64, error)
}

func IsValid(elev *float64) bool {
	return elev != nil && !math.IsNaN(*elev)
}
