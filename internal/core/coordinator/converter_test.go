package coordinator

import (
	"math"
	"testing"
)

// Test data: known coordinate conversions
// Reference: Online coordinate conversion tools
var testCases = []struct {
	name       string
	wgs84Lat   float64
	wgs84Lon   float64
	gcj02Lat   float64
	gcj02Lon   float64
	bd09Lat    float64
	bd09Lon    float64
	inChina    bool
}{
	{
		name:       "Tiananmen Square",
		wgs84Lat:   39.908722,
		wgs84Lon:   116.397499,
		gcj02Lat:   39.911119,
		gcj02Lon:   116.403963,
		bd09Lat:    39.917458,
		bd09Lon:    116.410394,
		inChina:    true,
	},
	{
		name:       "Oriental Pearl Tower",
		wgs84Lat:   31.239702,
		wgs84Lon:   121.499763,
		gcj02Lat:   31.237823,
		gcj02Lon:   121.505854,
		bd09Lat:    31.243603,
		bd09Lon:    121.512350,
		inChina:    true,
	},
	{
		name:       "Canton Tower",
		wgs84Lat:   23.106593,
		wgs84Lon:   113.324553,
		gcj02Lat:   23.104373,
		gcj02Lon:   113.331200,
		bd09Lat:    23.110476,
		bd09Lon:    113.337710,
		inChina:    true,
	},
	{
		name:       "Tokyo Tower (outside China)",
		wgs84Lat:   35.658581,
		wgs84Lon:   139.745438,
		gcj02Lat:   35.658581, // No conversion outside China
		gcj02Lon:   139.745438,
		bd09Lat:    35.664663,
		bd09Lon:    139.752036,
		inChina:    false,
	},
}

// tolerance for coordinate comparison (about 200 meters)
// Different implementations may have slight variations
const tolerance = 0.002

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

func TestWGS84ToGCJ02(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotLat, gotLon := WGS84ToGCJ02(tc.wgs84Lat, tc.wgs84Lon)

			if tc.inChina {
				if !almostEqual(gotLat, tc.gcj02Lat, tolerance) {
					t.Errorf("Latitude: got %f, want %f (diff: %f)", gotLat, tc.gcj02Lat, math.Abs(gotLat-tc.gcj02Lat))
				}
				if !almostEqual(gotLon, tc.gcj02Lon, tolerance) {
					t.Errorf("Longitude: got %f, want %f (diff: %f)", gotLon, tc.gcj02Lon, math.Abs(gotLon-tc.gcj02Lon))
				}
			} else {
				// Outside China, should return original coordinates
				if gotLat != tc.wgs84Lat || gotLon != tc.wgs84Lon {
					t.Errorf("Outside China should not convert: got (%f, %f), want (%f, %f)",
						gotLat, gotLon, tc.wgs84Lat, tc.wgs84Lon)
				}
			}
		})
	}
}

func TestGCJ02ToBD09(t *testing.T) {
	for _, tc := range testCases {
		if !tc.inChina {
			continue // Skip non-China test cases for this test
		}
		t.Run(tc.name, func(t *testing.T) {
			gotLat, gotLon := GCJ02ToBD09(tc.gcj02Lat, tc.gcj02Lon)

			if !almostEqual(gotLat, tc.bd09Lat, tolerance) {
				t.Errorf("Latitude: got %f, want %f (diff: %f)", gotLat, tc.bd09Lat, math.Abs(gotLat-tc.bd09Lat))
			}
			if !almostEqual(gotLon, tc.bd09Lon, tolerance) {
				t.Errorf("Longitude: got %f, want %f (diff: %f)", gotLon, tc.bd09Lon, math.Abs(gotLon-tc.bd09Lon))
			}
		})
	}
}

func TestWGS84ToBD09(t *testing.T) {
	for _, tc := range testCases {
		if !tc.inChina {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			gotLat, gotLon := WGS84ToBD09(tc.wgs84Lat, tc.wgs84Lon)

			if !almostEqual(gotLat, tc.bd09Lat, tolerance) {
				t.Errorf("Latitude: got %f, want %f (diff: %f)", gotLat, tc.bd09Lat, math.Abs(gotLat-tc.bd09Lat))
			}
			if !almostEqual(gotLon, tc.bd09Lon, tolerance) {
				t.Errorf("Longitude: got %f, want %f (diff: %f)", gotLon, tc.bd09Lon, math.Abs(gotLon-tc.bd09Lon))
			}
		})
	}
}

func TestGCJ02ToWGS84(t *testing.T) {
	for _, tc := range testCases {
		if !tc.inChina {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			gotLat, gotLon := GCJ02ToWGS84(tc.gcj02Lat, tc.gcj02Lon)

			// Reverse conversion should be close to original
			if !almostEqual(gotLat, tc.wgs84Lat, tolerance) {
				t.Errorf("Latitude: got %f, want %f (diff: %f)", gotLat, tc.wgs84Lat, math.Abs(gotLat-tc.wgs84Lat))
			}
			if !almostEqual(gotLon, tc.wgs84Lon, tolerance) {
				t.Errorf("Longitude: got %f, want %f (diff: %f)", gotLon, tc.wgs84Lon, math.Abs(gotLon-tc.wgs84Lon))
			}
		})
	}
}

func TestBD09ToGCJ02(t *testing.T) {
	for _, tc := range testCases {
		if !tc.inChina {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			gotLat, gotLon := BD09ToGCJ02(tc.bd09Lat, tc.bd09Lon)

			if !almostEqual(gotLat, tc.gcj02Lat, tolerance) {
				t.Errorf("Latitude: got %f, want %f (diff: %f)", gotLat, tc.gcj02Lat, math.Abs(gotLat-tc.gcj02Lat))
			}
			if !almostEqual(gotLon, tc.gcj02Lon, tolerance) {
				t.Errorf("Longitude: got %f, want %f (diff: %f)", gotLon, tc.gcj02Lon, math.Abs(gotLon-tc.gcj02Lon))
			}
		})
	}
}

func TestOutOfChina(t *testing.T) {
	tests := []struct {
		lat, lon float64
		expected bool
	}{
		{39.9, 116.4, false},   // Beijing - in China
		{31.2, 121.5, false},   // Shanghai - in China
		{35.6, 139.7, true},    // Tokyo - outside China
		{40.7, -74.0, true},    // New York - outside China
		{-33.8, 151.2, true},   // Sydney - outside China
		{22.3, 114.1, false},   // Hong Kong - in China boundary
	}

	for _, tt := range tests {
		got := outOfChina(tt.lat, tt.lon)
		if got != tt.expected {
			t.Errorf("outOfChina(%f, %f) = %v, want %v", tt.lat, tt.lon, got, tt.expected)
		}
	}
}

func TestConverter(t *testing.T) {
	t.Run("Both enabled", func(t *testing.T) {
		c := New(true, true)
		result := c.Convert(39.908722, 116.397499) // Tiananmen

		if result.LatGCJ02 == nil || result.LonGCJ02 == nil {
			t.Error("GCJ02 should be set")
		}
		if result.LatBD09 == nil || result.LonBD09 == nil {
			t.Error("BD09 should be set")
		}
	})

	t.Run("Only GCJ02", func(t *testing.T) {
		c := New(true, false)
		result := c.Convert(39.908722, 116.397499)

		if result.LatGCJ02 == nil || result.LonGCJ02 == nil {
			t.Error("GCJ02 should be set")
		}
		if result.LatBD09 != nil || result.LonBD09 != nil {
			t.Error("BD09 should not be set")
		}
	})

	t.Run("Both disabled", func(t *testing.T) {
		c := New(false, false)
		result := c.Convert(39.908722, 116.397499)

		if result.LatGCJ02 != nil || result.LonGCJ02 != nil {
			t.Error("GCJ02 should not be set")
		}
		if result.LatBD09 != nil || result.LonBD09 != nil {
			t.Error("BD09 should not be set")
		}
	})
}

func BenchmarkWGS84ToGCJ02(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WGS84ToGCJ02(39.908722, 116.397499)
	}
}

func BenchmarkWGS84ToBD09(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WGS84ToBD09(39.908722, 116.397499)
	}
}
