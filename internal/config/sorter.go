package config

import (
	"fmt"
)

func SortResources(resources []Resource) ([]Resource, error) {
	inDegree := make(map[string]int)
	adj := make(map[string][]string)
	resMap := make(map[string]Resource)

	// 1. Grafı Identify() kullanarak oluştur
	for _, res := range resources {
		id := res.Identify() // Artık 'id' veya 'type:name' döner
		resMap[id] = res

		if _, exists := inDegree[id]; !exists {
			inDegree[id] = 0
		}

		for _, dep := range res.DependsOn {
			adj[dep] = append(adj[dep], id)
			inDegree[id]++
		}
	}

	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var sorted []Resource
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]

		if res, exists := resMap[u]; exists {
			sorted = append(sorted, res)
		}

		for _, v := range adj[u] {
			inDegree[v]--
			if inDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}

	if len(sorted) < len(resources) {
		return nil, fmt.Errorf("döngüsel bağımlılık tespit edildi veya eksik bir bağımlılık tanımlandı")
	}

	return sorted, nil
}
