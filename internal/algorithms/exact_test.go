package algorithms

import (
	"math"
	"math/big"
	"slices"
	"testing"

	"github.com/ali/convex-hull-advanced-algorithms/internal/dataset"
	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

// The algorithms use float64 cross products, which can get the sign wrong when
// points are almost collinear. These tests compare them against the exact hull
// of the same float coordinates, computed with exact rational arithmetic.

// exactCross returns the cross product (a-o)×(b-o) with no rounding error.
// big.Rat.SetFloat64 converts a float64 exactly.
func exactCross(o, a, b geometry.Point) *big.Rat {
	r := func(v float64) *big.Rat { return new(big.Rat).SetFloat64(v) }
	ax, ay := new(big.Rat).Sub(r(a.X), r(o.X)), new(big.Rat).Sub(r(a.Y), r(o.Y))
	bx, by := new(big.Rat).Sub(r(b.X), r(o.X)), new(big.Rat).Sub(r(b.Y), r(o.Y))
	return new(big.Rat).Sub(new(big.Rat).Mul(ax, by), new(big.Rat).Mul(ay, bx))
}

func exactOrient(o, a, b geometry.Point) int { return exactCross(o, a, b).Sign() }

// exactHull is the true convex hull of the given coordinates (monotone chain with
// exact orientation tests), in the same canonical form the algorithms return.
func exactHull(points []geometry.Point) []geometry.Point {
	ps := slices.Clone(points)
	slices.SortFunc(ps, comparePoints)
	ps = slices.Compact(ps)
	if len(ps) <= 1 {
		return append([]geometry.Point{}, ps...)
	}
	hull := make([]geometry.Point, 0, 2*len(ps))
	for _, p := range ps {
		for len(hull) >= 2 && exactOrient(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	lowerLen := len(hull) + 1
	for i := len(ps) - 2; i >= 0; i-- {
		p := ps[i]
		for len(hull) >= lowerLen && exactOrient(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	return rotateToLowest(hull[:len(hull)-1])
}

// maxOutside returns how far (perpendicular distance) the input point farthest
// outside the hull lies beyond the hull's edge lines. 0 means every point is
// inside or on the hull. A 2-vertex hull is a segment, so any distance off its line counts.
func maxOutside(input, hull []geometry.Point) float64 {
	worst := 0.0
	h := len(hull)
	if h < 2 {
		return worst
	}
	for i := 0; i < h; i++ {
		a, b := hull[i], hull[(i+1)%h]
		length := math.Hypot(b.X-a.X, b.Y-a.Y)
		for _, p := range input {
			c, _ := exactCross(a, b, p).Float64()
			if h == 2 {
				c = -math.Abs(c)
			}
			if d := -c / length; d > worst {
				worst = d
			}
		}
		if h == 2 {
			break // both directions of a segment are the same line
		}
	}
	return worst
}

// On ordinary (non-degenerate) data every algorithm must return exactly the true hull.
func TestAlgorithmsMatchExactHull(t *testing.T) {
	check := func(name string, input []geometry.Point) {
		want := exactHull(input)
		for _, alg := range All() {
			if got := alg.Compute(input).Hull; !sameHull(got, want) {
				t.Fatalf("%s %s: %d vertices, exact hull has %d\n got %v\nwant %v",
					alg.Name(), name, len(got), len(want), got, want)
			}
		}
	}
	for _, kind := range []string{"uniform", "clustered", "circular"} {
		for _, n := range []int{3, 100, 1000, 10000} {
			for seed := int64(1); seed <= 3; seed++ {
				input, _ := dataset.Generate(kind, n, seed)
				check(kind, input)
			}
		}
	}
	for name, gen := range randomGenerators() {
		for _, n := range []int{5, 50, 1000} {
			for seed := int64(0); seed < 5; seed++ {
				check(name, gen(n, seed))
			}
		}
	}
}

// The "collinear" dataset is the line y = 0.5x + 25, but rounding pushes points
// up to one rounding step off it, so the true hull is a very thin polygon.
// float64 orientation tests can't resolve it reliably, so the algorithms
// approximate it. This measures how close each one gets; the log is the table
// used in the report.
func TestCollinearDatasetAgainstExactHull(t *testing.T) {
	for _, n := range []int{100, 500, 1000, 5000, 10000} {
		input, _ := dataset.Generate("collinear", n, 42)
		exact := exactHull(input)
		isExactVertex := make(map[geometry.Point]bool, len(exact))
		for _, p := range exact {
			isExactVertex[p] = true
		}
		t.Logf("n=%-5d exact hull: %d vertices", n, len(exact))
		for _, alg := range All() {
			hull := alg.Compute(input).Hull
			matching := 0
			for _, p := range hull {
				if isExactVertex[p] {
					matching++
				}
			}
			outside := maxOutside(input, hull)
			t.Logf("        %-12s %2d vertices, %2d of them true hull vertices, farthest point outside: %.1e",
				alg.Name(), len(hull), matching, outside)
			// Coordinates are up to ~1000, where one rounding step is ~1e-13.
			if outside > 1e-9 {
				t.Errorf("%s n=%d: a point lies %.2g outside the hull", alg.Name(), n, outside)
			}
		}
	}
}
