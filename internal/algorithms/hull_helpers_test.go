package algorithms

import (
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

// Shared helpers for testing any ConvexHullAlgorithm.

func pts(coords ...float64) []geometry.Point {
	out := make([]geometry.Point, 0, len(coords)/2)
	for i := 0; i+1 < len(coords); i += 2 {
		out = append(out, geometry.Point{X: coords[i], Y: coords[i+1]})
	}
	return out
}

// assertValidHull checks the full output contract:
//   - every hull vertex comes from the input and appears once
//   - the hull starts at the lowest (then leftmost) input point
//   - consecutive vertices turn left (convex, counter-clockwise), up to rounding error
//   - every input point is inside or on the boundary
func assertValidHull(t *testing.T, input, hull []geometry.Point) {
	t.Helper()
	if len(input) == 0 {
		if len(hull) != 0 {
			t.Fatalf("empty input produced hull %v", hull)
		}
		return
	}

	inInput := make(map[geometry.Point]bool, len(input))
	for _, p := range input {
		inInput[p] = true
	}
	seen := make(map[geometry.Point]bool, len(hull))
	for _, p := range hull {
		if !inInput[p] {
			t.Fatalf("hull vertex %v is not an input point", p)
		}
		if seen[p] {
			t.Fatalf("hull vertex %v appears twice in %v", p, hull)
		}
		seen[p] = true
	}

	lowest, _ := FindPivot(input)
	if len(hull) == 0 || hull[0] != lowest {
		t.Fatalf("hull should start at lowest point %v, got %v", lowest, hull)
	}

	if len(inInput) > 1 && len(hull) < 2 {
		t.Fatalf("%d distinct points produced hull %v", len(inInput), hull)
	}

	// Tolerance for floating-point rounding, scaled to the coordinate range.
	eps := 1e-9 * math.Max(1, coordScale(input)*coordScale(input))

	h := len(hull)
	if h >= 3 {
		for i := 0; i < h; i++ {
			a, b, c := hull[i], hull[(i+1)%h], hull[(i+2)%h]
			// Allow rounding error here; exact strictness is checked against referenceHull.
			if CrossProduct(a, b, c) < -eps {
				t.Fatalf("hull is not convex counter-clockwise at %v,%v,%v", a, b, c)
			}
		}
		for _, p := range input {
			for i := 0; i < h; i++ {
				if CrossProduct(hull[i], hull[(i+1)%h], p) < -eps {
					t.Fatalf("input point %v lies outside hull edge %v→%v", p, hull[i], hull[(i+1)%h])
				}
			}
		}
	}
	if h == 2 {
		// Degenerate hull: every point must lie on the segment.
		a, b := hull[0], hull[1]
		for _, p := range input {
			if math.Abs(CrossProduct(a, b, p)) > eps*10 {
				t.Fatalf("point %v is off the segment hull %v→%v", p, a, b)
			}
		}
	}
}

func coordScale(points []geometry.Point) float64 {
	m := 0.0
	for _, p := range points {
		m = math.Max(m, math.Max(math.Abs(p.X), math.Abs(p.Y)))
	}
	return m
}

// referenceHull is Andrew's monotone chain, an independent algorithm used as an
// oracle. It returns the same canonical form the algorithms must produce.
func referenceHull(points []geometry.Point) []geometry.Point {
	ps := append([]geometry.Point(nil), points...)
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].X != ps[j].X {
			return ps[i].X < ps[j].X
		}
		return ps[i].Y < ps[j].Y
	})
	uniq := ps[:0]
	for i, p := range ps {
		if i == 0 || p != ps[i-1] {
			uniq = append(uniq, p)
		}
	}
	if len(uniq) <= 1 {
		return append([]geometry.Point{}, uniq...)
	}
	hull := make([]geometry.Point, 0, 2*len(uniq))
	for _, p := range uniq {
		for len(hull) >= 2 && CrossProduct(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	lowerLen := len(hull) + 1
	for i := len(uniq) - 2; i >= 0; i-- {
		p := uniq[i]
		for len(hull) >= lowerLen && CrossProduct(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	return rotateToLowest(hull[:len(hull)-1])
}

func sameHull(a, b []geometry.Point) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Point generators for tests and benchmarks (dataset.Generate caps at 10,000 points).

func uniformSquare(n int, seed int64) []geometry.Point {
	r := rand.New(rand.NewSource(seed))
	out := make([]geometry.Point, n)
	for i := range out {
		out[i] = geometry.Point{X: r.Float64() * 1000, Y: r.Float64() * 1000}
	}
	return out
}

func uniformDisk(n int, seed int64) []geometry.Point {
	r := rand.New(rand.NewSource(seed))
	out := make([]geometry.Point, n)
	for i := range out {
		a := r.Float64() * 2 * math.Pi
		rad := 500 * math.Sqrt(r.Float64())
		out[i] = geometry.Point{X: 500 + rad*math.Cos(a), Y: 500 + rad*math.Sin(a)}
	}
	return out
}

// onCircle puts every point on the hull, the hardest typical case for output-sensitive work.
func onCircle(n int, seed int64) []geometry.Point {
	r := rand.New(rand.NewSource(seed))
	out := make([]geometry.Point, n)
	for i := range out {
		a := r.Float64() * 2 * math.Pi
		out[i] = geometry.Point{X: 500 + 500*math.Cos(a), Y: 500 + 500*math.Sin(a)}
	}
	return out
}

// quickHullWorstCase builds points on the parabola y = -x² at x = 0 and x = 2^-k.
// The farthest point from each chord is always the next one in the sequence,
// so every QuickHull split removes just one point: O(n²) work.
// Powers of two keep the arithmetic exact for n up to about 25.
func quickHullWorstCase(n int) []geometry.Point {
	out := []geometry.Point{{X: 0, Y: 0}}
	x := 1.0
	for k := 0; k < n-1; k++ {
		out = append(out, geometry.Point{X: x, Y: -x * x})
		x /= 2
	}
	return out
}
