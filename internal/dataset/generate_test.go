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
