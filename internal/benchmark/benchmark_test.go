package benchmark

import (
	"testing"

	"github.com/ali/convex-hull-advanced-algorithms/internal/algorithms"
	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

func TestRunReportsImplementedAndNotImplementedAlgorithms(t *testing.T) {
	r := Run([]geometry.Point{{X: 0, Y: 0}}, algorithms.All())
	if len(r.Results) != len(algorithms.All()) {
		t.Fatalf("got %d results", len(r.Results))
	}
	for _, e := range r.Results {
		if e.InputSize != 1 {
			t.Fatalf("unexpected result: %+v", e)
		}
		switch e.Result.Algorithm {
		case "graham_scan", "chan":
			if e.Result.Status != algorithms.StatusOK || len(e.Result.Hull) != 1 {
				t.Fatalf("unexpected %s result: %+v", e.Result.Algorithm, e.Result)
			}
		default:
			if e.Result.Status != algorithms.StatusNotImplemented || len(e.Result.Hull) != 0 {
				t.Fatalf("unexpected placeholder result: %+v", e.Result)
			}
		}
	}
}
