package algorithms

import (
	"fmt"
	"sort"

	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

type GrahamScan struct{}

func (GrahamScan) Name() string { return "graham_scan" }
func (GrahamScan) Compute(points []geometry.Point) Result {
	result := Result{Algorithm: "graham_scan", Status: StatusOK}
	if len(points) == 0 {
		result.Hull = []geometry.Point{}
		return result
	}

	pivot, _ := FindPivot(points)
	unique := make([]geometry.Point, 0, len(points))
	seen := make(map[geometry.Point]struct{}, len(points))
	for _, point := range points {
		if _, exists := seen[point]; exists {
			continue
		}
		seen[point] = struct{}{}
		if point != pivot {
			unique = append(unique, point)
		}
	}
	if len(unique) == 0 {
		result.Hull = []geometry.Point{pivot}
		return result
	}

	// Sort by angle around the pivot without trigonometry. The pivot is the
	// lowest point, so every other point lies in the half-plane above it, and
	// a comes before b exactly when pivot -> a -> b is a left turn. Points at
	// the same angle are ordered closest first.
	sort.Slice(unique, func(i, j int) bool {
		left, right := unique[i], unique[j]
		if turn := CrossProduct(pivot, left, right); turn != 0 {
			return turn > 0
		}
		return squaredDistance(pivot, left) < squaredDistance(pivot, right)
	})

	hull := make([]geometry.Point, 0, len(unique)+1)
	hull = append(hull, pivot)
	for _, point := range unique {
		for len(hull) >= 2 && CrossProduct(hull[len(hull)-2], hull[len(hull)-1], point) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, point)
	}
	result.Hull = hull
	return result
}

func CrossProduct(o, a, b geometry.Point) float64 {
	return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X)
}

func FindPivot(points []geometry.Point) (geometry.Point, error) {
	if len(points) == 0 {
		return geometry.Point{}, fmt.Errorf("No points provided.")
	}
	pivot := points[0]

	for _, p := range points[1:] {
		if p.Y < pivot.Y || (p.Y == pivot.Y && p.X < pivot.X) {
			pivot = p
		}
	}
	return pivot, nil
}

func squaredDistance(a, b geometry.Point) float64 {
	dx, dy := b.X-a.X, b.Y-a.Y
	return dx*dx + dy*dy
}
