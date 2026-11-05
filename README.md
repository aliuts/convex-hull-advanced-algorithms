# Convex Hull Algorithm Comparison

Implementations, tests and benchmarks for three convex hull algorithms in Go:

| Algorithm | File | Time |
|---|---|---|
| Graham Scan | `internal/algorithms/graham.go` | O(n log n) |
| Jarvis March (Gift Wrapping) | `internal/algorithms/jarvis.go` | O(nh) |
| Chan's algorithm | `internal/algorithms/chan.go` | O(n log h) |

n is the number of input points and h is the number of hull vertices. The project measures how each algorithm behaves as n, h and the point distribution change. A small web page is included for visualising hulls.

## Requirements

- Go 1.22 or newer
- Optional: [uv](https://docs.astral.sh/uv/) (or Python with matplotlib) to draw the report's figures

There are no third-party Go dependencies.

## Tests

```sh
go test ./...                                          # all tests, including wall-clock timing checks
go test -short ./...                                   # skip the timing checks (for noisy machines/CI)
go test ./internal/algorithms -v -run 'Chan|Jarvis'    # print the efficiency measurements
```

### Correctness

Every algorithm runs the shared correctness suite in `internal/algorithms/contract_test.go`:

- known edge cases (empty input, single points, duplicates, collinear points)
- random inputs checked against a reference algorithm
- every dataset kind
- agreement with Graham Scan

`exact_test.go` compares each algorithm with the exact hull of the same coordinates, computed with exact rational arithmetic (`math/big`). This catches sign errors from float64 cross products on nearly collinear points. `hull_helpers_test.go` holds the hull validator, the reference algorithm and the point generators.

### Efficiency

The efficiency tests count orientation tests and sort comparisons instead of relying only on time, so their results are exact and repeatable.

**Jarvis March** (`jarvis_test.go`) checks that:

- it does exactly n·h comparisons
- work is linear in n for a fixed h, and linear in h for a fixed n
- work is O(n²) when every point is on the hull
- memory depends only on h, not n
- the crossover point where Chan's algorithm overtakes it

**Chan's algorithm** (`chan_test.go`) checks that:

- work is linear in n when h is fixed, and grows with log h when n is fixed
- work is O(n log n) when every point is on the hull
- it uses about log log h rounds
- it makes a fixed number of allocations regardless of input size
- the binary-search tangent finder matches a linear scan
- it beats Graham Scan when the hull is small

## Benchmarks

### Go benchmarks

```sh
go test ./internal/algorithms -run '^$' -bench . -benchmem
```

Each algorithm has a benchmark over input size (`BenchmarkGrahamScan`, `BenchmarkJarvisMarch`, `BenchmarkChan`) and one over hull size (`...ByHullSize`). Use `-bench ByHullSize` to run only the hull size ones.

### Report experiments

```sh
go run ./cmd/experiments                     # both experiments, 3 trials, markdown tables (~5 s)
go run ./cmd/experiments -trials 25          # more trials for steadier medians
go run ./cmd/experiments -csv results.csv    # also save every trial, e.g. for graphs
go run ./cmd/experiments -controlled=false   # skip experiment 2
go run ./cmd/experiments -h                  # all options
```

- **Experiment 1: dataset kinds.** Runs every algorithm on uniform, circular, clustered and collinear points at 100 to 10,000 points.
- **Experiment 2: controlled hull size.** Fixes n = 100,000 and varies h from 8 up to n, using a regular polygon with random points inside it. Jarvis March is skipped where n·h would be too slow (`-max-jarvis-work`).

Each algorithm gets its own copy of the same points, and only the hull computation is timed. Within a trial, each algorithm is repeated until at least 100 ms has passed and the average is used, so times stay accurate where the system clock is coarse (for example on Windows). Each table cell is the median over all trials. Datasets are deterministic for a given `-seed` (default 42).

### Results and figures

The results used in the report are in `report/`:

- `results.md`: the markdown tables
- `results.csv`: every trial
- `run.log`: the run's log

To regenerate them and redraw the figures in `report/figures/`:

```sh
go run ./cmd/experiments -csv report/results.csv > report/results.md
uv run report/plot.py report/results.csv
```

## How the algorithms work

**Graham Scan** finds the lowest point, sorts the other points by angle around it, then walks them in order with a stack, removing any point that doesn't make a left turn. The angle sort uses cross products, not trigonometry.

**Jarvis March** gift-wraps from the lowest point, scanning every point at each step to find the next hull vertex. It does exactly n·h comparisons. That makes it the fastest option when the hull is small, and O(n²) when every point is on the hull. It needs no working copy of the input, so its memory depends only on the hull size.

**Chan's algorithm** is output-sensitive. For a guessed hull size m, it splits the points into groups of m and finds each group's hull with a monotone chain scan. It then gift-wraps around all the groups, using binary search to find each group's tangent point. If the hull doesn't close within m steps, m is squared (4, 16, 256, 65536, …) and it tries again.

## Visualisation (web app)

A Go server with a single browser page draws the hulls, which is useful for inspecting results on specific point sets:

```sh
go run ./cmd/server          # then open http://localhost:8080
ADDR=:9000 go run ./cmd/server
```

You can generate point sets (same kinds and seeds as the experiments), click to add points, right-click to remove them, and run one algorithm or all three to see the hulls, runtimes and hull sizes. The page uses a small JSON API (`/api/algorithms`, `/api/dataset`, `/api/run`, `/api/benchmark`) defined in `internal/api/server.go`.

## Project layout

```text
cmd/experiments/       Benchmark experiments used in the report
cmd/server/            Visualisation server entry point
internal/algorithms/   Algorithm interface, registry, implementations and their tests
internal/benchmark/    Timing and repeated multi-size trials
internal/dataset/      Deterministic dataset generators
internal/geometry/     Point and hull types
internal/api/          HTTP endpoints for the visualisation page
web/                   Visualisation page
report/                Experiment results, figures and plotting script
```

## Adding an algorithm

1. Implement `algorithms.ConvexHullAlgorithm` in its own file. `Compute` should return an `algorithms.Result` with `Status: algorithms.StatusOK` and the hull points in boundary order, without repeating the first point.
2. Register it in `algorithms.All()`. The experiments, benchmarks and web page pick it up from there (add a display name to `labels` in `cmd/experiments/main.go` and `LABELS` in `web/app.js`).
3. Run the shared correctness suite on it with `testHullContract` from `contract_test.go`.

Keep correctness checks in tests, not inside `Compute`, so they don't count towards the timed work.
