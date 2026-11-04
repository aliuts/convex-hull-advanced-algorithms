// Command experiments runs the benchmark experiments from the report and prints
// the results as markdown tables, ready to paste.
//
//	go run ./cmd/experiments                     # both experiments, 3 trials each
//	go run ./cmd/experiments -trials 25          # more trials for steadier medians
//	go run ./cmd/experiments -csv results.csv    # also save every trial (for graphs)
//	go run ./cmd/experiments -controlled=false   # skip the slower controlled-h experiment
//
// Experiment 1 (Section 4.1) runs every algorithm on the four dataset kinds at
// several sizes. Experiment 2 (Section 4.3) fixes n and varies the hull size h.
//
// Timing uses benchmark.Run, the same code as the /api/benchmark endpoint: each
// algorithm gets its own copy of the same points, and only the hull computation
// is timed. Each table cell is the median over all trials.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ali/convex-hull-advanced-algorithms/internal/algorithms"
	"github.com/ali/convex-hull-advanced-algorithms/internal/benchmark"
	"github.com/ali/convex-hull-advanced-algorithms/internal/dataset"
	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

var labels = map[string]string{
	"graham_scan":  "Graham Scan",
	"jarvis_march": "Jarvis March",
	"chan":         "Chan's",
}

func main() {
	trials := flag.Int("trials", 3, "trials per configuration (the median is reported)")
	seed := flag.Int64("seed", 42, "random seed for every dataset")
	sizes := flag.String("sizes", "100,500,1000,5000,10000", "experiment 1: input sizes (max 10000)")
	kinds := flag.String("kinds", "uniform,circular,clustered,collinear", "experiment 1: dataset kinds")
	controlled := flag.Bool("controlled", true, "run experiment 2 (fixed n, varying h)")
	controlledN := flag.Int("controlled-n", 100000, "experiment 2: number of points")
	hullSizes := flag.String("hulls", "8,64,512,4096,n", "experiment 2: hull sizes (n = every point on the hull)")
	maxWork := flag.Float64("max-jarvis-work", 1e9, "experiment 2: skip Jarvis March when n·h exceeds this (it is O(nh))")
	csvPath := flag.String("csv", "", "also write every trial to this CSV file")
	flag.Parse()

	if *trials < 1 {
		log.Fatal("-trials must be at least 1")
	}

	var rows [][]string // raw trials for the CSV
	record := func(experiment, kind string, n, trial int, e benchmark.Entry) {
		rows = append(rows, []string{experiment, kind, strconv.Itoa(n), strconv.Itoa(trial),
			e.Result.Algorithm, strconv.Itoa(e.HullVertices), strconv.FormatInt(e.DurationNS, 10)})
	}

	fmt.Printf("%s, %s/%s, %d CPU threads, %s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH,
		runtime.NumCPU(), time.Now().Format("2006-01-02 15:04"))
	fmt.Printf("Median of %d trials per cell, seed %d. Fastest in bold.\n", *trials, *seed)

	// Experiment 1: dataset kinds × sizes.
	for _, kind := range splitList(*kinds) {
		fmt.Printf("\n### %s (times in µs)\n\n", kind)
		fmt.Println("| n | h | Graham Scan | Jarvis March | Chan's |")
		fmt.Println("|---|---|---|---|---|")
		for _, n := range parseInts(*sizes, 0) {
			points, err := dataset.Generate(kind, n, *seed)
			if err != nil {
				log.Fatalf("%s n=%d: %v", kind, n, err)
			}
			log.Printf("experiment 1: %s n=%d", kind, n)
			res := run(points, algorithms.All(), *trials, func(trial int, e benchmark.Entry) {
				record("distribution", kind, n, trial, e)
			})
			fmt.Println(res.row(formatNumber(float64(n), 0), 1e3, 1))
		}
	}

	// Experiment 2: fixed n, controlled hull size.
	if *controlled {
		n := *controlledN
		fmt.Printf("\n### Controlled hull size, n = %s (times in ms)\n\n", formatNumber(float64(n), 0))
		fmt.Println("| target h | h | Graham Scan | Jarvis March | Chan's |")
		fmt.Println("|---|---|---|---|---|")
		skipped := false
		for _, h := range parseInts(*hullSizes, n) {
			points, err := dataset.PolygonWithInterior(n, h, *seed)
			if err != nil {
				log.Fatalf("controlled n=%d h=%d: %v", n, h, err)
			}
			selected := algorithms.All()
			if float64(n)*float64(h) > *maxWork {
				skipped = true
				selected = slices.DeleteFunc(selected, func(a algorithms.ConvexHullAlgorithm) bool {
					return a.Name() == "jarvis_march"
				})
			}
			log.Printf("experiment 2: n=%d h=%d", n, h)
			res := run(points, selected, *trials, func(trial int, e benchmark.Entry) {
				record("controlled", "polygon", n, trial, e)
			})
			fmt.Println(res.row(formatNumber(float64(h), 0), 1e6, 1))
		}
		if skipped {
			fmt.Println("\nskipped: Jarvis March would take too long (raise -max-jarvis-work to run it)")
		}
	}

	if *csvPath != "" {
		if err := writeCSV(*csvPath, rows); err != nil {
			log.Fatal(err)
		}
		log.Printf("wrote %d trials to %s", len(rows), *csvPath)
	}
}

// results holds every trial's duration and hull size, per algorithm.
type results struct {
	durations map[string][]int64
	hulls     map[string]int
}

func run(points []geometry.Point, selected []algorithms.ConvexHullAlgorithm, trials int,
	each func(trial int, e benchmark.Entry)) results {
	res := results{durations: map[string][]int64{}, hulls: map[string]int{}}
	for trial := 1; trial <= trials; trial++ {
		for _, e := range benchmark.Run(points, selected).Results {
			name := e.Result.Algorithm
			res.durations[name] = append(res.durations[name], e.DurationNS)
			res.hulls[name] = e.HullVertices
			each(trial, e)
		}
	}
	return res
}

// row formats one table row: medians divided by unit (1e3 → µs, 1e6 → ms).
func (r results) row(first string, unit float64, decimals int) string {
	order := []string{"graham_scan", "jarvis_march", "chan"}
	fastest, best := "", 0.0
	for _, name := range order {
		if d, ok := r.durations[name]; ok {
			if m := median(d); fastest == "" || m < best {
				fastest, best = name, m
			}
		}
	}

	// h is the same for every algorithm except on near-degenerate input
	// (see the collinear dataset), where it is shown as a range.
	var hs []int
	for _, h := range r.hulls {
		hs = append(hs, h)
	}
	slices.Sort(hs)
	h := strconv.Itoa(hs[0])
	if hs[0] != hs[len(hs)-1] {
		h = fmt.Sprintf("%d–%d", hs[0], hs[len(hs)-1])
	}

	cells := []string{first, h}
	for _, name := range order {
		d, ok := r.durations[name]
		switch {
		case !ok:
			cells = append(cells, "skipped")
		case name == fastest:
			cells = append(cells, "**"+formatNumber(median(d)/unit, decimals)+"**")
		default:
			cells = append(cells, formatNumber(median(d)/unit, decimals))
		}
	}
	return "| " + strings.Join(cells, " | ") + " |"
}

func median(values []int64) float64 {
	s := slices.Clone(values)
	slices.Sort(s)
	if len(s)%2 == 1 {
		return float64(s[len(s)/2])
	}
	return float64(s[len(s)/2-1]+s[len(s)/2]) / 2
}

// formatNumber adds thousands separators and drops decimals for values ≥ 1000.
func formatNumber(v float64, decimals int) string {
	if v >= 1000 {
		decimals = 0
	}
	s := strconv.FormatFloat(v, 'f', decimals, 64)
	whole, frac, _ := strings.Cut(s, ".")
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + "," + whole[i:]
	}
	if frac != "" {
		return whole + "." + frac
	}
	return whole
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

// parseInts parses a comma-separated list; "n" stands for the value of n.
func parseInts(s string, n int) []int {
	var out []int
	for _, part := range splitList(s) {
		if part == "n" {
			out = append(out, n)
			continue
		}
		v, err := strconv.Atoi(part)
		if err != nil {
			log.Fatalf("invalid number %q", part)
		}
		out = append(out, v)
	}
	return out
}

func writeCSV(path string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	w.Write([]string{"experiment", "kind", "n", "trial", "algorithm", "hull_vertices", "duration_ns"})
	w.WriteAll(rows)
	if err := w.Error(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
