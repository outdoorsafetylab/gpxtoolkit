package gpx

import "time"

func (p *Point) Time() time.Time {
	return time.Unix(0, p.GetNanoTime()).UTC()
}

func (p *Point) Millis() int64 {
	return p.Time().UnixNano() / int64(time.Millisecond)
}

func (p *Point) distanceTo(o *Point) float64 {
	return GeoDistance(p.GetLatitude(), p.GetLongitude(), o.GetLatitude(), o.GetLongitude())
}
