package config

// SortResources, kaynakları bağımlılıklarına göre sıralar ve katmanlara ayırır.
// Şimdilik basit bir implementasyon: Tüm kaynakları tek bir katmanda döndürür.
// İleride 'depends_on' mantığı eklendiğinde burası topolojik sıralama (DAG) yapacaktır.
func SortResources(resources []ResourceConfig) ([][]ResourceConfig, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	// TODO: Topological Sort Implementation
	// Şu an için bağımlılık yönetimi yok varsayıyoruz ve
	// YAML dosyasındaki sırayı koruyarak tek bir 'batch' (katman) oluşturuyoruz.
	// Engine bu katmanları sırayla işler, katman içindekileri ise paralel işleyebilir.

	// Tüm kaynakları tek bir katmana koy
	firstLayer := make([]ResourceConfig, len(resources))
	copy(firstLayer, resources)

	return [][]ResourceConfig{firstLayer}, nil
}

// CheckCycles, döngüsel bağımlılık olup olmadığını kontrol eder.
// (Şimdilik placeholder)
func CheckCycles(resources []ResourceConfig) error {
	return nil
}

/*
İleride kullanılacak DAG yapısı örneği:

type Graph struct {
    Nodes map[string]ResourceConfig
    Edges map[string][]string
}

func buildGraph(resources []ResourceConfig) *Graph {
    ...
}
*/
