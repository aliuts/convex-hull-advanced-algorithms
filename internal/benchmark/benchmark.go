package benchmark

import (
	"time"

	"github.com/ali/convex-hull-advanced-algorithms/internal/algorithms"
	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

type Entry struct {
	Result       algorithms.Result `json:"result"`
	InputSize    int               `json:"inputSize"`
	HullVertices int               `json:"hullVertices"`
	DurationNS   int64             `json:"durationNs"`
	DurationMS   float64           `json:"durationMs"`
}
type Report struct {
	Results []Entry `json:"results"`
}

// Run times each algorithm against the same immutable input snapshot; correctness checks belong outside this timed path.
func Run(points []geometry.Point, selected []algorithms.ConvexHullAlgorithm) Report {
	input := append([]geometry.Point(nil), points...)
	report := Report{Results: make([]Entry, 0, len(selected))}
	for _, algorithm := range selected {
		start := time.Now()
		result := algorithm.Compute(append([]geometry.Point(nil), input...))
		elapsed := time.Since(start)
		entry := Entry{Result: result, InputSize: len(input), HullVertices: len(result.Hull), DurationNS: elapsed.Nanoseconds(), DurationMS: float64(elapsed.Nanoseconds()) / 1e6}
		report.Results = append(report.Results, entry)
	}
	return report
}

// RunTrials repeats identical inputs for each size and returns raw trial measurements.
func RunTrials(pointsBySize map[int][]geometry.Point, selected []algorithms.ConvexHullAlgorithm, trials int) map[int][]Report {
	if trials < 1 {
		trials = 1
	}
	out := make(map[int][]Report, len(pointsBySize))
	for size, points := range pointsBySize {
		for i := 0; i < trials; i++ {
			out[size] = append(out[size], Run(points, selected))
		}
	}
	return out
}
