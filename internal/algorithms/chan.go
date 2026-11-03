package algorithms

import (
	"slices"

	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

// ChansAlgorithm is Chan's output-sensitive convex hull algorithm (1996).
// It runs in O(n log h) time, where h is the number of hull vertices, the
// best possible bound.
//
// It combines two simpler algorithms. For a guess m of the hull size:
//  1. Split the points into groups of at most m and compute each group's hull
//     with a monotone chain scan (a Graham scan variant): O(n log m).
//  2. Gift-wrap (Jarvis March) around the whole set for at most m steps. Each
//     step finds every group's tangent point by binary search: O((n/m) log m)
//     per step, O(n log m) in total.
//
// If the hull doesn't close within m steps the guess was too small, so m is
// squared (4, 16, 256, 65536, …). The final guess is below h², so the total cost
// stays O(n log h).
//
// Output matches GrahamScan: counter-clockwise, starting at the lowest point
// (leftmost on ties), with no duplicates and no collinear boundary points.
type ChansAlgorithm struct{}

func (ChansAlgorithm) Name() string { return "chan" }

func (ChansAlgorithm) Compute(points []geometry.Point) Result {
	hull, _ := chanHull(points)
	return Result{Algorithm: "chan", Status: StatusOK, Hull: hull}
}

// chanStats records the work done. The tests use it to check complexity without
// relying on wall-clock time.
type chanStats struct {
	Orientations int // cross product evaluations
	Comparisons  int // point comparisons while sorting groups
	Rounds       int // hull-size guesses tried
	Fallbacks    int // tangent searches that needed a linear scan (expected 0 without duplicate points)
}

func (s chanStats) Ops() int { return s.Orientations + s.Comparisons }

func chanHull(points []geometry.Point) ([]geometry.Point, chanStats) {
	c := chanState{}
	n := len(points)
	if n == 0 {
		return []geometry.Point{}, c.stats
	}

	c.start = points[0]
	for _, p := range points[1:] {
		if lowerLeft(p, c.start) {
			c.start = p
		}
	}

	// Work on a copy so the caller's slice is never reordered. Building a group's
	// hull needs at most k+1 slots for k points, and groups hold at least 4 points
	// (or there is one group), so n + n/4 + 1 slots always suffice. All buffers are
	// allocated once here and reused by every round.
	c.work = slices.Clone(points)
	maxGroups := n/4 + 1
	c.buf = make([]geometry.Point, 0, n+maxGroups)
	c.offsets = make([]int, 0, maxGroups+1)
	c.out = make([]geometry.Point, 0, 64)

	for t := 1; ; t++ {
		m := n
		if t < 6 { // 2^(2^t) overflows int past t = 5
			if guess := 1 << (1 << t); guess < n {
				m = guess
			}
		}
		c.stats.Rounds++
		if hull, ok := c.round(m); ok {
			return hull, c.stats
		}
	}
}

type chanState struct {
	work    []geometry.Point // copy of the input, sorted within groups
	buf     []geometry.Point // all group hulls, back to back (each counter-clockwise)
	offsets []int            // group g's hull is buf[offsets[g]:offsets[g+1]]
	out     []geometry.Point // the hull being wrapped
	start   geometry.Point   // lowest (then leftmost) point: always on the hull
	stats   chanStats
}

// round tries to build the hull assuming it has at most m vertices.
func (c *chanState) round(m int) ([]geometry.Point, bool) {
	c.buf = c.buf[:0]
	c.offsets = append(c.offsets[:0], 0)
	for lo := 0; lo < len(c.work); lo += m {
		group := c.work[lo:min(lo+m, len(c.work))]
		slices.SortFunc(group, func(a, b geometry.Point) int {
			c.stats.Comparisons++
			return comparePoints(a, b)
		})
		c.buf = c.monotoneChain(c.buf, group)
		c.offsets = append(c.offsets, len(c.buf))
	}
	groups := len(c.offsets) - 1

	// The start point is the lowest point of whichever group holds it, so it
	// is a vertex of that group's hull.
	curG, curI := c.locateStart(m)

	c.out = append(c.out[:0], c.start)
	for step := 0; step < m; step++ {
		p := c.hull(curG)[curI]

		// Each group offers its tangent point; take the most clockwise one.
		bestG, bestI := -1, 0
		var best geometry.Point
		for g := 0; g < groups; g++ {
			q := c.hull(g)
			var i int
			if g == curG {
				i = (curI + 1) % len(q) // p is a vertex here, so the tangent is its successor
			} else {
				i = c.tangent(q, p)
			}
			if q[i] == p {
				continue
			}
			if bestG < 0 || c.moreClockwise(p, q[i], best) {
				best, bestG, bestI = q[i], g, i
			}
		}

		if bestG < 0 || best == c.start {
			return c.out, true // wrapped back to the start: done
		}
		c.out = append(c.out, best)
		curG, curI = bestG, bestI
	}
	return nil, false // more than m vertices: guess again
}

func (c *chanState) hull(g int) []geometry.Point {
	return c.buf[c.offsets[g]:c.offsets[g+1]]
}

func (c *chanState) locateStart(m int) (int, int) {
	g := slices.Index(c.work, c.start) / m
	return g, slices.Index(c.hull(g), c.start)
}

// monotoneChain appends the strict counter-clockwise hull of sorted (by x, then y)
// to dst. Duplicates and collinear points are dropped.
func (c *chanState) monotoneChain(dst, sorted []geometry.Point) []geometry.Point {
	base := len(dst)
	for i, p := range sorted {
		if i > 0 && p == sorted[i-1] {
			continue
		}
		for len(dst)-base >= 2 && c.cross(dst[len(dst)-2], dst[len(dst)-1], p) <= 0 {
			dst = dst[:len(dst)-1]
		}
		dst = append(dst, p)
	}
	lowerEnd := len(dst) - base + 1
	for i := len(sorted) - 2; i >= 0; i-- {
		p := sorted[i]
		if p == sorted[i+1] {
			continue
		}
		for len(dst)-base >= lowerEnd && c.cross(dst[len(dst)-2], dst[len(dst)-1], p) <= 0 {
			dst = dst[:len(dst)-1]
		}
		dst = append(dst, p)
	}
	if len(dst)-base > 1 {
		dst = dst[:len(dst)-1] // the upper chain ends where the lower began
	}
	return dst
}

// tangent returns the index of the vertex of convex polygon q that is most
// clockwise as seen from p (every vertex of q is left of or on p→q[i]), taking
// the farthest on ties. p is a hull vertex of the whole set, so it lies outside q
// or is one of q's vertices (only possible with duplicate input points).
//
// Seen from p, the "clockwise rank" of q's vertices rises to one maximum and falls
// to one minimum around the polygon (ties broken by distance, p itself ranked
// lowest), so the maximum can be found by binary search in O(log k).
func (c *chanState) tangent(q []geometry.Point, p geometry.Point) int {
	k := len(q)
	if k <= 3 {
		return c.tangentLinear(q, p)
	}
	at := func(i int) geometry.Point { return q[i%k] }
	rising := func(i int) bool { return c.moreClockwise(p, at(i+1), at(i)) }

	t := 0
	if rising(0) || c.moreClockwise(p, q[k-1], q[0]) {
		// q[0] is not the maximum. Find the first index at or past the maximum.
		// If the rank rises from q[0], the indices before the maximum are the ones
		// still rising and above q[0]. If it falls from q[0], they're the ones
		// rising, or still below q[0] on the way down.
		up0 := rising(0)
		before := func(i int) bool {
			if up0 {
				return rising(i) && c.moreClockwise(p, q[i], q[0])
			}
			return rising(i) || !c.moreClockwise(p, q[i], q[0])
		}
		lo, hi := 1, k-1
		for lo < hi {
			mid := (lo + hi) / 2
			if before(mid) {
				lo = mid + 1
			} else {
				hi = mid
			}
		}
		t = lo
	}

	// Floating-point rounding can make the ranking inconsistent on nearly
	// degenerate input. Check the answer locally, which is enough on a convex polygon.
	if q[t] != p && !c.moreClockwise(p, at(t+k-1), q[t]) && !c.moreClockwise(p, at(t+1), q[t]) {
		return t
	}
	c.stats.Fallbacks++
	return c.tangentLinear(q, p)
}

func (c *chanState) tangentLinear(q []geometry.Point, p geometry.Point) int {
	best := 0
	for i := 1; i < len(q); i++ {
		if c.moreClockwise(p, q[i], q[best]) {
			best = i
		}
	}
	return best
}

// moreClockwise reports whether u ranks above v as seen from p: u is clockwise
// of v, or in the same direction and farther. p itself ranks below everything.
func (c *chanState) moreClockwise(p, u, v geometry.Point) bool {
	if u == p {
		return false
	}
	if v == p {
		return true
	}
	if o := c.cross(p, v, u); o != 0 {
		return o < 0
	}
	return distSq(p, u) > distSq(p, v)
}

func (c *chanState) cross(o, a, b geometry.Point) float64 {
	c.stats.Orientations++
	return CrossProduct(o, a, b)
}

func distSq(a, b geometry.Point) float64 {
	dx, dy := a.X-b.X, a.Y-b.Y
	return dx*dx + dy*dy
}

func comparePoints(a, b geometry.Point) int {
	switch {
	case a.X < b.X:
		return -1
	case a.X > b.X:
		return 1
	case a.Y < b.Y:
		return -1
	case a.Y > b.Y:
		return 1
	}
	return 0
}

// lowerLeft reports whether a comes before b in (y, then x) order.
func lowerLeft(a, b geometry.Point) bool {
	return a.Y < b.Y || (a.Y == b.Y && a.X < b.X)
}

// rotateToLowest rotates a counter-clockwise hull in place so it starts at the
// lowest (then leftmost) point.
func rotateToLowest(hull []geometry.Point) []geometry.Point {
	start := 0
	for i, p := range hull {
		if lowerLeft(p, hull[start]) {
			start = i
		}
	}
	slices.Reverse(hull[:start])
	slices.Reverse(hull[start:])
	slices.Reverse(hull)
	return hull
}
