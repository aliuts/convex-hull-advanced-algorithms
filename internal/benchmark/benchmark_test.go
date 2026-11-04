package benchmark

import (
	"time"

	"testing"

	"github.com/ali/convex-hull-advanced-algorithms/internal/algorithms"
	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

func TestRunReportsEveryAlgorithm(t *testing.T) {
	r := Run([]geometry.Point{{X: 0, Y: 0}}, algorithms.All())
	if len(r.Results) != len(algorithms.All()) {
		t.Fatalf("got %d results", len(r.Results))
	}
	for _, e := range r.Results {
		if e.InputSize != 1 || e.HullVertices != 1 || e.Result.Status != algorithms.StatusOK || len(e.Result.Hull) != 1 {
			t.Fatalf("unexpected %s result: %+v", e.Result.Algorithm, e)
		}
	}
}

// A single run on a tiny input is far below a coarse clock's resolution, so Run
// must repeat it and still report a non-zero average time.
func TestRunRepeatsShortAlgorithms(t *testing.T) {
	old := MinDuration
	MinDuration = 5 * time.Millisecond
	defer func() { MinDuration = old }()

	points := []geometry.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 0.2, Y: 0.2}}
	start := time.Now()
	r := Run(points, algorithms.All())
	if took := time.Since(start); took < time.Duration(len(r.Results))*MinDuration {
		t.Fatalf("Run took %v, want at least %v per algorithm", took, MinDuration)
	}
	for _, e := range r.Results {
		if e.Repeats < 2 || e.DurationNS <= 0 || e.HullVertices != 3 {
			t.Fatalf("%s: repeats=%d durationNs=%d hull=%d", e.Result.Algorithm, e.Repeats, e.DurationNS, e.HullVertices)
		}
		if avg := time.Duration(e.DurationNS) * time.Duration(e.Repeats); avg < MinDuration {
			t.Fatalf("%s: %d repeats × %dns is under MinDuration", e.Result.Algorithm, e.Repeats, e.DurationNS)
		}
	}
}
