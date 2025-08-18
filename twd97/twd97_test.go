package twd97

import (
	"math"
	"testing"
)

const epsilon = 1e-6 // tolerance for floating point comparisons

// Test data from known coordinate pairs
// These are approximate values for testing purposes
var testCases = []struct {
	name     string
	wgs84Lat float64
	wgs84Lng float64
	twd97E   float64
	twd97N   float64
	pkm      bool
}{
	{
		name:     "Taipei 101 area",
		wgs84Lat: 25.0330,
		wgs84Lng: 121.5654,
		twd97E:   302925.0,
		twd97N:   2772325.0,
		pkm:      false,
	},
	{
		name:     "Taipei Main Station",
		wgs84Lat: 25.0478,
		wgs84Lng: 121.5170,
		twd97E:   297000.0,
		twd97N:   2772000.0,
		pkm:      false,
	},
	{
		name:     "Penghu area (PKM)",
		wgs84Lat: 23.5711,
		wgs84Lng: 119.5794,
		twd97E:   180000.0,
		twd97N:   2600000.0,
		pkm:      true,
	},
}

func TestFromWGS84(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			E, N := FromWGS84(tc.wgs84Lng, tc.wgs84Lat, tc.pkm)

			// Check that results are reasonable (positive values, reasonable ranges)
			if E <= 0 || N <= 0 {
				t.Errorf("FromWGS84(%f, %f, %t) returned invalid coordinates: E=%f, N=%f",
					tc.wgs84Lng, tc.wgs84Lat, tc.pkm, E, N)
			}

			// Check that E is in reasonable range for Taiwan (roughly 200k-400k meters)
			if E < 100000 || E > 500000 {
				t.Errorf("FromWGS84 E coordinate %f is outside expected range [100000, 500000]", E)
			}

			// Check that N is in reasonable range for Taiwan (roughly 2.5M-3M meters)
			if N < 2000000 || N > 3500000 {
				t.Errorf("FromWGS84 N coordinate %f is outside expected range [2000000, 3500000]", N)
			}
		})
	}
}

func TestToWGS84(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lat, lng := ToWGS84(tc.twd97E, tc.twd97N, tc.pkm)

			// Check that results are reasonable (valid latitude/longitude ranges)
			if lat < -90 || lat > 90 {
				t.Errorf("ToWGS84(%f, %f, %t) returned invalid latitude: %f",
					tc.twd97E, tc.twd97N, tc.pkm, lat)
			}

			if lng < -180 || lng > 180 {
				t.Errorf("ToWGS84(%f, %f, %t) returned invalid longitude: %f",
					tc.twd97E, tc.twd97N, tc.pkm, lng)
			}

			// Check that latitude is in Taiwan region (roughly 21-26 degrees)
			if lat < 21 || lat > 26 {
				t.Errorf("ToWGS84 latitude %f is outside Taiwan region [21, 26]", lat)
			}

			// Check that longitude is in reasonable range for Taiwan and PKM areas
			if tc.pkm {
				// PKM areas (Penghu, Kinmen, Matsu) can have longitude < 119
				if lng < 118 || lng > 120 {
					t.Errorf("ToWGS84 longitude %f is outside PKM region [118, 120]", lng)
				}
			} else {
				// Main Taiwan area
				if lng < 119 || lng > 122 {
					t.Errorf("ToWGS84 longitude %f is outside Taiwan region [119, 122]", lng)
				}
			}
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// WGS84 -> TWD97 -> WGS84
			E, N := FromWGS84(tc.wgs84Lng, tc.wgs84Lat, tc.pkm)
			lat, lng := ToWGS84(E, N, tc.pkm)

			// Check that round-trip conversion is close to original
			latDiff := math.Abs(lat - tc.wgs84Lat)
			lngDiff := math.Abs(lng - tc.wgs84Lng)

			// Allow some tolerance for floating point precision and conversion errors
			if latDiff > 0.001 { // ~100 meters
				t.Errorf("Round-trip conversion latitude difference too large: %f (original: %f, converted: %f)",
					latDiff, tc.wgs84Lat, lat)
			}

			if lngDiff > 0.001 { // ~100 meters
				t.Errorf("Round-trip conversion longitude difference too large: %f (original: %f, converted: %f)",
					lngDiff, tc.wgs84Lng, lng)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("taiwan region coordinates", func(t *testing.T) {
		// Test coordinates within Taiwan region
		E, N := FromWGS84(121.0, 23.5, false) // Southern Taiwan
		if E <= 0 || N <= 0 {
			t.Errorf("FromWGS84(121.0, 23.5, false) returned invalid coordinates: E=%f, N=%f", E, N)
		}

		lat, lng := ToWGS84(E, N, false)
		if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
			t.Errorf("ToWGS84(%f, %f, false) returned invalid coordinates: lat=%f, lng=%f", E, N, lat, lng)
		}
	})

	t.Run("boundary coordinates", func(t *testing.T) {
		// Test coordinates at Taiwan boundaries
		lat, _ := ToWGS84(150000, 2500000, false) // Southern boundary
		if lat < 20 || lat > 30 {
			t.Errorf("ToWGS84 near southern boundary returned unexpected latitude: %f", lat)
		}

		lat, lng := ToWGS84(350000, 2800000, false) // Northern boundary
		if lat < 20 || lat > 30 {
			t.Errorf("ToWGS84 near northern boundary returned unexpected latitude: %f", lat)
		}

		// Verify longitude is reasonable for Taiwan
		if lng < 115 || lng > 125 {
			t.Errorf("ToWGS84 returned longitude %f outside Taiwan region [115, 125]", lng)
		}
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("radians", func(t *testing.T) {
		testCases := []struct {
			degrees float64
			radians float64
		}{
			{0, 0},
			{90, math.Pi / 2},
			{180, math.Pi},
			{270, 3 * math.Pi / 2},
			{360, 2 * math.Pi},
		}

		for _, tc := range testCases {
			result := radians(tc.degrees)
			if math.Abs(result-tc.radians) > epsilon {
				t.Errorf("radians(%f) = %f, want %f", tc.degrees, result, tc.radians)
			}
		}
	})

	t.Run("todegdec", func(t *testing.T) {
		testValues := []float64{-180, -90, 0, 90, 180}
		for _, val := range testValues {
			result := todegdec(val)
			if result != val {
				t.Errorf("todegdec(%f) = %f, want %f", val, result, val)
			}
		}
	})

	t.Run("trigonometric functions", func(t *testing.T) {
		testAngle := math.Pi / 4 // 45 degrees

		// Test sin
		expected := math.Sin(testAngle)
		result := sin(testAngle)
		if math.Abs(result-expected) > epsilon {
			t.Errorf("sin(%f) = %f, want %f", testAngle, result, expected)
		}

		// Test cos
		expected = math.Cos(testAngle)
		result = cos(testAngle)
		if math.Abs(result-expected) > epsilon {
			t.Errorf("cos(%f) = %f, want %f", testAngle, result, expected)
		}

		// Test sinh
		expected = math.Sinh(testAngle)
		result = sinh(testAngle)
		if math.Abs(result-expected) > epsilon {
			t.Errorf("sinh(%f) = %f, want %f", testAngle, result, expected)
		}

		// Test cosh
		expected = math.Cosh(testAngle)
		result = cosh(testAngle)
		if math.Abs(result-expected) > epsilon {
			t.Errorf("cosh(%f) = %f, want %f", testAngle, result, expected)
		}
	})

	t.Run("inverse trigonometric functions", func(t *testing.T) {
		testValue := 0.5

		// Test asin
		expected := math.Asin(testValue)
		result := asin(testValue)
		if math.Abs(result-expected) > epsilon {
			t.Errorf("asin(%f) = %f, want %f", testValue, result, expected)
		}

		// Test atan
		expected = math.Atan(testValue)
		result = atan(testValue)
		if math.Abs(result-expected) > epsilon {
			t.Errorf("atan(%f) = %f, want %f", testValue, result, expected)
		}

		// Test atanh
		expected = math.Atanh(testValue)
		result = atanh(testValue)
		if math.Abs(result-expected) > epsilon {
			t.Errorf("atanh(%f) = %f, want %f", testValue, result, expected)
		}
	})

	t.Run("pow", func(t *testing.T) {
		testCases := []struct {
			x, y, expected float64
		}{
			{2, 3, 8},
			{4, 0.5, 2},
			{10, 0, 1},
			{5, 1, 5},
		}

		for _, tc := range testCases {
			result := pow(tc.x, tc.y)
			if math.Abs(result-tc.expected) > epsilon {
				t.Errorf("pow(%f, %f) = %f, want %f", tc.x, tc.y, result, tc.expected)
			}
		}
	})
}

func TestConstants(t *testing.T) {
	// Test that constants are reasonable
	if a <= 0 {
		t.Errorf("Earth radius 'a' should be positive, got %f", a)
	}

	if f <= 0 || f >= 1 {
		t.Errorf("Flattening 'f' should be between 0 and 1, got %f", f)
	}

	if k0 <= 0 || k0 >= 2 {
		t.Errorf("Scale factor 'k0' should be between 0 and 2, got %f", k0)
	}

	if E0 < 0 {
		t.Errorf("False easting 'E0' should be non-negative, got %f", E0)
	}

	if N0 != 0 {
		t.Errorf("False northing 'N0' should be 0, got %f", N0)
	}
}

func BenchmarkFromWGS84(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FromWGS84(121.5654, 25.0330, false)
	}
}

func BenchmarkToWGS84(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ToWGS84(302925.0, 2772325.0, false)
	}
}
