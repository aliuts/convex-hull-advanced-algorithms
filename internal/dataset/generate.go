package dataset

import (
	"errors"
	"math"
	"math/rand"

	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

const MaxPoints = 10000

var ErrInvalidDataset = errors.New("invalid dataset parameters")

// Generate deterministically produces one of uniform, clustered, circular, or collinear datasets.
func Generate(kind string, count int, seed int64) ([]geometry.Point, error) {
	if count < 0 || count > MaxPoints {
		return nil, ErrInvalidDataset
	}
	r := rand.New(rand.NewSource(seed))
	out := make([]geometry.Point, count)
	switch kind {
	case "uniform":
		for i := range out {
			out[i] = geometry.Point{X: r.Float64() * 1000, Y: r.Float64() * 1000}
		}
	case "clustered":
		centers := []geometry.Point{{X: 250, Y: 250}, {X: 750, Y: 250}, {X: 500, Y: 750}}
		for i := range out {
			c := centers[i%len(centers)]
			out[i] = geometry.Point{X: c.X + r.NormFloat64()*70, Y: c.Y + r.NormFloat64()*70}
		}
	case "circular":
		for i := range out {
			a := r.Float64() * 2 * math.Pi
			radius := 350 + r.Float64()*150
			out[i] = geometry.Point{X: 500 + math.Cos(a)*radius, Y: 500 + math.Sin(a)*radius}
		}
	case "collinear":
		for i := range out {
			x := r.Float64() * 1000
			out[i] = geometry.Point{X: x, Y: 0.5*x + 25}
		}
	default:
		return nil, ErrInvalidDataset
	}
	return out, nil
}
