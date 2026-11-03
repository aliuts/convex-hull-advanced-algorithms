package algorithms

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/ali/convex-hull-advanced-algorithms/internal/dataset"
	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

// ---------------------------------------------------------------------------
// Correctness: hand-made cases with known answers
// ---------------------------------------------------------------------------

func TestChanKnownCases(t *testing.T) {
	cases := []struct {
		name  string
		input []geometry.Point
		want  []geometry.Point
	}{
		{"empty", nil, []geometry.Point{}},
		{"single point", pts(3, 4), pts(3, 4)},
		{"same point repeated", pts(2, 2, 2, 2, 2, 2), pts(2, 2)},
		{"two points", pts(5, 5, 1, 1), pts(1, 1, 5, 5)},
		{"two points with duplicates", pts(5, 5, 1, 1, 5, 5, 1, 1), pts(1, 1, 5, 5)},
		{"horizontal line", pts(3, 0, 1, 0, 4, 0, 2, 0), pts(1, 0, 4, 0)},
		{"vertical line", pts(0, 3, 0, 1, 0, 4, 0, 2), pts(0, 1, 0, 4)},
		{"diagonal line", pts(2, 2, 0, 0, 3, 3, 1, 1), pts(0, 0, 3, 3)},
		{"triangle", pts(0, 0, 4, 0, 2, 3), pts(0, 0, 4, 0, 2, 3)},
		{"clockwise triangle input", pts(2, 3, 4, 0, 0, 0), pts(0, 0, 4, 0, 2, 3)},
		{
			"square with interior points",
			pts(0, 0, 10, 0, 10, 10, 0, 10, 5, 5, 2, 7, 8, 1),
			pts(0, 0, 10, 0, 10, 10, 0, 10),
		},
		{
			"square with points on every edge",
			pts(0, 0, 5, 0, 10, 0, 10, 5, 10, 10, 5, 10, 0, 10, 0, 5),
			pts(0, 0, 10, 0, 10, 10, 0, 10),
		},
		{
			"square with duplicated corners",
			pts(0, 0, 0, 0, 10, 0, 10, 0, 10, 10, 0, 10, 0, 10),
			pts(0, 0, 10, 0, 10, 10, 0, 10),
		},
		{
			"several points share min and max x",
			pts(0, 0, 0, 5, 0, 10, 10, 0, 10, 5, 10, 10, 5, 5),
			pts(0, 0, 10, 0, 10, 10, 0, 10),
		},
		{
			"three points tie for the lowest y",
			pts(6, 0, 2, 0, 4, 0, 4, 5),
			pts(2, 0, 6, 0, 4, 5),
		},
		{
			"negative coordinates",
			pts(-5, -5, 5, -5, 5, 5, -5, 5, 0, 0, -1, 2),
			pts(-5, -5, 5, -5, 5, 5, -5, 5),
		},
		{
			"hexagon",
			pts(2, 0, 4, 0, 6, 2, 4, 4, 2, 4, 0, 2, 3, 2),
			pts(2, 0, 4, 0, 6, 2, 4, 4, 2, 4, 0, 2),
		},
		{
			// 8 vertices: more than the first guess m = 4, so a second round is needed.
			"octagon needs a second round",
			pts(1, 0, 2, 0, 3, 1, 3, 2, 2, 3, 1, 3, 0, 2, 0, 1, 1.5, 1.5, 1, 1, 2, 2, 1, 2, 2, 1),
			pts(1, 0, 2, 0, 3, 1, 3, 2, 2, 3, 1, 3, 0, 2, 0, 1),
		},
		{
			"large coordinates",
			pts(1e9, 1e9, -1e9, 1e9, -1e9, -1e9, 1e9, -1e9, 0, 0),
			pts(-1e9, -1e9, 1e9, -1e9, 1e9, 1e9, -1e9, 1e9),
		},
		{
			"tiny coordinates",
			pts(0, 0, 1e-9, 0, 1e-9, 1e-9, 0, 1e-9, 5e-10, 5e-10),
			pts(0, 0, 1e-9, 0, 1e-9, 1e-9, 0, 1e-9),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ChansAlgorithm{}.Compute(tc.input)
			if got.Status != StatusOK || got.Algorithm != "chan" {
				t.Fatalf("unexpected result metadata: %+v", got)
			}
			if got.Hull == nil {
				t.Fatal("hull must be non-nil so it serialises as [] rather than being omitted")
			}
			if !reflect.DeepEqual(got.Hull, tc.want) {
				t.Fatalf("hull = %v, want %v", got.Hull, tc.want)
			}
			assertValidHull(t, tc.input, got.Hull)
		})
	}
}

func TestChanName(t *testing.T) {
	if (ChansAlgorithm{}).Name() != "chan" {
		t.Fatal("name must match the API identifier")
	}
	found := false
	for _, a := range All() {
		found = found || a.Name() == "chan"
	}
	if !found {
		t.Fatal("ChansAlgorithm is not registered in All()")
	}
}

func TestChanDoesNotModifyInput(t *testing.T) {
	input := uniformSquare(5000, 1)
	snapshot := append([]geometry.Point(nil), input...)
	ChansAlgorithm{}.Compute(input)
	if !reflect.DeepEqual(input, snapshot) {
		t.Fatal("Compute reordered or changed its input slice")
	}
}

func TestChanIsDeterministic(t *testing.T) {
	input := uniformDisk(2000, 9)
	first := ChansAlgorithm{}.Compute(input).Hull
	for i := 0; i < 5; i++ {
		if !sameHull(first, ChansAlgorithm{}.Compute(input).Hull) {
			t.Fatal("same input produced different hulls")
		}
	}
}

// Input order decides which points share a group, but must not change the hull.
func TestChanIgnoresInputOrder(t *testing.T) {
	input := uniformDisk(3000, 3)
	want := ChansAlgorithm{}.Compute(input).Hull
	r := rand.New(rand.NewSource(4))
	for i := 0; i < 10; i++ {
		shuffled := append([]geometry.Point(nil), input...)
		r.Shuffle(len(shuffled), func(a, b int) { shuffled[a], shuffled[b] = shuffled[b], shuffled[a] })
		if got := (ChansAlgorithm{}).Compute(shuffled).Hull; !sameHull(got, want) {
			t.Fatalf("shuffle %d changed the hull", i)
		}
	}
}

func TestChanTranslationAndScaling(t *testing.T) {
	input := pts(0, 0, 8, 1, 9, 7, 3, 9, -1, 4, 4, 4, 2, 2, 6, 8, 1, 7)
	base := ChansAlgorithm{}.Compute(input).Hull
	for _, tf := range []struct{ scale, dx, dy float64 }{{1, 100, -50}, {2, 0, 0}, {0.5, 3, 3}, {16, -1000, 1000}} {
		moved := make([]geometry.Point, len(input))
		for i, p := range input {
			moved[i] = geometry.Point{X: p.X*tf.scale + tf.dx, Y: p.Y*tf.scale + tf.dy}
		}
		got := ChansAlgorithm{}.Compute(moved).Hull
		if len(got) != len(base) {
			t.Fatalf("transform %+v changed hull size: %d vs %d", tf, len(got), len(base))
		}
		for i, p := range base {
			if want := (geometry.Point{X: p.X*tf.scale + tf.dx, Y: p.Y*tf.scale + tf.dy}); got[i] != want {
				t.Fatalf("transform %+v: vertex %d = %v, want %v", tf, i, got[i], want)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Correctness: the round structure (guesses m = 4, 16, 256, 65536, …)
// ---------------------------------------------------------------------------

// A hull with h vertices needs the first guess m >= h. Sizes just either side
// of each guess check that rounds stop and restart at the right moment.
func TestChanRoundsMatchHullSize(t *testing.T) {
	cases := []struct{ h, rounds int }{
		{3, 1}, {4, 1}, {5, 2}, {16, 2}, {17, 3}, {256, 3}, {257, 4}, {1000, 4},
	}
	for _, tc := range cases {
		input := polygonWithInterior(70000, tc.h, 1)
		hull, stats := chanHull(input)
		assertValidHull(t, input, hull)
		if len(hull) != tc.h {
			t.Fatalf("h=%d: got %d vertices", tc.h, len(hull))
		}
		if stats.Rounds != tc.rounds {
			t.Errorf("h=%d: used %d rounds, want %d", tc.h, stats.Rounds, tc.rounds)
		}
	}
}

// When the input is smaller than the next guess, that guess is capped at n
// (one group covering everything), which always succeeds.
func TestChanSmallInputsFinishWhenGuessReachesN(t *testing.T) {
	for n := 1; n <= 40; n++ {
		input := onCircle(n, int64(n))
		hull, stats := chanHull(input)
		assertValidHull(t, input, hull)
		if want := referenceHull(input); !sameHull(hull, want) {
			t.Fatalf("n=%d: got %v, want %v", n, hull, want)
		}
		if stats.Rounds > 3 {
			t.Fatalf("n=%d: %d rounds", n, stats.Rounds)
		}
	}
}

// ---------------------------------------------------------------------------
// Correctness: the binary-search tangent, the trickiest part of the algorithm
// ---------------------------------------------------------------------------

func TestChanTangentMatchesLinearScan(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	checked := 0
	for trial := 0; trial < 3000; trial++ {
		var cloud []geometry.Point
		if trial%2 == 0 {
			cloud = uniformDisk(4+r.Intn(200), int64(trial))
		} else {
			// Integer grid: many points collinear with polygon edges, which
			// exercises the "same direction, take the farther one" tie-break.
			cloud = make([]geometry.Point, 4+r.Intn(60))
			for i := range cloud {
				cloud[i] = geometry.Point{X: float64(r.Intn(20)), Y: float64(r.Intn(20))}
			}
		}
		poly := referenceHull(cloud)
		if len(poly) < 4 {
			continue
		}
		poly = rotateRandomly(poly, r) // the search must not depend on where the polygon starts

		for k := 0; k < 20; k++ {
			var p geometry.Point
			if k < 4 {
				p = poly[r.Intn(len(poly))] // p is one of the polygon's own vertices
			} else if trial%2 == 0 {
				p = geometry.Point{X: r.Float64()*3000 - 1000, Y: r.Float64()*3000 - 1000}
			} else {
				p = geometry.Point{X: float64(r.Intn(60) - 20), Y: float64(r.Intn(60) - 20)}
			}
			if strictlyInside(poly, p) || onBoundaryNotVertex(poly, p) {
				continue // never happens in the algorithm: p is always a hull vertex of the whole set
			}
			c := chanState{}
			got := c.tangent(poly, p)
			want := c.tangentLinear(poly, p)
			if poly[got] != poly[want] {
				t.Fatalf("tangent from %v to %v: binary search gave %v, linear scan gave %v", p, poly, poly[got], poly[want])
			}
			if c.stats.Fallbacks != 0 {
				t.Fatalf("tangent from %v to %v needed the linear fallback", p, poly)
			}
			checked++
		}
	}
	if checked < 20000 {
		t.Fatalf("only %d tangent queries were checked", checked)
	}
}

func TestChanTangentRespectsAllVertices(t *testing.T) {
	poly := referenceHull(onCircle(500, 2))
	r := rand.New(rand.NewSource(5))
	for i := 0; i < 2000; i++ {
		a := r.Float64() * 2 * math.Pi
		d := 501 + r.Float64()*5000
		p := geometry.Point{X: 500 + d*math.Cos(a), Y: 500 + d*math.Sin(a)}
		c := chanState{}
		q := poly[c.tangent(poly, p)]
		for _, v := range poly {
			if CrossProduct(p, q, v) < 0 {
				t.Fatalf("vertex %v is right of tangent %v→%v", v, p, q)
			}
		}
	}
}

func rotateRandomly(poly []geometry.Point, r *rand.Rand) []geometry.Point {
	s := r.Intn(len(poly))
	return append(append([]geometry.Point{}, poly[s:]...), poly[:s]...)
}

func strictlyInside(poly []geometry.Point, p geometry.Point) bool {
	for i := range poly {
		if CrossProduct(poly[i], poly[(i+1)%len(poly)], p) <= 0 {
			return false
		}
	}
	return true
}

func onBoundaryNotVertex(poly []geometry.Point, p geometry.Point) bool {
	for i := range poly {
		a, b := poly[i], poly[(i+1)%len(poly)]
		if p == a || p == b {
			return false
		}
		if CrossProduct(a, b, p) == 0 &&
			p.X >= math.Min(a.X, b.X) && p.X <= math.Max(a.X, b.X) &&
			p.Y >= math.Min(a.Y, b.Y) && p.Y <= math.Max(a.Y, b.Y) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Correctness: randomised checks against an independent reference
// ---------------------------------------------------------------------------

func TestChanMatchesReferenceOnRandomInputs(t *testing.T) {
	generators := map[string]func(int, int64) []geometry.Point{
		"square": uniformSquare,
		"disk":   uniformDisk,
		"circle": onCircle,
		"polygon-12": func(n int, seed int64) []geometry.Point {
			return polygonWithInterior(max(n, 12), 12, seed)
		},
		"grid": func(n int, seed int64) []geometry.Point {
			// Integer grid: heavy duplication and many exactly collinear points.
			r := rand.New(rand.NewSource(seed))
			out := make([]geometry.Point, n)
			for i := range out {
				out[i] = geometry.Point{X: float64(r.Intn(12)), Y: float64(r.Intn(12))}
			}
			return out
		},
	}
	sizes := []int{1, 2, 3, 4, 5, 8, 16, 17, 50, 200, 1000, 5000}
	for name, gen := range generators {
		for _, n := range sizes {
			for seed := int64(0); seed < 15; seed++ {
				input := gen(n, seed)
				got, stats := chanHull(input)
				assertValidHull(t, input, got)
				if want := referenceHull(input); !sameHull(got, want) {
					t.Fatalf("%s n=%d seed=%d:\n got %v\nwant %v", name, n, seed, got, want)
				}
				// Without duplicate points, the binary search should always succeed on its own.
				if name != "grid" && stats.Fallbacks != 0 {
					t.Fatalf("%s n=%d seed=%d: %d tangent fallbacks", name, n, seed, stats.Fallbacks)
				}
			}
		}
	}
}

func TestChanOnAllDatasetKinds(t *testing.T) {
	for _, kind := range []string{"uniform", "clustered", "circular", "collinear"} {
		for _, n := range []int{0, 1, 2, 3, 10, 100, 1000, dataset.MaxPoints} {
			input, err := dataset.Generate(kind, n, 42)
			if err != nil {
				t.Fatal(err)
			}
			hull := ChansAlgorithm{}.Compute(input).Hull
			assertValidHull(t, input, hull)
			if kind == "collinear" {
				// Points are on y = 0.5x + 25 only up to rounding, so the exact vertex
				// count depends on rounding noise. The hull must be a thin sliver along the line.
				for _, p := range hull {
					if off := math.Abs(p.Y - (0.5*p.X + 25)); off > 1e-9 {
						t.Fatalf("collinear n=%d: hull vertex %v is %.2g off the line", n, p, off)
					}
				}
			}
		}
	}
}

// Chan's algorithm and Graham scan follow the same output contract, so they must agree exactly.
func TestChanAgreesWithGrahamScan(t *testing.T) {
	for _, kind := range []string{"uniform", "clustered", "circular"} {
		for seed := int64(1); seed <= 10; seed++ {
			input, _ := dataset.Generate(kind, 5000, seed)
			c := ChansAlgorithm{}.Compute(input).Hull
			g := GrahamScan{}.Compute(input).Hull
			if !sameHull(c, g) {
				t.Fatalf("%s seed=%d: chan has %d vertices, graham has %d", kind, seed, len(c), len(g))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Efficiency: hardware-independent operation counts
//
// chanHull counts orientation tests and sort comparisons. Unlike wall-clock
// time these are exact and repeatable, so the bounds below are reliable.
// ---------------------------------------------------------------------------

func chanOps(points []geometry.Point) chanStats {
	_, stats := chanHull(points)
	return stats
}

func log2(x int) float64 { return math.Log2(float64(x)) }

// The defining property: with h fixed, the work is linear in n, not n log n.
func TestChanWorkIsLinearInNForFixedHullSize(t *testing.T) {
	const h = 16
	var perPoint []float64
	for _, n := range []int{1000, 8000, 64000, 512000} {
		s := chanOps(polygonWithInterior(n, h, 1))
		perPoint = append(perPoint, float64(s.Ops())/float64(n))
		t.Logf("h=%d n=%-7d ops=%-9d ops/n=%.2f rounds=%d", h, n, s.Ops(), float64(s.Ops())/float64(n), s.Rounds)
	}
	// n grows 512x. O(n log n) would push ops/n up by about 1.9x (log 512000 / log 1000);
	// O(n log h) keeps it flat.
	if g := perPoint[len(perPoint)-1] / perPoint[0]; g > 1.25 {
		t.Errorf("ops per point grew %.2fx while h stayed at %d; expected flat", g, h)
	}
}

// With n fixed, the work grows with log h, far slower than h itself.
func TestChanWorkGrowsWithLogOfHullSize(t *testing.T) {
	const n = 65536
	type row struct {
		h   int
		ops float64
	}
	var rows []row
	for _, h := range []int{8, 64, 512, 4096, 32768} {
		s := chanOps(polygonWithInterior(n, h, 1))
		rows = append(rows, row{h, float64(s.Ops())})
		t.Logf("n=%d h=%-6d ops=%-9d ops/(n log h)=%.2f rounds=%d", n, h, s.Ops(), float64(s.Ops())/(n*log2(h)), s.Rounds)
	}
	for _, r := range rows {
		if c := r.ops / (n * log2(r.h)); c > 12 {
			t.Errorf("h=%d: ops/(n log h) = %.2f, expected a bounded constant", r.h, c)
		}
	}
	// h grows 4096x (8 → 32768). log h grows 5x. Work should be in that range, not near 4096x.
	if g := rows[len(rows)-1].ops / rows[0].ops; g > 12 {
		t.Errorf("work grew %.1fx while h grew 4096x; expected roughly log h growth", g)
	}
}

// Every point on the hull (h = n) is the hardest case and costs O(n log n).
func TestChanWorkIsNLogNWhenAllPointsOnHull(t *testing.T) {
	for _, n := range []int{1000, 4000, 16000, 64000} {
		s := chanOps(onCircle(n, 1))
		c := float64(s.Ops()) / (float64(n) * log2(n))
		t.Logf("circle n=%-6d ops=%-9d ops/(n log n)=%.2f rounds=%d", n, s.Ops(), c, s.Rounds)
		if c > 10 {
			t.Errorf("n=%d: ops/(n log n) = %.2f, expected a bounded constant", n, c)
		}
	}
}

// Uniform points in a square have only O(log n) hull vertices, so Chan's
// algorithm should be close to linear.
func TestChanWorkIsNearLinearForUniformSquare(t *testing.T) {
	var perPoint []float64
	for _, n := range []int{1000, 10000, 100000} {
		s := chanOps(uniformSquare(n, 1))
		perPoint = append(perPoint, float64(s.Ops())/float64(n))
		t.Logf("square n=%-7d ops=%-9d ops/n=%.2f rounds=%d", n, s.Ops(), perPoint[len(perPoint)-1], s.Rounds)
	}
	if g := perPoint[2] / perPoint[0]; g > 1.4 {
		t.Errorf("ops per point grew %.2fx over a 100x larger input", g)
	}
}

// Squaring the guess means only about log log h rounds are ever needed.
func TestChanRoundCountIsLogLogH(t *testing.T) {
	for _, h := range []int{3, 10, 100, 1000, 30000} {
		s := chanOps(polygonWithInterior(70000, h, 2))
		limit := int(math.Ceil(math.Log2(math.Log2(float64(h))))) + 1
		if s.Rounds > limit {
			t.Errorf("h=%d: %d rounds, expected at most %d", h, s.Rounds, limit)
		}
	}
}

// ---------------------------------------------------------------------------
// Efficiency: memory
// ---------------------------------------------------------------------------

// Buffers are allocated once and reused across rounds, so allocations stay small
// and don't grow with n.
func TestChanAllocationsDoNotGrowWithInput(t *testing.T) {
	allocs := func(input []geometry.Point) float64 {
		return testing.AllocsPerRun(10, func() { ChansAlgorithm{}.Compute(input) })
	}
	small, large := allocs(polygonWithInterior(1000, 16, 1)), allocs(polygonWithInterior(200000, 16, 1))
	t.Logf("allocations per call (h=16): n=1000 → %.0f, n=200000 → %.0f", small, large)
	if large > 6 {
		t.Errorf("n=200000 made %.0f allocations per call; expected a handful", large)
	}
	if large > small+2 {
		t.Errorf("allocations grew from %.0f to %.0f with input size", small, large)
	}
}

// One working copy (n points) plus group-hull storage (n + n/4 points): about 2.3x the input.
func TestChanMemoryIsProportionalToInput(t *testing.T) {
	const n = 100000
	input := uniformSquare(n, 1)
	res := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ChansAlgorithm{}.Compute(input)
		}
	})
	perCall := float64(res.AllocedBytesPerOp())
	inputBytes := float64(n * 16) // a point is two float64s
	t.Logf("bytes allocated per call at n=%d: %.0f (%.2fx the input)", n, perCall, perCall/inputBytes)
	if perCall > 2.6*inputBytes {
		t.Errorf("allocated %.0f bytes; expected about 2.3x the input (%.0f)", perCall, 2.3*inputBytes)
	}
}

// ---------------------------------------------------------------------------
// Efficiency: wall-clock time (skipped with -short; uses the best of several
// runs and wide margins so it stays stable on busy machines)
// ---------------------------------------------------------------------------

func bestTime(runs int, f func()) time.Duration {
	best := time.Duration(math.MaxInt64)
	for i := 0; i < runs; i++ {
		start := time.Now()
		f()
		if d := time.Since(start); d < best {
			best = d
		}
	}
	return best
}

func TestChanRuntimeIsLinearForFixedHullSize(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test skipped in -short mode")
	}
	small, large := polygonWithInterior(50000, 16, 1), polygonWithInterior(800000, 16, 1)
	tSmall := bestTime(7, func() { ChansAlgorithm{}.Compute(small) })
	tLarge := bestTime(5, func() { ChansAlgorithm{}.Compute(large) })
	ratio := float64(tLarge) / float64(tSmall)
	t.Logf("h=16  n=50k: %v  n=800k: %v  ratio %.1fx for 16x more points", tSmall, tLarge, ratio)
	// Linear is 16x; n log n would be about 20x on top of cache effects. Allow headroom for noise.
	if ratio > 40 {
		t.Errorf("16x more points took %.1fx longer; expected near-linear scaling", ratio)
	}
}

// Graham scan always sorts all n points. Chan's algorithm only sorts small groups
// when the hull is small, so it should win clearly on typical data.
func TestChanFasterThanGrahamScanWhenHullIsSmall(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test skipped in -short mode")
	}
	for name, input := range map[string][]geometry.Point{
		"uniform square": uniformSquare(200000, 7),
		"16-gon":         polygonWithInterior(200000, 16, 7),
	} {
		c := bestTime(5, func() { ChansAlgorithm{}.Compute(input) })
		g := bestTime(5, func() { GrahamScan{}.Compute(input) })
		t.Logf("n=200k %s: chan %v, graham scan %v (%.1fx faster)", name, c, g, float64(g)/float64(c))
		if c >= g {
			t.Errorf("%s: chan (%v) was not faster than graham scan (%v)", name, c, g)
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: go test ./internal/algorithms -run '^$' -bench . -benchmem
// ---------------------------------------------------------------------------

func BenchmarkChan(b *testing.B) {
	benchmarkHull(b, ChansAlgorithm{})
}

// Same inputs as BenchmarkChan, for side-by-side comparison.
func BenchmarkGrahamScan(b *testing.B) {
	benchmarkHull(b, GrahamScan{})
}

func benchmarkHull(b *testing.B, alg ConvexHullAlgorithm) {
	generators := []struct {
		name string
		gen  func(int, int64) []geometry.Point
	}{
		{"square", uniformSquare},
		{"disk", uniformDisk},
		{"circle", onCircle},
		{"16-gon", func(n int, seed int64) []geometry.Point { return polygonWithInterior(n, 16, seed) }},
		{"clustered", func(n int, seed int64) []geometry.Point {
			p, _ := dataset.Generate("clustered", n, seed)
			return p
		}},
	}
	for _, g := range generators {
		for _, n := range []int{100, 1000, 10000, 100000} {
			if g.name == "clustered" && n > dataset.MaxPoints {
				continue
			}
			input := g.gen(n, 1)
			b.Run(fmt.Sprintf("%s/n=%d", g.name, n), func(b *testing.B) {
				b.ReportAllocs()
				var hull int
				for i := 0; i < b.N; i++ {
					hull = len(alg.Compute(input).Hull)
				}
				b.ReportMetric(float64(hull), "hull-vertices")
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(n), "ns/point")
			})
		}
	}
}

// Fixed n, growing h: shows the log h dependence directly.
func BenchmarkChanByHullSize(b *testing.B) {
	const n = 100000
	for _, h := range []int{8, 64, 512, 4096, 32768} {
		input := polygonWithInterior(n, h, 1)
		b.Run(fmt.Sprintf("h=%d", h), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ChansAlgorithm{}.Compute(input)
			}
		})
	}
}
