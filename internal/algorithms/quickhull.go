package algorithms

import "github.com/ali/convex-hull-advanced-algorithms/internal/geometry"

type QuickHull struct{}

func (QuickHull) Name() string { return "quickhull" }
func (QuickHull) Compute(points []geometry.Point) Result {
	// TODO: Implement QuickHull. Keep the shared result contract unchanged.
	return Result{Algorithm: "quickhull", Status: StatusNotImplemented, Message: "QuickHull is not implemented yet."}
}
