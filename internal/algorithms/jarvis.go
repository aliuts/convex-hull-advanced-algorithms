package algorithms

import "github.com/ali/convex-hull-advanced-algorithms/internal/geometry"

type JarvisMarch struct{}

func (JarvisMarch) Name() string { return "jarvis_march" }
func (JarvisMarch) Compute(points []geometry.Point) Result {
	// TODO: Implement Jarvis March (Gift Wrapping). Keep the shared result contract unchanged.
	return Result{Algorithm: "jarvis_march", Status: StatusNotImplemented, Message: "Jarvis March is not implemented yet."}
}
