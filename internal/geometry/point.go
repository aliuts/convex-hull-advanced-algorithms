package geometry

import "math"

// Point is a two-dimensional Cartesian coordinate.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Validate rejects coordinates that cannot be represented in JSON or arithmetic.
func (p Point) Validate() bool {
	return !math.IsNaN(p.X) && !math.IsNaN(p.Y) && !math.IsInf(p.X, 0) && !math.IsInf(p.Y, 0)
}
