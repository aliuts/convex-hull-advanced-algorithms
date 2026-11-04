package algorithms

import "testing"

// Shared correctness checks (known cases, random inputs vs reference, dataset kinds, …).
func TestGrahamContract(t *testing.T) {
	testHullContract(t, GrahamScan{}, 5000)
}

func TestGrahamName(t *testing.T) {
	if (GrahamScan{}).Name() != "graham_scan" {
		t.Fatal("name must match the API identifier")
	}
}
