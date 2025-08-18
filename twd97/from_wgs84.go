package twd97

// ported from https://github.com/yychen/twd97

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
