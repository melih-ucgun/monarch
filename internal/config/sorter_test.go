package config

import (
	"testing"
)

func TestSortResources(t *testing.T) {
	// Girdi artık düz bir liste: []ResourceConfig
	input := []ResourceConfig{
		{ID: "res1", Type: "file"},
		{ID: "res2", Type: "package"},
	}

	levels, err := SortResources(input)
	if err != nil {
		t.Fatalf("SortResources hata döndü: %v", err)
	}

	// Mevcut basit implementasyon her şeyi tek katmana koyduğu için:
	if len(levels) != 1 {
		t.Errorf("Beklenen katman sayısı 1, gelen %d", len(levels))
	}

	if len(levels[0]) != 2 {
		t.Errorf("Beklenen kaynak sayısı 2, gelen %d", len(levels[0]))
	}
}

func TestSortEmptyResources(t *testing.T) {
	levels, err := SortResources(nil)
	if err != nil {
		t.Fatalf("Boş listede hata döndü: %v", err)
	}
	if levels != nil {
		t.Error("Boş girdi için nil dönmeliydi")
	}
}
