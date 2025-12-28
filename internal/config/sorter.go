package config

import (
	"fmt"
)

// SortResources, kaynakları bağımlılıklarına göre paralel çalışabilecek seviyelere (layer) ayırır.
func SortResources(resources []Resource) ([][]Resource, error) {
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	resMap := make(map[string]Resource)

	// Kaynak haritasını ve giriş derecelerini (in-degree) hazırla
	for _, r := range resources {
		resMap[r.Name] = r
		if _, ok := inDegree[r.Name]; !ok {
			inDegree[r.Name] = 0
		}
		for _, dep := range r.DependsOn {
			adj[dep] = append(adj[dep], r.Name)
			inDegree[r.Name]++
		}
	}

	// Bağımlılıkların mevcut olup olmadığını kontrol et
	for _, r := range resources {
		for _, dep := range r.DependsOn {
			if _, ok := resMap[dep]; !ok {
				return nil, fmt.Errorf("kaynak bulunamadı: %s (bağımlılık hatası: %s)", dep, r.Name)
			}
		}
	}

	var levels [][]Resource
	for len(inDegree) > 0 {
		var currentLevel []Resource
		var currentLevelNames []string

		// Bağımlılığı kalmayan (derecesi 0 olan) kaynakları bu seviyeye ekle
		for name, degree := range inDegree {
			if degree == 0 {
				currentLevel = append(currentLevel, resMap[name])
				currentLevelNames = append(currentLevelNames, name)
			}
		}

		// Eğer derecesi 0 olan kaynak kalmadıysa ama hala bekleyen kaynak varsa döngü vardır
		if len(currentLevel) == 0 {
			return nil, fmt.Errorf("döngüsel bağımlılık tespit edildi")
		}

		// İşlenenleri grafikten çıkar ve onlara bağlı olanların derecesini azalt
		for _, name := range currentLevelNames {
			delete(inDegree, name)
			for _, neighbor := range adj[name] {
				inDegree[neighbor]--
			}
		}
		levels = append(levels, currentLevel)
	}

	return levels, nil
}
