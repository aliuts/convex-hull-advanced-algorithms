package algorithms

import "github.com/ali/convex-hull-advanced-algorithms/internal/geometry"

// Result is the common output contract for all hull algorithms.
type Result struct {
	Algorithm string           `json:"algorithm"`
	Status    string           `json:"status"`
	Message   string           `json:"message,omitempty"`
	Hull      []geometry.Point `json:"hull,omitempty"`
}

// ConvexHullAlgorithm is the integration point for independently implemented algorithms.
type ConvexHullAlgorithm interface {
	Name() string
	Compute(points []geometry.Point) Result
}

const (
	StatusOK             = "implemented"
	StatusNotImplemented = "not_implemented"
)
