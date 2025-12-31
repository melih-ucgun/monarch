package config

import (
	"fmt"
	"strings"
)

// SortResources sorts resources based on their dependencies and separates them into layers.
// Uses Kahn's Algorithm.
func SortResources(resources []ResourceConfig) ([][]ResourceConfig, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	// 1. Create resource map and dependency graph
	resourceMap := make(map[string]ResourceConfig)
	graph := make(map[string][]string) // key -> dependants (dependants of key)
	inDegree := make(map[string]int)   // key -> dependency count (how many things key depends on)

	// ID collision check and preparation
	for _, res := range resources {
		if _, exists := resourceMap[res.ID]; exists {
			return nil, fmt.Errorf("duplicate resource ID found: %s", res.ID)
		}
		resourceMap[res.ID] = res
		inDegree[res.ID] = 0 // Initial value
	}

	// 2. Populate graph
	for _, res := range resources {
		for _, depID := range res.DependsOn {
			// Does dependent resource exist?
			if _, exists := resourceMap[depID]; !exists {
				return nil, fmt.Errorf("resource '%s' depends on unknown resource '%s'", res.ID, depID)
			}

			// Graph: depID -> res.ID (res.ID can open when depID is finished)
			graph[depID] = append(graph[depID], res.ID)
			inDegree[res.ID]++
		}
	}

	// 3. Initial set (Independent nodes - Layer 0)
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// Is lexicographical sortQueue needed? Not for now but
	// queue can be sorted to be deterministic.

	var layers [][]ResourceConfig
	processedCount := 0

	for len(queue) > 0 {
		var nextLayer []string
		var currentLayerConfigs []ResourceConfig

		// Current queue represents a layer (parallelizable)
		// However, standard Kahn algorithm proceeds node by node.
		// To create parallel layers, we must "freeze" and process the current queue.

		layerSize := len(queue)
		for i := 0; i < layerSize; i++ {
			node := queue[i]
			processedCount++
			currentLayerConfigs = append(currentLayerConfigs, resourceMap[node])

			// Decrease degree of dependants
			for _, neighbour := range graph[node] {
				inDegree[neighbour]--
				if inDegree[neighbour] == 0 {
					nextLayer = append(nextLayer, neighbour)
				}
			}
		}

		layers = append(layers, currentLayerConfigs)
		queue = nextLayer // Proceed to next layer
	}

	// 4. Cycle check
	if processedCount != len(resources) {
		// There is a cycle. Which nodes were not processed?
		var unprocessed []string
		for id, degree := range inDegree {
			if degree > 0 {
				unprocessed = append(unprocessed, id)
			}
		}
		return nil, fmt.Errorf("dependency cycle detected involves: %v", strings.Join(unprocessed, ", "))
	}

	return layers, nil
}
