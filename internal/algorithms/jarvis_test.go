package algorithms

import (
	"testing"

	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

// Shared correctness checks (known cases, random inputs vs reference, dataset kinds, …).
func TestJarvisContract(t *testing.T) {
	testHullContract(t, JarvisMarch{}, 5000)
}

func TestJarvisName(t *testing.T) {
	if (JarvisMarch{}).Name() != "jarvis_march" {
		t.Fatal("name must match the API identifier")
	}
}

// ---------------------------------------------------------------------------
// Efficiency: exact operation counts
//
// Jarvis March compares every point once per hull vertex, so its work is
// exactly n·h comparisons. The tests check that exactly.
// ---------------------------------------------------------------------------

func TestJarvisWorkIsExactlyNTimesH(t *testing.T) {
	inputs := map[string][]geometry.Point{
		"square":   uniformSquare(20000, 1),
		"disk":     uniformDisk(20000, 1),
		"circle":   onCircle(3000, 1),
		"8-gon":    polygonWithInterior(20000, 8, 1),
		"300-gon":  polygonWithInterior(20000, 300, 1),
		"triangle": pts(0, 0, 4, 0, 2, 3, 2, 1),
		"single":   pts(1, 1, 1, 1),
	}
	for name, input := range inputs {
		hull, stats := jarvisHull(input)
		n, h := len(input), len(hull)
		if stats.Steps != h {
			t.Errorf("%s: %d steps for %d hull vertices", name, stats.Steps, h)
		}
		if stats.Comparisons != n*h {
			t.Errorf("%s: %d comparisons, want n·h = %d·%d = %d", name, stats.Comparisons, n, h, n*h)
		}
	}
}

// With h fixed, work grows linearly with n.
func TestJarvisWorkIsLinearInNForFixedHullSize(t *testing.T) {
	const h = 16
	for _, n := range []int{1000, 8000, 64000, 512000} {
		_, s := jarvisHull(polygonWithInterior(n, h, 1))
		t.Logf("h=%d n=%-7d comparisons=%-9d per point=%d", h, n, s.Comparisons, s.Comparisons/n)
		if s.Comparisons != n*h {
			t.Errorf("n=%d: %d comparisons, want %d", n, s.Comparisons, n*h)
		}
	}
}

// With n fixed, work grows linearly with h. Compare with Chan's algorithm,
// whose work only grows with log h on the same inputs.
func TestJarvisWorkIsLinearInHullSizeUnlikeChan(t *testing.T) {
	const n = 65536
	var jarvis, chans []float64
	hs := []int{8, 64, 512, 4096}
	for _, h := range hs {
		input := polygonWithInterior(n, h, 1)
		_, js := jarvisHull(input)
		_, cs := chanHull(input)
		jarvis = append(jarvis, float64(js.Comparisons))
		chans = append(chans, float64(cs.Ops()))
		t.Logf("n=%d h=%-5d jarvis=%-10d chan=%-9d", n, h, js.Comparisons, cs.Ops())
	}
	last := len(hs) - 1
	if g := jarvis[last] / jarvis[0]; g != 512 {
		t.Errorf("jarvis work grew %.1fx while h grew 512x; expected exactly 512x", g)
	}
	if g := chans[last] / chans[0]; g > 12 {
		t.Errorf("chan work grew %.1fx while h grew 512x; expected roughly log h growth", g)
	}
	// Small hull: Jarvis's n·h beats Chan's constant factors. Large hull: Chan wins by far.
	if jarvis[0] >= chans[0] {
		t.Errorf("h=8: jarvis (%v) should do less work than chan (%v)", jarvis[0], chans[0])
	}
	if jarvis[last] <= 10*chans[last] {
		t.Errorf("h=4096: jarvis (%v) should do far more work than chan (%v)", jarvis[last], chans[last])
	}
}

// Every point on the hull is the worst case: n·n comparisons.
func TestJarvisWorkIsQuadraticWhenAllPointsOnHull(t *testing.T) {
	for _, n := range []int{250, 500, 1000, 2000} {
		_, s := jarvisHull(onCircle(n, 1))
		if s.Comparisons != n*n {
			t.Errorf("n=%d: %d comparisons, want n² = %d", n, s.Comparisons, n*n)
		}
	}
}

// ---------------------------------------------------------------------------
// Efficiency: memory. Jarvis March reads the input in place and needs no
// working copy, so memory depends only on the hull size, never on n.
// ---------------------------------------------------------------------------

func TestJarvisMemoryDependsOnlyOnHullSize(t *testing.T) {
	measure := func(input []geometry.Point) (allocs, bytes int64) {
		res := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				JarvisMarch{}.Compute(input)
			}
		})
		return res.AllocsPerOp(), res.AllocedBytesPerOp()
	}
	smallA, smallB := measure(polygonWithInterior(1000, 16, 1))
	largeA, largeB := measure(polygonWithInterior(200000, 16, 1))
	t.Logf("h=16: n=1000 → %d allocs, %d bytes; n=200000 → %d allocs, %d bytes (input is %d bytes)",
		smallA, smallB, largeA, largeB, 200000*16)
	if largeA != smallA || largeB != smallB {
		t.Errorf("memory changed with n: %d allocs/%d bytes vs %d allocs/%d bytes", smallA, smallB, largeA, largeB)
	}
	if largeB > 1024 {
		t.Errorf("allocated %d bytes for a 16-vertex hull; expected only the output", largeB)
	}
}

// ---------------------------------------------------------------------------
// Efficiency: wall-clock time (skipped with -short; best of several runs and
// wide margins so it stays stable on busy machines)
// ---------------------------------------------------------------------------

func TestJarvisRuntimeScalesWithNAndH(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test skipped in -short mode")
	}
	base := polygonWithInterior(50000, 16, 1)
	moreN := polygonWithInterior(800000, 16, 1)
	moreH := polygonWithInterior(50000, 256, 1)
	tBase := bestTime(7, func() { JarvisMarch{}.Compute(base) })
	tN := bestTime(5, func() { JarvisMarch{}.Compute(moreN) })
	tH := bestTime(5, func() { JarvisMarch{}.Compute(moreH) })
	rN, rH := float64(tN)/float64(tBase), float64(tH)/float64(tBase)
	t.Logf("n=50k h=16: %v | 16x n: %v (%.1fx) | 16x h: %v (%.1fx)", tBase, tN, rN, tH, rH)
	// Both should be close to 16x.
	if rN > 40 {
		t.Errorf("16x more points took %.1fx longer; expected about 16x", rN)
	}
	if rH < 6 || rH > 40 {
		t.Errorf("16x more hull vertices took %.1fx longer; expected about 16x", rH)
	}
}

// Jarvis March is the fastest choice for tiny hulls and the slowest for big ones.
func TestJarvisCrossoverWithChanAndGraham(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test skipped in -short mode")
	}
	const n = 200000
	small := polygonWithInterior(n, 8, 3)
	large := polygonWithInterior(n, 2048, 3)
	jS := bestTime(5, func() { JarvisMarch{}.Compute(small) })
	cS := bestTime(5, func() { ChansAlgorithm{}.Compute(small) })
	gS := bestTime(3, func() { GrahamScan{}.Compute(small) })
	jL := bestTime(2, func() { JarvisMarch{}.Compute(large) })
	cL := bestTime(5, func() { ChansAlgorithm{}.Compute(large) })
	t.Logf("n=200k h=8:    jarvis %v, chan %v, graham %v", jS, cS, gS)
	t.Logf("n=200k h=2048: jarvis %v, chan %v", jL, cL)
	if jS >= cS || jS >= gS {
		t.Errorf("h=8: jarvis (%v) should beat chan (%v) and graham (%v)", jS, cS, gS)
	}
	if cL >= jL {
		t.Errorf("h=2048: chan (%v) should beat jarvis (%v)", cL, jL)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: go test ./internal/algorithms -run '^$' -bench Jarvis -benchmem
// ---------------------------------------------------------------------------

// Cases with n·h above 10^8 (all-on-hull at 100k points) are skipped: they take minutes.
func BenchmarkJarvisMarch(b *testing.B) { benchmarkHull(b, JarvisMarch{}, 100_000_000) }

func BenchmarkJarvisMarchByHullSize(b *testing.B) { benchmarkByHullSize(b, JarvisMarch{}) }
