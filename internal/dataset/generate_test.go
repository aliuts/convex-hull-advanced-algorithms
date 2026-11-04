package dataset

import (
	"reflect"
	"testing"
)

func TestGenerateDeterministicAndKinds(t *testing.T) {
	for _, kind := range []string{"uniform", "clustered", "circular", "collinear"} {
		a, err := Generate(kind, 20, 42)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := Generate(kind, 20, 42)
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("%s should be deterministic", kind)
		}
	}
}
func TestGenerateRejectsInvalid(t *testing.T) {
	if _, err := Generate("unknown", 3, 1); err == nil {
		t.Fatal("expected kind error")
	}
	if _, err := Generate("uniform", MaxPoints+1, 1); err == nil {
		t.Fatal("expected count error")
	}
}

func TestPolygonWithInterior(t *testing.T) {
	a, err := PolygonWithInterior(1000, 16, 7)
	if err != nil || len(a) != 1000 {
		t.Fatalf("got %d points, err %v", len(a), err)
	}
	b, _ := PolygonWithInterior(1000, 16, 7)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("should be deterministic")
	}
	if all, _ := PolygonWithInterior(50, 50, 1); len(all) != 50 {
		t.Fatal("h = n should give n points")
	}
	for _, bad := range [][2]int{{10, 2}, {10, 11}} {
		if _, err := PolygonWithInterior(bad[0], bad[1], 1); err == nil {
			t.Fatalf("expected error for n=%d h=%d", bad[0], bad[1])
		}
	}
}
