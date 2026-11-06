package dog

import (
	"math/rand"
	"testing"
)

func TestNamesLoadsEmbeddedCatalog(t *testing.T) {
	names, err := Names()
	if err != nil {
		t.Fatalf("Names() error = %v", err)
	}
	if len(names) < 100 {
		t.Fatalf("Names() returned %d names, want the full embedded catalog", len(names))
	}
	for index, name := range names {
		if name == "" {
			t.Fatalf("Names() entry %d is empty", index)
		}
	}
}

func TestRandomNameUsesSeededRandomIndex(t *testing.T) {
	first, err := RandomName(rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatalf("RandomName() error = %v", err)
	}
	second, err := RandomName(rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatalf("RandomName() second error = %v", err)
	}
	if first != second {
		t.Fatalf("RandomName() with the same seed returned %q and %q", first, second)
	}
}
