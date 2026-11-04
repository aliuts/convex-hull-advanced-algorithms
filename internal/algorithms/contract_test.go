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

// testHullContract runs the correctness checks every algorithm must pass.
// maxN caps the random input sizes so O(nh) algorithms stay quick on all-on-hull inputs.
func testHullContract(t *testing.T, alg ConvexHullAlgorithm, maxN int) {
	t.Run("KnownCases", func(t *testing.T) { testKnownCases(t, alg) })
	t.Run("Registered", func(t *testing.T) {
		for _, a := range All() {
			if a.Name() == alg.Name() {
				return
			}
		}
		t.Fatalf("%s is not registered in All()", alg.Name())
	})
	t.Run("DoesNotModifyInput", func(t *testing.T) {
		input := uniformSquare(5000, 1)
		snapshot := append([]geometry.Point(nil), input...)
		alg.Compute(input)
		if !reflect.DeepEqual(input, snapshot) {
			t.Fatal("Compute reordered or changed its input slice")
		}
	})
	t.Run("Deterministic", func(t *testing.T) {
		input := uniformDisk(2000, 9)
		first := alg.Compute(input).Hull
		for i := 0; i < 5; i++ {
			if !sameHull(first, alg.Compute(input).Hull) {
				t.Fatal("same input produced different hulls")
			}
		}
	})
	t.Run("IgnoresInputOrder", func(t *testing.T) {
		input := uniformDisk(3000, 3)
		want := alg.Compute(input).Hull
		r := rand.New(rand.NewSource(4))
		for i := 0; i < 10; i++ {
			shuffled := append([]geometry.Point(nil), input...)
			r.Shuffle(len(shuffled), func(a, b int) { shuffled[a], shuffled[b] = shuffled[b], shuffled[a] })
			if got := alg.Compute(shuffled).Hull; !sameHull(got, want) {
				t.Fatalf("shuffle %d changed the hull", i)
			}
		}
	})
	t.Run("TranslationAndScaling", func(t *testing.T) {
		input := pts(0, 0, 8, 1, 9, 7, 3, 9, -1, 4, 4, 4, 2, 2, 6, 8, 1, 7)
		base := alg.Compute(input).Hull
		for _, tf := range []struct{ scale, dx, dy float64 }{{1, 100, -50}, {2, 0, 0}, {0.5, 3, 3}, {16, -1000, 1000}} {
			moved := make([]geometry.Point, len(input))
			for i, p := range input {
				moved[i] = geometry.Point{X: p.X*tf.scale + tf.dx, Y: p.Y*tf.scale + tf.dy}
			}
			got := alg.Compute(moved).Hull
			if len(got) != len(base) {
				t.Fatalf("transform %+v changed hull size: %d vs %d", tf, len(got), len(base))
			}
			for i, p := range base {
				if want := (geometry.Point{X: p.X*tf.scale + tf.dx, Y: p.Y*tf.scale + tf.dy}); got[i] != want {
					t.Fatalf("transform %+v: vertex %d = %v, want %v", tf, i, got[i], want)
				}
			}
		}
	})
	t.Run("MatchesReferenceOnRandomInputs", func(t *testing.T) {
		for name, gen := range randomGenerators() {
			for _, n := range []int{1, 2, 3, 4, 5, 8, 16, 17, 50, 200, 1000, 5000} {
				if n > maxN {
					continue
				}
				for seed := int64(0); seed < 15; seed++ {
					input := gen(n, seed)
					got := alg.Compute(input).Hull
					assertValidHull(t, input, got)
					if want := referenceHull(input); !sameHull(got, want) {
						t.Fatalf("%s n=%d seed=%d:\n got %v\nwant %v", name, n, seed, got, want)
					}
				}
			}
		}
	})
	t.Run("ExactCollinear", func(t *testing.T) {
		// Integer points on a line: every cross product is computed exactly, so
		// rounding can't excuse any result other than the two endpoints.
		lines := map[string]func(k float64) geometry.Point{
			"sloped":     func(k float64) geometry.Point { return geometry.Point{X: k, Y: 2*k + 3} },
			"horizontal": func(k float64) geometry.Point { return geometry.Point{X: k, Y: 7} },
			"vertical":   func(k float64) geometry.Point { return geometry.Point{X: 7, Y: k} },
		}
		r := rand.New(rand.NewSource(11))
		for name, line := range lines {
			for _, n := range []int{2, 3, 10, 100, 1000, 10000} {
				input := make([]geometry.Point, n)
				for i := range input {
					input[i] = line(float64(r.Intn(2_000_001) - 1_000_000))
				}
				input = append(input, input[:n/4]...) // duplicates
				r.Shuffle(len(input), func(a, b int) { input[a], input[b] = input[b], input[a] })

				lo, hi := input[0], input[0]
				for _, p := range input {
					if lowerLeft(p, lo) {
						lo = p
					}
					if lowerLeft(hi, p) {
						hi = p
					}
				}
				want := []geometry.Point{lo, hi}
				if lo == hi {
					want = want[:1]
				}
				if got := alg.Compute(input).Hull; !sameHull(got, want) {
					t.Fatalf("%s n=%d: got %v, want endpoints %v", name, n, got, want)
				}
			}
		}
	})
	t.Run("AllDatasetKinds", func(t *testing.T) {
		for _, kind := range []string{"uniform", "clustered", "circular", "collinear"} {
			for _, n := range []int{0, 1, 2, 3, 10, 100, 1000, dataset.MaxPoints} {
				input, err := dataset.Generate(kind, n, 42)
				if err != nil {
					t.Fatal(err)
				}
				hull := alg.Compute(input).Hull
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
	})
	t.Run("AgreesWithGrahamScan", func(t *testing.T) {
		// All algorithms share one output contract, so they must agree exactly.
		for _, kind := range []string{"uniform", "clustered", "circular"} {
			for seed := int64(1); seed <= 10; seed++ {
				input, _ := dataset.Generate(kind, 5000, seed)
				got := alg.Compute(input).Hull
				if g := (GrahamScan{}).Compute(input).Hull; !sameHull(got, g) {
					t.Fatalf("%s seed=%d: %s has %d vertices, graham has %d", kind, seed, alg.Name(), len(got), len(g))
				}
			}
		}
	})
}

func randomGenerators() map[string]func(int, int64) []geometry.Point {
	return map[string]func(int, int64) []geometry.Point{
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
}

func testKnownCases(t *testing.T, alg ConvexHullAlgorithm) {
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
			// The point closest to the start along the first edge must be skipped for the farther one.
			"collinear points on the first edge",
			pts(0, 0, 1, 1, 2, 2, 3, 3, 0, 3),
			pts(0, 0, 3, 3, 0, 3),
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
			"octagon",
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
			got := alg.Compute(tc.input)
			if got.Status != StatusOK || got.Algorithm != alg.Name() {
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

// ---------------------------------------------------------------------------
// Shared timing and benchmark helpers
// ---------------------------------------------------------------------------

func log2(x int) float64 { return math.Log2(float64(x)) }

// bestTime returns the fastest of several runs, which filters out scheduler noise.
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

// benchmarkHull runs alg over a fixed matrix of inputs so results line up across
// algorithms. maxWork skips cases where n·h exceeds it (for O(nh) algorithms);
// 0 means no limit.
func benchmarkHull(b *testing.B, alg ConvexHullAlgorithm, maxWork int) {
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
			if h := len(referenceHull(input)); maxWork > 0 && n*h > maxWork {
				continue
			}
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

func BenchmarkGrahamScan(b *testing.B) {
	benchmarkHull(b, GrahamScan{}, 0)
}

// Same hull sizes for every algorithm, to show how each depends on h.
func benchmarkByHullSize(b *testing.B, alg ConvexHullAlgorithm) {
	const n = 100000
	for _, h := range []int{8, 64, 512, 4096} {
		input := polygonWithInterior(n, h, 1)
		b.Run(fmt.Sprintf("h=%d", h), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				alg.Compute(input)
			}
		})
	}
}

func BenchmarkGrahamScanByHullSize(b *testing.B) { benchmarkByHullSize(b, GrahamScan{}) }
