package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ali/convex-hull-advanced-algorithms/internal/algorithms"
	"github.com/ali/convex-hull-advanced-algorithms/internal/benchmark"
	"github.com/ali/convex-hull-advanced-algorithms/internal/dataset"
	"github.com/ali/convex-hull-advanced-algorithms/internal/geometry"
)

type Server struct {
	algorithms []algorithms.ConvexHullAlgorithm
	byName     map[string]algorithms.ConvexHullAlgorithm
	mux        *http.ServeMux
}

func NewServer(available []algorithms.ConvexHullAlgorithm) *Server {
	s := &Server{algorithms: available, byName: make(map[string]algorithms.ConvexHullAlgorithm), mux: http.NewServeMux()}
	for _, a := range available {
		s.byName[a.Name()] = a
	}
	s.mux.HandleFunc("/api/algorithms", s.handleAlgorithms)
	s.mux.HandleFunc("/api/run", s.handleRun)
	s.mux.HandleFunc("/api/benchmark", s.handleBenchmark)
	s.mux.HandleFunc("/api/dataset", s.handleDataset)
	s.mux.Handle("/", http.FileServer(http.Dir("web")))
	return s
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func (s *Server) handleAlgorithms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	names := make([]string, 0, len(s.algorithms))
	for _, a := range s.algorithms {
		names = append(names, a.Name())
	}
	jsonResponse(w, http.StatusOK, map[string]any{"algorithms": names})
}

type runRequest struct {
	Points    []geometry.Point `json:"points"`
	Algorithm string           `json:"algorithm"`
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
		return false
	}
	return true
}
func validatePoints(points []geometry.Point) error {
	if len(points) > dataset.MaxPoints {
		return errors.New("point limit exceeded")
	}
	for _, p := range points {
		if !p.Validate() {
			return errors.New("coordinates must be finite numbers")
		}
	}
	return nil
}
func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req runRequest
	if !decode(w, r, &req) {
		return
	}
	if err := validatePoints(req.Points); err != nil {
		jsonResponse(w, 400, map[string]string{"error": err.Error()})
		return
	}
	selected, err := s.selectAlgorithms(req.Algorithm)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, benchmark.Run(req.Points, selected))
}

type datasetRequest struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
	Seed  int64  `json:"seed"`
}

func (s *Server) handleDataset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req datasetRequest
	if !decode(w, r, &req) {
		return
	}
	points, err := dataset.Generate(req.Kind, req.Count, req.Seed)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "kind must be uniform, clustered, circular, or collinear; count must be between 0 and 10000"})
		return
	}
	jsonResponse(w, 200, map[string]any{"points": points, "count": len(points), "kind": req.Kind, "seed": req.Seed})
}

type benchRequest struct {
	Sizes     []int  `json:"sizes"`
	Seed      int64  `json:"seed"`
	Kind      string `json:"kind"`
	Trials    int    `json:"trials"`
	Algorithm string `json:"algorithm"`
}

func (s *Server) handleBenchmark(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req benchRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Trials < 1 || req.Trials > 100 || len(req.Sizes) == 0 || len(req.Sizes) > 20 {
		jsonResponse(w, 400, map[string]string{"error": "provide 1-20 sizes and 1-100 trials"})
		return
	}
	selected, err := s.selectAlgorithms(req.Algorithm)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": err.Error()})
		return
	}
	data := make(map[int][]geometry.Point, len(req.Sizes))
	for _, size := range req.Sizes {
		pts, e := dataset.Generate(req.Kind, size, req.Seed)
		if e != nil {
			jsonResponse(w, 400, map[string]string{"error": "invalid dataset size or kind"})
			return
		}
		data[size] = pts
	}
	jsonResponse(w, 200, benchmark.RunTrials(data, selected, req.Trials))
}
func (s *Server) selectAlgorithms(name string) ([]algorithms.ConvexHullAlgorithm, error) {
	if name == "all" {
		return s.algorithms, nil
	}
	a, ok := s.byName[strings.TrimSpace(name)]
	if !ok {
		return nil, errors.New("unknown algorithm")
	}
	return []algorithms.ConvexHullAlgorithm{a}, nil
}
