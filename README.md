# HullLab — Convex Hull Algorithm Comparison

A Go-backed, browser-based workspace for exploring and benchmarking Graham Scan, Jarvis March (Gift Wrapping), QuickHull, and Chan's algorithm. Graham Scan and Chan's algorithm are implemented; Jarvis March and QuickHull are still stubs that return an explicit `not_implemented` status.

## Requirements

- Go 1.22 or newer
- A modern browser

No third-party Go dependencies are required.

## Run

```sh
```

Open [http://localhost:8080](http://localhost:8080). Set `ADDR` to change the listen address, for example `ADDR=:9000 go run ./cmd/server`.

```sh
```

## Features

- Click the coordinate plane to add points; drag to move and right-click to delete.
- Generate deterministic uniform, clustered, circular, or collinear point sets by count and seed.
- Run an individual algorithm or compare them all against a shared input snapshot.
- View runtime, input size, hull vertex count, and explicit implementation status.
- Run multi-size benchmark trials (100, 500, 1,000, and 5,000 points; three trials per size by default).
- Implemented hull results automatically draw as a polygon in the plane.

## HTTP API

All request and response bodies use JSON.

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/algorithms` | GET | Stable algorithm identifiers |
| `/api/dataset` | POST | Generate `{ "kind": "uniform", "count": 100, "seed": 42 }` |
| `/api/run` | POST | Run `{ "algorithm": "all", "points": [{"x":1,"y":2}] }` |
| `/api/benchmark` | POST | Run `{ "sizes": [100,500], "kind": "uniform", "seed": 42, "trials": 3, "algorithm": "all" }` |

Algorithm identifiers are `graham_scan`, `jarvis_march`, `quickhull`, and `chan`; use `all` for the comparison. Run results contain `durationNs`, `durationMs`, `inputSize`, `hullVertices`, and an algorithm result. Stub responses have status `not_implemented` and no hull.

## Project layout

```text
cmd/server/            HTTP server entry point
internal/algorithms/   Shared interface, registry, and algorithm implementations
internal/geometry/     Point and hull types
internal/api/           HTTP routing, validation, and JSON endpoints
internal/benchmark/     Timing and repeated multi-size trial orchestration
internal/dataset/       Deterministic dataset generators
web/                    Responsive frontend assets served by Go
```

## Adding an algorithm

Implement `algorithms.ConvexHullAlgorithm` in its own file. `Compute` should return the shared `algorithms.Result` with `Status: algorithms.StatusOK` and the hull points in boundary order (without repeating the first point). Register the type in `algorithms.All()`. API selection, comparison, rendering, and benchmark reporting use the shared interface and result contract. Keep input correctness validation in separate tests/validation paths rather than including it in timed computation.

## Chan's algorithm

`internal/algorithms/chan.go` is output-sensitive: O(n log h), where h is the number of hull vertices. For a guessed hull size m it splits the points into groups of m, finds each group's hull with a monotone chain scan, then gift-wraps around all groups, finding each group's tangent point by binary search. If the hull doesn't close within m steps, m is squared (4, 16, 256, 65536, …) and it tries again.

## Tests and benchmarks

```sh
go test ./...                    # all tests, including wall-clock timing checks
go test -short ./...             # skip the timing checks (for noisy machines/CI)
go test ./internal/algorithms -v -run Chan   # print the efficiency measurements
go test ./internal/algorithms -run '^$' -bench . -benchmem   # Chan vs Graham scan benchmarks
```

The efficiency tests count orientation tests and sort comparisons rather than relying only on time, so they are exact and repeatable. For Chan's algorithm they check that work is linear in n when h is fixed, grows with log h when n is fixed, is O(n log n) when every point is on the hull, uses about log log h rounds, and makes a fixed handful of allocations regardless of input size. `hull_helpers_test.go` holds a hull validator, a reference algorithm, and point generators that other implementations can reuse.

## GitHub setup

The repository is initialized with a Go module and `.gitignore`. To connect it to a newly created GitHub repository:

```sh
```
