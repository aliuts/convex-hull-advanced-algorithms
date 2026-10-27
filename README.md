# HullLab — Convex Hull Algorithm Comparison

A Go-backed, browser-based workspace for exploring and benchmarking Graham Scan, Jarvis March (Gift Wrapping), and QuickHull. The server and UI are operational; the three algorithm implementations intentionally remain stubs and return an explicit `not_implemented` status.

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
- Run an individual algorithm or compare all three against a shared input snapshot.
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

Algorithm identifiers are `graham_scan`, `jarvis_march`, and `quickhull`; use `all` for the comparison. Run results contain `durationNs`, `durationMs`, `inputSize`, `hullVertices`, and an algorithm result. Stub responses have status `not_implemented` and no hull.

## Project layout

```text
cmd/server/            HTTP server entry point
internal/algorithms/   Shared interface, registry, and independent algorithm stubs
internal/geometry/     Point and hull types
internal/api/           HTTP routing, validation, and JSON endpoints
internal/benchmark/     Timing and repeated multi-size trial orchestration
internal/dataset/       Deterministic dataset generators
web/                    Responsive frontend assets served by Go
```

## Adding an algorithm

Implement `algorithms.ConvexHullAlgorithm` in its own file. `Compute` should return the shared `algorithms.Result` with `Status: algorithms.StatusOK` and the hull points in boundary order (without repeating the first point). Register the type in `algorithms.All()`. API selection, comparison, rendering, and benchmark reporting use the shared interface and result contract. Keep input correctness validation in separate tests/validation paths rather than including it in timed computation.

## GitHub setup

The repository is initialized with a Go module and `.gitignore`. To connect it to a newly created GitHub repository:

```sh
```
