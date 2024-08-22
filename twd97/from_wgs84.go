package twd97

import "math"

// ported from https://github.com/yychen/twd97

var (
	a       = 6378.137
	f       = 1 / 298.257222101
	k0      = 0.9999
	N0      = 0.0
	E0      = 250.000
	lng0    = radians(121)
	lng0pkm = radians(119)

	n      = f / (2 - f)
	A      = a / (1 + n) * (1 + pow(n, 2)/4.0 + pow(n, 4)/64.0)
	alpha1 = n/2 - 2*pow(n, 2)/3.0 + 5*pow(n, 3)/16.0
	alpha2 = 13*pow(n, 2)/48.0 - 3*pow(n, 3)/5.0
	alpha3 = 61 * pow(n, 3) / 240.0
)

func FromWGS84(lng, lat float64, pkm bool) (float64, float64) {
	lat = radians(todegdec(lat))
	lng = radians(todegdec(lng))

	_lng0 := lng0
	if pkm {
		_lng0 = lng0pkm
	}

	t := sinh((atanh(sin(lat)) - 2*pow(n, 0.5)/(1+n)*atanh(2*pow(n, 0.5)/(1+n)*sin(lat))))
	epsilonp := atan(t / cos(lng-_lng0))
	etap := atan(sin(lng-_lng0) / pow(1+t*t, 0.5))

	E := E0 + k0*A*(etap+alpha1*cos(2*1*epsilonp)*sinh(2*1*etap)+
		alpha2*cos(2*2*epsilonp)*sinh(2*2*etap)+
		alpha3*cos(2*3*epsilonp)*sinh(2*3*etap))
	N := N0 + k0*A*(epsilonp+alpha1*sin(2*1*epsilonp)*cosh(2*1*etap)+
		alpha2*sin(2*2*epsilonp)*cosh(2*2*etap)+
		alpha3*sin(2*3*epsilonp)*cosh(2*3*etap))

	return E * 1000, N * 1000
}

func radians(degrees float64) float64 {
	return degrees * (math.Pi / 180)
}

func todegdec(x float64) float64 {
	return x
}

func sinh(x float64) float64 {
	return math.Sinh(x)
}

func cosh(x float64) float64 {
	return math.Cosh(x)
}

func sin(x float64) float64 {
	return math.Sin(x)
}

func cos(x float64) float64 {
	return math.Cos(x)
}

func atanh(x float64) float64 {
	return math.Atanh(x)
}

func atan(x float64) float64 {
	return math.Atan(x)
}

func pow(x, y float64) float64 {
	return math.Pow(x, y)
}
