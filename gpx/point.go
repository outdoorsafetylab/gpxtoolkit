package gpx

import (
	"time"
)

func (p *Point) Time() time.Time {
	return time.Unix(0, p.GetNanoTime()).UTC()
}

func (p *Point) Millis() int64 {
	return p.Time().UnixNano() / int64(time.Millisecond)
}

func (p *Point) distanceTo(o *Point) float64 {
	return GeoDistance(p.GetLatitude(), p.GetLongitude(), o.GetLatitude(), o.GetLongitude())
}

func (p *Point) Equals(o *Point) bool {
	if p.GetLatitude() != o.GetLatitude() || p.GetLongitude() != o.GetLongitude() {
		return false
	}
	if p.Elevation == nil && o.Elevation != nil {
		return false
	}
	if p.Elevation != nil && o.Elevation == nil {
		return false
	}
	if p.GetElevation() != o.GetElevation() {
		return false
	}
	if p.NanoTime == nil && o.NanoTime != nil {
		return false
	}
	if p.NanoTime != nil && o.NanoTime == nil {
		return false
	}
	if p.GetNanoTime() != o.GetNanoTime() {
		return false
	}
	return true
}
