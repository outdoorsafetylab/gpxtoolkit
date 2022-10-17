package gpx

import "math"

type BoundingBox struct {
	Min, Max *struct {
		Latitude, Longitude float64
	}
}

func (b *BoundingBox) Merge(o *BoundingBox) {
	if b.Min == nil {
		b.Min = o.Min
	} else {
		b.Min.Latitude = math.Min(b.Min.Latitude, o.Min.Latitude)
		b.Min.Longitude = math.Min(b.Min.Longitude, o.Min.Longitude)
	}
	if b.Max == nil {
		b.Max = o.Max
	} else {
		b.Max.Latitude = math.Max(b.Max.Latitude, o.Max.Latitude)
		b.Max.Longitude = math.Max(b.Max.Longitude, o.Max.Longitude)
	}
}

func (b *BoundingBox) Add(latitude, longitude float64) {
	if b.Min == nil {
		b.Min = &struct {
			Latitude  float64
			Longitude float64
		}{
			Latitude:  latitude,
			Longitude: longitude,
		}
	} else {
		b.Min.Latitude = math.Min(b.Min.Latitude, latitude)
		b.Min.Longitude = math.Min(b.Min.Longitude, longitude)
	}
	if b.Max == nil {
		b.Max = &struct {
			Latitude  float64
			Longitude float64
		}{
			Latitude:  latitude,
			Longitude: longitude,
		}
	} else {
		b.Max.Latitude = math.Max(b.Max.Latitude, latitude)
		b.Max.Longitude = math.Max(b.Max.Longitude, longitude)
	}
}

func (b *BoundingBox) Expand(ratio float64) {
	lat := (b.Max.Latitude - b.Min.Latitude) * ratio
	lon := (b.Max.Longitude - b.Min.Longitude) * ratio
	b.Min.Latitude -= lat
	b.Max.Latitude += lat
	b.Min.Longitude -= lon
	b.Max.Longitude += lon
}
