package gpxutil

import (
	"math"
)

func toWebMercator(lat, lon float64) (float64, float64) {
	x := lon/360 + 0.5
	sin := math.Sin(lat * math.Pi / 180)
	y := (0.5 - 0.25*math.Log((1+sin)/(1-sin))/math.Pi)
	if y < 0 {
		y = 0
	} else if y > 1 {
		y = 1
	}
	return x, y
}

// func fromWebMercator(x, y float64) (float64, float64) {
// 	y2 := (180 - y*360) * math.Pi / 180
// 	lat := 360*math.Atan(math.Exp(y2))/math.Pi - 90
// 	lon := (x - 0.5) * 360
// 	return lat, lon
// }
