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

func TestQuickHullKnownCases(t *testing.T) {
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
			// Ties on the leftmost and rightmost x must not lose a corner.
			"several points share min and max x",
			pts(0, 0, 0, 5, 0, 10, 10, 0, 10, 5, 10, 10, 5, 5),
			pts(0, 0, 10, 0, 10, 10, 0, 10),
		},
		{
			"two lowest points tie, starts at leftmost",
			pts(6, 0, 2, 0, 4, 5),
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
			got := QuickHull{}.Compute(tc.input)
			if got.Status != StatusOK || got.Algorithm != "quickhull" {
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

func TestQuickHullName(t *testing.T) {
	if QuickHull.Name(QuickHull{}) != "quickhull" {
		t.Fatal("name must match the API identifier")
	}
}

func TestQuickHullDoesNotModifyInput(t *testing.T) {
	input := uniformSquare(500, 1)
	snapshot := append([]geometry.Point(nil), input...)
	QuickHull{}.Compute(input)
	if !reflect.DeepEqual(input, snapshot) {
		t.Fatal("Compute reordered or changed its input slice")
	}
}

func TestQuickHullIsDeterministic(t *testing.T) {
	input := uniformDisk(2000, 9)
	first := QuickHull{}.Compute(input).Hull
	for i := 0; i < 5; i++ {
		if !sameHull(first, QuickHull{}.Compute(input).Hull) {
			t.Fatal("same input produced different hulls")
		}
	}
}

// The hull is a property of the point set, not of its order.
func TestQuickHullIgnoresInputOrder(t *testing.T) {
	input := uniformSquare(1000, 3)
	want := QuickHull{}.Compute(input).Hull
	r := rand.New(rand.NewSource(4))
	for i := 0; i < 10; i++ {
		shuffled := append([]geometry.Point(nil), input...)
		r.Shuffle(len(shuffled), func(a, b int) { shuffled[a], shuffled[b] = shuffled[b], shuffled[a] })
		if got := (QuickHull{}).Compute(shuffled).Hull; !sameHull(got, want) {
			t.Fatalf("shuffle %d changed the hull", i)
		}
	}
}

// Moving or scaling every point should move or scale the hull the same way.
func TestQuickHullTranslationAndScaling(t *testing.T) {
	input := pts(0, 0, 8, 1, 9, 7, 3, 9, -1, 4, 4, 4, 2, 2)
	base := QuickHull{}.Compute(input).Hull
	for _, tf := range []struct{ scale, dx, dy float64 }{{1, 100, -50}, {2, 0, 0}, {0.5, 3, 3}, {16, -1000, 1000}} {
		moved := make([]geometry.Point, len(input))
		for i, p := range input {
			moved[i] = geometry.Point{X: p.X*tf.scale + tf.dx, Y: p.Y*tf.scale + tf.dy}
		}
		got := QuickHull{}.Compute(moved).Hull
		if len(got) != len(base) {
			t.Fatalf("transform %+v changed hull size: %d vs %d", tf, len(got), len(base))
		}
		for i, p := range base {
			want := geometry.Point{X: p.X*tf.scale + tf.dx, Y: p.Y*tf.scale + tf.dy}
			if got[i] != want {
				t.Fatalf("transform %+v: vertex %d = %v, want %v", tf, i, got[i], want)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Correctness: randomised checks against an independent reference
// ---------------------------------------------------------------------------

func TestQuickHullMatchesReferenceOnRandomInputs(t *testing.T) {
	generators := map[string]func(int, int64) []geometry.Point{
		"square": uniformSquare,
		"disk":   uniformDisk,
		"circle": onCircle,
		"grid": func(n int, seed int64) []geometry.Point {
			// Integer grid with heavy duplication and many exactly collinear points.
			r := rand.New(rand.NewSource(seed))
			out := make([]geometry.Point, n)
			for i := range out {
				out[i] = geometry.Point{X: float64(r.Intn(12)), Y: float64(r.Intn(12))}
			}
			return out
		},
	}
	sizes := []int{1, 2, 3, 4, 5, 8, 16, 50, 200, 1000}
	for name, gen := range generators {
		for _, n := range sizes {
			for seed := int64(0); seed < 20; seed++ {
				input := gen(n, seed)
				got := QuickHull{}.Compute(input).Hull
				assertValidHull(t, input, got)
				if want := referenceHull(input); !sameHull(got, want) {
					t.Fatalf("%s n=%d seed=%d:\n got %v\nwant %v", name, n, seed, got, want)
				}
			}
		}
	}
}

// Every dataset kind the UI can generate must produce a valid hull.
func TestQuickHullOnAllDatasetKinds(t *testing.T) {
	for _, kind := range []string{"uniform", "clustered", "circular", "collinear"} {
		for _, n := range []int{0, 1, 2, 3, 10, 100, 1000, dataset.MaxPoints} {
			input, err := dataset.Generate(kind, n, 42)
			if err != nil {
				t.Fatal(err)
			}
			hull := QuickHull{}.Compute(input).Hull
			assertValidHull(t, input, hull)
			if kind == "collinear" {
				// Points are on y = 0.5x + 25 only up to rounding, so the exact vertex
				// count depends on rounding noise (every algorithm gives a different one).
				// What matters is that the hull is a thin sliver along that line.
				for _, p := range hull {
					if off := math.Abs(p.Y - (0.5*p.X + 25)); off > 1e-9 {
						t.Fatalf("collinear n=%d: hull vertex %v is %.2g off the line", n, p, off)
					}
				}
			}
		}
	}
}

// QuickHull and Graham scan follow the same output contract, so they must agree exactly.
func TestQuickHullAgreesWithGrahamScan(t *testing.T) {
	for _, kind := range []string{"uniform", "clustered", "circular"} {
		for seed := int64(1); seed <= 10; seed++ {
			input, _ := dataset.Generate(kind, 2000, seed)
			q := QuickHull{}.Compute(input).Hull
			g := GrahamScan{}.Compute(input).Hull
			if !sameHull(q, g) {
				t.Fatalf("%s seed=%d: quickhull has %d vertices, graham has %d", kind, seed, len(q), len(g))
			}
		}
	}
}

func TestQuickHullWorstCaseInputIsCorrect(t *testing.T) {
	for n := 2; n <= 25; n++ {
		input := quickHullWorstCase(n)
		got := QuickHull{}.Compute(input).Hull
		assertValidHull(t, input, got)
		if len(got) != n {
			t.Fatalf("n=%d: every point is on the hull, got %d vertices", n, len(got))
		}
	}
}

// ---------------------------------------------------------------------------
// Efficiency: hardware-independent operation counts
//
// quickHull reports how many orientation tests it ran. Unlike wall-clock time,
// these counts are exact and repeatable, so the bounds below are reliable.
// ---------------------------------------------------------------------------

func opsFor(points []geometry.Point) int {
	_, ops := quickHull(points)
	return ops
}

// Averaged over several seeds to smooth out randomness in the point sets.
func averageOps(gen func(int, int64) []geometry.Point, n int) float64 {
	const seeds = 5
	total := 0
	for s := int64(0); s < seeds; s++ {
		total += opsFor(gen(n, s))
	}
	return float64(total) / seeds
}

func nLogN(n int) float64 { return float64(n) * math.Log2(float64(n)) }

// When few points are on the hull, almost everything is discarded in the first
// couple of passes, so the work per point stays roughly constant.
func TestQuickHullOpsAreLinearForUniformSquare(t *testing.T) {
	var perPoint []float64
	for _, n := range []int{1000, 10000, 100000} {
		ops := averageOps(uniformSquare, n)
		perPoint = append(perPoint, ops/float64(n))
		t.Logf("square n=%-7d ops=%-10.0f ops/n=%.2f", n, ops, ops/float64(n))
	}
	for i, v := range perPoint {
		if v > 6 {
			t.Errorf("ops per point = %.2f at step %d; expected a small constant", v, i)
		}
	}
	if growth := perPoint[2] / perPoint[0]; growth > 1.5 {
		t.Errorf("ops per point grew %.2fx over a 100x larger input; expected roughly flat", growth)
	}
}

func TestQuickHullOpsForUniformDiskAreNearLinear(t *testing.T) {
	for _, n := range []int{1000, 10000, 100000} {
		ops := averageOps(uniformDisk, n)
		t.Logf("disk   n=%-7d ops=%-10.0f ops/n=%.2f", n, ops, ops/float64(n))
		if ops > 8*float64(n) {
			t.Errorf("n=%d: %.0f ops is more than 8 per point", n, ops)
		}
	}
}

// With every point on the hull nothing can be discarded early. Random splits are
// balanced on average, which gives O(n log n).
func TestQuickHullOpsAreNLogNWhenAllPointsOnHull(t *testing.T) {
	prev := 0.0
	for _, n := range []int{1000, 2000, 4000, 8000, 16000, 32000} {
		ops := averageOps(onCircle, n)
		ratio := ops / nLogN(n)
		t.Logf("circle n=%-6d ops=%-10.0f ops/(n log n)=%.3f", n, ops, ratio)
		if ratio > 3 {
			t.Errorf("n=%d: ops/(n log n) = %.2f, expected a bounded constant", n, ratio)
		}
		if prev > 0 {
			// Doubling n should roughly double the work (2x for linear, ~2.15x for n log n; 4x would be quadratic).
			if g := ops / prev; g > 2.6 {
				t.Errorf("n=%d: doubling n grew work %.2fx, faster than n log n", n, g)
			}
		}
		prev = ops
	}
}

// This documents QuickHull's known weakness: adversarial input forces one-sided
// splits, and the work grows quadratically.
func TestQuickHullOpsAreQuadraticOnWorstCase(t *testing.T) {
	ops6, ops12, ops24 := opsFor(quickHullWorstCase(6)), opsFor(quickHullWorstCase(12)), opsFor(quickHullWorstCase(24))
	t.Logf("worst  n=6 ops=%d  n=12 ops=%d  n=24 ops=%d", ops6, ops12, ops24)
	if g := float64(ops24) / float64(ops12); g < 3.2 {
		t.Errorf("doubling n grew work only %.2fx; expected close to 4x (quadratic)", g)
	}
	if lower := 24 * 23 / 2; ops24 < lower {
		t.Errorf("n=24: %d ops, expected at least n(n-1)/2 = %d", ops24, lower)
	}
}

// Same number of points, but more of them on the hull means more work.
func TestQuickHullWorkGrowsWithHullSize(t *testing.T) {
	const n = 20000
	square, disk, circle := averageOps(uniformSquare, n), averageOps(uniformDisk, n), averageOps(onCircle, n)
	t.Logf("n=%d ops: square=%.0f disk=%.0f circle=%.0f", n, square, disk, circle)
	if !(square < circle && disk < circle) {
		t.Errorf("expected all-on-hull input to cost the most: square=%.0f disk=%.0f circle=%.0f", square, disk, circle)
	}
	if circle < 3*square {
		t.Errorf("expected all-on-hull input to cost several times more than uniform: %.0f vs %.0f", circle, square)
	}
}

// ---------------------------------------------------------------------------
// Efficiency: memory
// ---------------------------------------------------------------------------

// QuickHull partitions one working copy in place. Allocations must not grow
// with the number of points (only a few for the copy and hull slice).
func TestQuickHullAllocationsDoNotGrowWithInput(t *testing.T) {
	allocs := func(input []geometry.Point) float64 {
		return testing.AllocsPerRun(20, func() { QuickHull{}.Compute(input) })
	}
	small, large := allocs(uniformSquare(1000, 1)), allocs(uniformSquare(100000, 1))
	t.Logf("allocations per call: n=1000 → %.0f, n=100000 → %.0f", small, large)
	if large > 6 {
		t.Errorf("n=100000 made %.0f allocations per call; expected a handful", large)
	}
	if large > small+2 {
		t.Errorf("allocations grew from %.0f to %.0f with input size", small, large)
	}
}

func TestQuickHullMemoryIsProportionalToInput(t *testing.T) {
	const n = 100000
	input := uniformSquare(n, 1)
	res := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			QuickHull{}.Compute(input)
		}
	})
	perCall := float64(res.AllocedBytesPerOp())
	pointSize := 16.0 // two float64s
	t.Logf("bytes allocated per call at n=%d: %.0f (%.2f× the input)", n, perCall, perCall/(n*pointSize))
	if perCall > 1.5*n*pointSize {
		t.Errorf("allocated %.0f bytes; expected about one copy of the input (%.0f)", perCall, n*pointSize)
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

func TestQuickHullRuntimeScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test skipped in -short mode")
	}
	small, large := uniformSquare(50000, 1), uniformSquare(800000, 1)
	tSmall := bestTime(7, func() { QuickHull{}.Compute(small) })
	tLarge := bestTime(7, func() { QuickHull{}.Compute(large) })
	ratio := float64(tLarge) / float64(tSmall)
	t.Logf("n=50k: %v  n=800k: %v  ratio %.1f× for 16× more points", tSmall, tLarge, ratio)
	// Linear would be 16x; quadratic would be 256x. Allow room for cache effects.
	if ratio > 48 {
		t.Errorf("16x more points took %.1fx longer; expected near-linear scaling", ratio)
	}
}

// QuickHull skips the full sort Graham scan needs, so on typical data it should be faster.
func TestQuickHullFasterThanGrahamScanOnUniformData(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test skipped in -short mode")
	}
	input := uniformSquare(200000, 7)
	q := bestTime(5, func() { QuickHull{}.Compute(input) })
	g := bestTime(5, func() { GrahamScan{}.Compute(input) })
	t.Logf("n=200k uniform: quickhull %v, graham scan %v (%.1f× faster)", q, g, float64(g)/float64(q))
	if q >= g {
		t.Errorf("quickhull (%v) was not faster than graham scan (%v)", q, g)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: go test ./internal/algorithms -bench . -benchmem
// ---------------------------------------------------------------------------

func BenchmarkQuickHull(b *testing.B) {
	benchmarkHull(b, QuickHull{})
}

// Same inputs as BenchmarkQuickHull, for side-by-side comparison.
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
		{"clustered", func(n int, seed int64) []geometry.Point {
			p, _ := dataset.Generate("clustered", min(n, dataset.MaxPoints), seed)
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

func BenchmarkQuickHullWorstCase(b *testing.B) {
	input := quickHullWorstCase(25)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		QuickHull{}.Compute(input)
	}
}
