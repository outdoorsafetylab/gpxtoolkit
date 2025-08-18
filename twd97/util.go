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
	beta1  = n/2 - 2*pow(n, 2)/3.0 + 37*pow(n, 3)/96.0
	beta2  = pow(n, 2)/48.0 + pow(n, 3)/15.0
	beta3  = 17 * pow(n, 3) / 480.0
	delta1 = 2*n - 2*pow(n, 2)/3.0 - 2*pow(n, 3)
	delta2 = 7*pow(n, 2)/3.0 - 8*pow(n, 3)/5.0
	delta3 = 56 * pow(n, 3) / 15.0
)

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

func asin(x float64) float64 {
	return math.Asin(x)
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
