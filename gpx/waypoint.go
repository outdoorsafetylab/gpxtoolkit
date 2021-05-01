package gpx

import "time"

func (p *WayPoint) Time() time.Time {
	return time.Unix(0, p.GetNanoTime()).UTC()
}

func (p *WayPoint) Millis() int64 {
	return p.Time().UnixNano() / int64(time.Millisecond)
}

func (p *WayPoint) DistanceTo(o *WayPoint) float64 {
	return GeoDistance(p.GetLatitude(), p.GetLongitude(), o.GetLatitude(), o.GetLongitude())
}
