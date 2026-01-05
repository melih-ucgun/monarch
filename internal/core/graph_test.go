package core

import (
	"reflect"
	"testing"
)

func TestBuildGraph_Duplicates(t *testing.T) {
	items := []ConfigItem{
		{Name: "A", Type: "file"},
		{Name: "A", Type: "service"},
	}

	g := NewGraph()
	err := g.BuildGraph(items)
	if err == nil {
		t.Error("Expected error for duplicate resource name, got nil")
	}
}

func TestBuildGraph_UnknownDependency(t *testing.T) {
	items := []ConfigItem{
		{Name: "A", Type: "file", DependsOn: []string{"B"}},
	}

	g := NewGraph()
	err := g.BuildGraph(items)
	if err == nil {
		t.Error("Expected error for unknown dependency, got nil")
	}
}

func TestTopologicalSort_Cycle(t *testing.T) {
	// A -> B -> A
	items := []ConfigItem{
		{Name: "A", Type: "file", DependsOn: []string{"B"}},
		{Name: "B", Type: "file", DependsOn: []string{"A"}},
	}

	g := NewGraph()
	_ = g.BuildGraph(items)
	_, err := g.TopologicalSort()
	if err == nil {
		t.Error("Expected error for circular dependency, got nil")
	}
}

func TestTopologicalSort_SimpleChain(t *testing.T) {
	// A -> B -> C
	items := []ConfigItem{
		{Name: "A", Type: "file"},
		{Name: "C", Type: "file", DependsOn: []string{"B"}},
		{Name: "B", Type: "file", DependsOn: []string{"A"}},
	}

	g := NewGraph()
	if err := g.BuildGraph(items); err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	layers, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	// Should be 3 layers: [A], [B], [C]
	if len(layers) != 3 {
		t.Errorf("Expected 3 layers, got %d", len(layers))
	}
	if layers[0][0].Name != "A" {
		t.Errorf("Layer 0 should be A, got %s", layers[0][0].Name)
	}
	if layers[1][0].Name != "B" {
		t.Errorf("Layer 1 should be B, got %s", layers[1][0].Name)
	}
	if layers[2][0].Name != "C" {
		t.Errorf("Layer 2 should be C, got %s", layers[2][0].Name)
	}
}

func TestTopologicalSort_ParallelDiamond(t *testing.T) {
	// A --> B
	// |--> C
	// B, C --> D
	items := []ConfigItem{
		{Name: "D", Type: "file", DependsOn: []string{"B", "C"}},
		{Name: "B", Type: "file", DependsOn: []string{"A"}},
		{Name: "C", Type: "file", DependsOn: []string{"A"}},
		{Name: "A", Type: "file"},
	}

	g := NewGraph()
	if err := g.BuildGraph(items); err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	layers, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	// Expected layers:
	// 1. [A]
	// 2. [B, C] (order within layer is implementation detail, likely sorted by name)
	// 3. [D]

	if len(layers) != 3 {
		t.Fatalf("Expected 3 layers, got %d", len(layers))
	}

	if layers[0][0].Name != "A" {
		t.Errorf("Layer 0 should be A, got %s", layers[0][0].Name)
	}

	layer1Names := []string{layers[1][0].Name, layers[1][1].Name}
	// Sort to compare set
	expectedLayer1 := []string{"B", "C"}
	// Check content
	if !equalStringSlices(layer1Names, expectedLayer1) {
		t.Errorf("Layer 1 should contain B and C, got %v", layer1Names)
	}

	if layers[2][0].Name != "D" {
		t.Errorf("Layer 2 should be D, got %s", layers[2][0].Name)
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	// Simple map check since strict order isn't guaranteed by us unless we sorted queue
	// But our implementation sorts the queue, so B and C should be deterministic?
	// Our graph.go sorts queue. So B vs C depends on sort.Strings("B", "C") -> B then C.
	// So layers[1][0] should be B, layers[1][1] should be C.
	return reflect.DeepEqual(a, b)
}
