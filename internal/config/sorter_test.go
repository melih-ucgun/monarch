package config

import (
	"reflect"
	"testing"
)

func TestSortResources(t *testing.T) {
	// Basit bir senaryo: Sıralama değişmeden dönmeli (şimdilik)
	input := [][]ResourceConfig{
		{
			{ID: "res1", Type: "file"},
		},
		{
			{ID: "res2", Type: "service"},
		},
	}

	expected := input

	result, err := SortResources(input)
	if err != nil {
		t.Fatalf("SortResources hata döndü: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Beklenen: %v, Alınan: %v", expected, result)
	}
}

func TestFlatten(t *testing.T) {
	input := [][]ResourceConfig{
		{{ID: "1"}, {ID: "2"}},
		{{ID: "3"}},
	}
	expectedLen := 3

	flat := Flatten(input)
	if len(flat) != expectedLen {
		t.Errorf("Beklenen uzunluk %d, alınan %d", expectedLen, len(flat))
	}
	if flat[2].ID != "3" {
		t.Errorf("Flatten sıralaması hatalı")
	}
}
