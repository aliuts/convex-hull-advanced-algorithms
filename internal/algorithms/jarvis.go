package algorithms

import (
	"slices"

	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

// JarvisMarch is the gift-wrapping algorithm.
//
// Starting from the lowest point (always on the hull), it repeatedly scans every
// point to find the next hull vertex: the one farthest clockwise, so all others
// lie to its left. It stops when it wraps back to the start.
//
// Each step costs O(n) and there is one step per hull vertex, so the total is
// O(nh): fast when the hull is small, O(n²) when every point is on it. It needs
// no sorting and no working copy, so the only memory it uses is the output.
//
// Output matches GrahamScan: counter-clockwise, starting at the lowest point
// (leftmost on ties), with no duplicates and no collinear boundary points.
type JarvisMarch struct{}

func (JarvisMarch) Name() string { return "jarvis_march" }

func (JarvisMarch) Compute(points []geometry.Point) Result {
	hull, _ := jarvisHull(points)
	return Result{Algorithm: "jarvis_march", Status: StatusOK, Hull: hull}
}

// jarvisStats records the work done, so tests can check complexity exactly.
type jarvisStats struct {
	Comparisons int // candidate comparisons: exactly n per hull vertex
	Steps       int // wrapping steps: one per hull vertex
}

func jarvisHull(points []geometry.Point) ([]geometry.Point, jarvisStats) {
	var stats jarvisStats
	if len(points) == 0 {
		return []geometry.Point{}, stats
	}

	start := points[0]
	for _, p := range points[1:] {
		if lowerLeft(p, start) {
			start = p
		}
	}

	hull := make([]geometry.Point, 0, 16)
	p := start
	for len(hull) < len(points) { // a hull can't have more vertices than points
		hull = append(hull, p)
		stats.Steps++

		// Find the point most clockwise from p (farthest on ties), so that every
		// other point is left of or on p→next.
		next := p
		for _, r := range points {
			stats.Comparisons++
			if moreClockwise(p, r, next) {
				next = r
			}
		}
		// Done on wrapping back to the start (or with a single distinct point).
		// On nearly collinear input, rounding can make the ranking inconsistent so
		// the wrap circles back to some other vertex; stop there too. Checking the
		// hull so far is O(h) per step, so the total stays O(nh).
		if next == p || slices.Contains(hull, next) {
			break
		}
		p = next
	}
	return hull, stats
}

// moreClockwise reports whether u ranks above v as seen from p: u is clockwise
// of v, or in the same direction and farther. p itself ranks below everything.
func moreClockwise(p, u, v geometry.Point) bool {
	if u == p {
		return false
	}
	if v == p {
		return true
	}
	if o := CrossProduct(p, v, u); o != 0 {
		return o < 0
	}
	return distSq(p, u) > distSq(p, v)
}
