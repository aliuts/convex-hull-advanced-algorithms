package algorithms

import "github.com/ali/convex-hull-advanced-algorithms/internal/geometry"

type GrahamScan struct{}

func (GrahamScan) Name() string { return "graham_scan" }
func (GrahamScan) Compute(points []geometry.Point) Result {
	// TODO: Implement Graham Scan. Keep the shared result contract unchanged.
	return Result{Algorithm: "graham_scan", Status: StatusNotImplemented, Message: "Graham Scan is not implemented yet."}
}
