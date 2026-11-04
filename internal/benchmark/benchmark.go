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
	DurationNS   int64             `json:"durationNs"` // average time of one run
	DurationMS   float64           `json:"durationMs"`
	Repeats      int               `json:"repeats"` // runs averaged into DurationNS
}
type Report struct {
	Results []Entry `json:"results"`
}

// MinDuration is how long each algorithm is run for before its time is taken.
// A single run is often shorter than the clock can measure (Windows' clock only
// ticks every 0.5–15.6 ms), so Run repeats the algorithm until this much time
// has passed and reports the average, like `go test -bench` does.
var MinDuration = 100 * time.Millisecond

// Run times each algorithm against the same immutable input snapshot; correctness checks belong outside this timed path.
func Run(points []geometry.Point, selected []algorithms.ConvexHullAlgorithm) Report {
	input := append([]geometry.Point(nil), points...)
	report := Report{Results: make([]Entry, 0, len(selected))}
	for _, algorithm := range selected {
		result, elapsed, repeats := timeRepeated(algorithm, input)
		ns := elapsed.Nanoseconds() / int64(repeats)
		report.Results = append(report.Results, Entry{
			Result: result, InputSize: len(input), HullVertices: len(result.Hull),
			DurationNS: ns, DurationMS: float64(ns) / 1e6, Repeats: repeats,
		})
	}
	return report
}

// timeRepeated runs the algorithm in batches, doubling the batch size until one
// batch takes at least MinDuration. Every run gets a fresh copy of the input, as
// algorithms may reorder it; the copy is a small part of the time, as before.
func timeRepeated(algorithm algorithms.ConvexHullAlgorithm, input []geometry.Point) (algorithms.Result, time.Duration, int) {
	buf := make([]geometry.Point, len(input))
	var result algorithms.Result
	for repeats := 1; ; repeats *= 2 {
		start := time.Now()
		for i := 0; i < repeats; i++ {
			copy(buf, input)
			result = algorithm.Compute(buf)
		}
		if elapsed := time.Since(start); elapsed >= MinDuration || repeats >= 1<<30 {
			return result, elapsed, repeats
		}
	}
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
