package algorithms

import "github.com/ali/convex-hull-advanced-algorithms/internal/geometry"

// QuickHull computes the convex hull by divide and conquer.
//
// It splits the points with the line between the leftmost and rightmost points,
// then for each side repeatedly picks the point farthest from the current edge.
// That point is on the hull; everything inside the triangle it forms is dropped,
// and the two new edges are processed the same way.
//
// Expected O(n log n) (close to O(n) when few points are on the hull), worst case O(n²).
//
// Output matches GrahamScan: counter-clockwise, starting at the lowest point
// (leftmost on ties), with no duplicates and no collinear boundary points.
type QuickHull struct{}

func (QuickHull) Name() string { return "quickhull" }

func (QuickHull) Compute(points []geometry.Point) Result {
	hull, _ := quickHull(points)
	return Result{Algorithm: "quickhull", Status: StatusOK, Hull: hull}
}

// quickHull returns the hull and the number of orientation (cross product) tests
// it performed. The count is a hardware-independent measure of work used by the tests.
func quickHull(points []geometry.Point) ([]geometry.Point, int) {
	if len(points) == 0 {
		return []geometry.Point{}, 0
	}

	// Leftmost (lowest on ties) and rightmost (highest on ties) are always hull vertices.
	left, right := points[0], points[0]
	for _, p := range points[1:] {
		if p.X < left.X || (p.X == left.X && p.Y < left.Y) {
			left = p
		}
		if p.X > right.X || (p.X == right.X && p.Y > right.Y) {
			right = p
		}
	}
	if left == right {
		return []geometry.Point{left}, 0
	}

	q := quickHullState{}

	// Partition a working copy (the input is never modified) into points strictly
	// below the left→right line and points strictly above it. Points on the line,
	// including duplicates of left and right, can't be hull vertices and are dropped.
	work := make([]geometry.Point, len(points))
	lower, upper := 0, len(work)
	for _, p := range points {
		d := q.cross(left, right, p)
		switch {
		case d < 0:
			work[lower] = p
			lower++
		case d > 0:
			upper--
			work[upper] = p
		}
	}

	// Counter-clockwise: left, lower chain, right, upper chain.
	q.hull = make([]geometry.Point, 0, 16)
	q.hull = append(q.hull, left)
	q.findHull(left, right, work[:lower])
	q.hull = append(q.hull, right)
	q.findHull(right, left, work[upper:])

	return rotateToLowest(q.hull), q.ops
}

type quickHullState struct {
	hull []geometry.Point
	ops  int
}

func (q *quickHullState) cross(o, a, b geometry.Point) float64 {
	q.ops++
	return CrossProduct(o, a, b)
}

// findHull appends, in order from a to b, the hull vertices among set.
// Every point in set lies strictly to the right of the directed edge a→b.
// set is reordered in place, so no allocation happens during recursion.
func (q *quickHullState) findHull(a, b geometry.Point, set []geometry.Point) {
	if len(set) == 0 {
		return
	}

	// The point farthest to the right of a→b is a hull vertex. Several points can
	// tie on a line parallel to a→b; only the ends of that line are vertices, so
	// on a tie take the one nearest a (a middle one would leave a collinear vertex).
	far, best := 0, 0.0
	dx, dy := b.X-a.X, b.Y-a.Y
	for i, p := range set {
		d := q.cross(b, a, p)
		if d > best || (d == best && p.X*dx+p.Y*dy < set[far].X*dx+set[far].Y*dy) {
			far, best = i, d
		}
	}
	c := set[far]

	// Move points right of a→c to the front and points right of c→b to the back.
	// Points inside triangle a-c-b (or on its edges) are discarded. The two
	// groups can't overlap because c is the farthest point from a→b.
	front, back := 0, len(set)
	for i := 0; i < back; {
		p := set[i]
		if q.cross(c, a, p) > 0 {
			set[front], set[i] = set[i], set[front]
			front++
			i++
		} else if q.cross(b, c, p) > 0 {
			back--
			set[back], set[i] = set[i], set[back]
		} else {
			i++
		}
	}

	q.findHull(a, c, set[:front])
	q.hull = append(q.hull, c)
	q.findHull(c, b, set[back:])
}

// rotateToLowest rotates a counter-clockwise hull in place so it starts at the
// lowest point (leftmost on ties), the same starting point GrahamScan uses.
func rotateToLowest(hull []geometry.Point) []geometry.Point {
	start := 0
	for i, p := range hull {
		s := hull[start]
		if p.Y < s.Y || (p.Y == s.Y && p.X < s.X) {
			start = i
		}
	}
	reverse(hull[:start])
	reverse(hull[start:])
	reverse(hull)
	return hull
}

func reverse(s []geometry.Point) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
