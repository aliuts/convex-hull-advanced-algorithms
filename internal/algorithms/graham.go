package algorithms

import "github.com/ali/convex-hull-advanced-algorithms/internal/geometry"

type GrahamScan struct{}

func (GrahamScan) Name() string { return "graham_scan" }
func (GrahamScan) Compute(points []geometry.Point) Result {
	// TODO: Implement Graham Scan. Keep the shared result contract unchanged.
	return Result{Algorithm: "graham_scan", Status: StatusNotImplemented, Message: "Graham Scan is not implemented yet."}
}

func crossProduct(o, a, b geometry.Point) float64 {
	return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X)
}
