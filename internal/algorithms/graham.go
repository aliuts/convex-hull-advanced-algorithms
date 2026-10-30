package algorithms

import (
	"fmt"
	"math"
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

	sort.Slice(unique, func(i, j int) bool {
		left, right := unique[i], unique[j]
		leftAngle := math.Atan2(left.Y-pivot.Y, left.X-pivot.X)
		rightAngle := math.Atan2(right.Y-pivot.Y, right.X-pivot.X)
		if leftAngle != rightAngle {
			return leftAngle < rightAngle
		}
		leftDistance := math.Hypot(left.X-pivot.X, left.Y-pivot.Y)
		rightDistance := math.Hypot(right.X-pivot.X, right.Y-pivot.Y)
		return leftDistance < rightDistance
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
