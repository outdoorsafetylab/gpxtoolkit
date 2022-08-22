package gpx

type AlphaFilter struct {
	Alpha float64
	Value float64
}

func (a *AlphaFilter) Accumulate(delta float64) float64 {
	delta *= a.Alpha
	a.Value += delta
	return delta
}
