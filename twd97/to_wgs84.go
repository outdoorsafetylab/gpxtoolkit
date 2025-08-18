package twd97

import "math"

// ported from https://github.com/yychen/twd97

/*
Convert coordintes from TWD97 to WGS84

The east and north coordinates should be in meters and in float
pkm true for Penghu, Kinmen and Matsu area
*/
func ToWGS84(E, N float64, pkm bool) (float64, float64) {
	var _lng0 float64
	if pkm {
		_lng0 = lng0pkm
	} else {
		_lng0 = lng0
	}

	E /= 1000.0
	N /= 1000.0
	epsilon := (N - N0) / (k0 * A)
	eta := (E - E0) / (k0 * A)

	epsilonp := epsilon - beta1*sin(2*1*epsilon)*cosh(2*1*eta) -
		beta2*sin(2*2*epsilon)*cosh(2*2*eta) -
		beta3*sin(2*3*epsilon)*cosh(2*3*eta)
	etap := eta - beta1*cos(2*1*epsilon)*sinh(2*1*eta) -
		beta2*cos(2*2*epsilon)*sinh(2*2*eta) -
		beta3*cos(2*3*epsilon)*sinh(2*3*eta)
	// sigmap := 1 - 2*1*beta1*cos(2*1*epsilon)*cosh(2*1*eta) -
	// 	2*2*beta2*cos(2*2*epsilon)*cosh(2*2*eta) -
	// 	2*3*beta3*cos(2*3*epsilon)*cosh(2*3*eta)
	// taup := 2*1*beta1*sin(2*1*epsilon)*sinh(2*1*eta) +
	// 	2*2*beta2*sin(2*2*epsilon)*sinh(2*2*eta) +
	// 	2*3*beta3*sin(2*3*epsilon)*sinh(2*3*eta)

	chi := asin(sin(epsilonp) / cosh(etap))

	latitude := chi + delta1*sin(2*1*chi) +
		delta2*sin(2*2*chi) +
		delta3*sin(2*3*chi)

	longitude := _lng0 + atan(sinh(etap)/cos(epsilonp))

	return latitude * (180 / math.Pi), longitude * (180 / math.Pi)
}
