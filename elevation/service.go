package elevation

type Service interface {
	Lookup(points []*LatLon) ([]*float64, error)
}
