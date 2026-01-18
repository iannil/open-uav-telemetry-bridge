// Package coordinator provides coordinate system conversion utilities.
// It handles conversion between WGS84 (GPS), GCJ02 (China), and BD09 (Baidu) coordinate systems.
package coordinator

import "math"

const (
	// WGS84 ellipsoid parameters
	a  = 6378245.0              // Semi-major axis
	ee = 0.00669342162296594323 // Eccentricity squared
)

// Converter handles coordinate system conversions
type Converter struct {
	EnableGCJ02 bool
	EnableBD09  bool
}

// New creates a new Converter with the specified options
func New(enableGCJ02, enableBD09 bool) *Converter {
	return &Converter{
		EnableGCJ02: enableGCJ02,
		EnableBD09:  enableBD09,
	}
}

// ConvertResult holds the conversion results for all coordinate systems
type ConvertResult struct {
	// Original WGS84 coordinates
	Lat float64
	Lon float64

	// GCJ02 (Mars coordinates) - used by Amap, Tencent, Google China
	LatGCJ02 *float64
	LonGCJ02 *float64

	// BD09 (Baidu coordinates)
	LatBD09 *float64
	LonBD09 *float64
}

// Convert converts WGS84 coordinates to other coordinate systems based on configuration
func (c *Converter) Convert(lat, lon float64) ConvertResult {
	result := ConvertResult{
		Lat: lat,
		Lon: lon,
	}

	if c.EnableGCJ02 || c.EnableBD09 {
		gcjLat, gcjLon := WGS84ToGCJ02(lat, lon)
		if c.EnableGCJ02 {
			result.LatGCJ02 = &gcjLat
			result.LonGCJ02 = &gcjLon
		}
		if c.EnableBD09 {
			bdLat, bdLon := GCJ02ToBD09(gcjLat, gcjLon)
			result.LatBD09 = &bdLat
			result.LonBD09 = &bdLon
		}
	}

	return result
}

// WGS84ToGCJ02 converts WGS84 coordinates to GCJ02 (Mars coordinates)
// Used by: Amap (高德), Tencent Maps (腾讯), Google China
func WGS84ToGCJ02(lat, lon float64) (float64, float64) {
	if outOfChina(lat, lon) {
		return lat, lon // Outside China, no conversion needed
	}

	dLat := transformLat(lon-105.0, lat-35.0)
	dLon := transformLon(lon-105.0, lat-35.0)

	radLat := lat / 180.0 * math.Pi
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)

	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * math.Pi)
	dLon = (dLon * 180.0) / (a / sqrtMagic * math.Cos(radLat) * math.Pi)

	return lat + dLat, lon + dLon
}

// GCJ02ToWGS84 converts GCJ02 coordinates back to WGS84 (approximate)
// Uses iterative method for better accuracy
func GCJ02ToWGS84(lat, lon float64) (float64, float64) {
	if outOfChina(lat, lon) {
		return lat, lon
	}

	// Iterative approach for better accuracy
	wgsLat, wgsLon := lat, lon
	for i := 0; i < 3; i++ {
		gcjLat, gcjLon := WGS84ToGCJ02(wgsLat, wgsLon)
		wgsLat = lat - (gcjLat - lat)
		wgsLon = lon - (gcjLon - lon)
	}
	return wgsLat, wgsLon
}

// GCJ02ToBD09 converts GCJ02 coordinates to BD09 (Baidu coordinates)
func GCJ02ToBD09(lat, lon float64) (float64, float64) {
	x := lon
	y := lat
	z := math.Sqrt(x*x+y*y) + 0.00002*math.Sin(y*math.Pi*3000.0/180.0)
	theta := math.Atan2(y, x) + 0.000003*math.Cos(x*math.Pi*3000.0/180.0)
	bdLon := z*math.Cos(theta) + 0.0065
	bdLat := z*math.Sin(theta) + 0.006
	return bdLat, bdLon
}

// BD09ToGCJ02 converts BD09 coordinates back to GCJ02
func BD09ToGCJ02(lat, lon float64) (float64, float64) {
	x := lon - 0.0065
	y := lat - 0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*math.Pi*3000.0/180.0)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*math.Pi*3000.0/180.0)
	gcjLon := z * math.Cos(theta)
	gcjLat := z * math.Sin(theta)
	return gcjLat, gcjLon
}

// WGS84ToBD09 converts WGS84 directly to BD09
func WGS84ToBD09(lat, lon float64) (float64, float64) {
	gcjLat, gcjLon := WGS84ToGCJ02(lat, lon)
	return GCJ02ToBD09(gcjLat, gcjLon)
}

// BD09ToWGS84 converts BD09 directly to WGS84
func BD09ToWGS84(lat, lon float64) (float64, float64) {
	gcjLat, gcjLon := BD09ToGCJ02(lat, lon)
	return GCJ02ToWGS84(gcjLat, gcjLon)
}

// outOfChina checks if the coordinate is outside China's boundary
func outOfChina(lat, lon float64) bool {
	// Rough boundary of China
	return lon < 72.004 || lon > 137.8347 || lat < 0.8293 || lat > 55.8271
}

// transformLat is a helper function for GCJ02 conversion
func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*math.Pi) + 40.0*math.Sin(y/3.0*math.Pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*math.Pi) + 320*math.Sin(y*math.Pi/30.0)) * 2.0 / 3.0
	return ret
}

// transformLon is a helper function for GCJ02 conversion
func transformLon(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*math.Pi) + 40.0*math.Sin(x/3.0*math.Pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*math.Pi) + 300.0*math.Sin(x/30.0*math.Pi)) * 2.0 / 3.0
	return ret
}
