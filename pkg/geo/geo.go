package geo

import "math"

// EarthRadiusMeters represents the radius of the earth in meters
const EarthRadiusMeters = 6378100.0

// Coord represents a (longitude, latitude) pair in that order
type Coord [2]float64

// from: https://gist.github.com/cdipaolo/d3f8db3848278b49db68

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(latlon1 Coord, latlon2 Coord) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = latlon1[1] * math.Pi / 180
	lo1 = latlon1[0] * math.Pi / 180
	la2 = latlon2[1] * math.Pi / 180
	lo2 = latlon2[0] * math.Pi / 180

	r = EarthRadiusMeters

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}
