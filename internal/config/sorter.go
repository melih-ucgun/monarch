package config

import (
	"fmt"
)

// SortResources, kaynakları bağımlılıklarına (DependsOn) göre topolojik olarak sıralar.
// Bu sayede bir paket, ona bağımlı olan bir dosyadan veya servisten önce kurulur.
// Kahn'ın Algoritması kullanılmıştır.
func SortResources(resources []Resource) ([]Resource, error) {
	// 1. Veri yapılarını hazırla
	inDegree := make(map[string]int)    // Her düğümün giriş derecesi (kaç bağımlılığı var)
	adj := make(map[string][]string)    // Komşuluk listesi (kim kime bağımlı)
	resMap := make(map[string]Resource) // ID ile kaynağa hızlı erişim

	// 2. Grafı oluştur
	for _, res := range resources {
		// Kaynağın ID'sinin boş olmadığından emin ol (ID yoksa Name kullan)
		id := res.ID
		if id == "" {
			id = res.Name
		}

		resMap[id] = res

		// Eğer inDegree'de henüz yoksa 0 olarak başlat
		if _, exists := inDegree[id]; !exists {
			inDegree[id] = 0
		}

		// Bağımlılıkları işle
		for _, dep := range res.DependsOn {
			// dep -> id (id, dep'e bağlıdır)
			adj[dep] = append(adj[dep], id)
			inDegree[id]++
		}
	}

	// 3. Giriş derecesi 0 olanları (hiçbir şeye bağımlı olmayanları) kuyruğa ekle
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// 4. Kuyruktakileri sırayla işle
	var sorted []Resource
	for len(queue) > 0 {
		// Kuyruğun başından al
		u := queue[0]
		queue = queue[1:]

		// Eğer bu ID ana listede varsa (harici bağımlılık değilse) sonuca ekle
		if res, exists := resMap[u]; exists {
			sorted = append(sorted, res)
		}

		// Bu kaynağa bağlı olan diğer kaynakların giriş derecesini azalt
		for _, v := range adj[u] {
			inDegree[v]--
			// Eğer yeni giriş derecesi 0 olduysa kuyruğa ekle
			if inDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}

	// 5. Döngüsel bağımlılık kontrolü (Circular Dependency)
	// Eğer sıralanan kaynak sayısı, toplam kaynak sayısından azsa döngü vardır.
	if len(sorted) < len(resources) {
		return nil, fmt.Errorf("döngüsel bağımlılık tespit edildi veya eksik bir bağımlılık tanımlandı")
	}

	return sorted, nil
}
