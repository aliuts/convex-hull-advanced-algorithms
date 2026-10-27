package benchmark

import (
	"testing"

	"github.com/ali/convex-hull-advanced-algorithms/internal/algorithms"
	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

func TestRunReportsNotImplementedWithoutInventingHull(t *testing.T) {
	r := Run([]geometry.Point{{X: 0, Y: 0}}, algorithms.All())
	if len(r.Results) != 3 {
		t.Fatalf("got %d results", len(r.Results))
	}
	for _, e := range r.Results {
		if e.Result.Status != algorithms.StatusNotImplemented || len(e.Result.Hull) != 0 || e.InputSize != 1 {
			t.Fatalf("unexpected result: %+v", e)
		}
	}
}
