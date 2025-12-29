package config

// SortResources, kaynakları bağımlılıklarına göre sıralar (Topological Sort).
// Not: Şu anki versiyonda sadece ResourceConfig listesini "level" (katman) mantığına göre
// ayırabiliriz veya basitçe sıralı döndürebiliriz.
// İleride "depends_on" alanı eklenirse burası gerçek bir DAG (Graph) algoritması içermeli.
func SortResources(resources [][]ResourceConfig) ([][]ResourceConfig, error) {
	// Şimdilik konfigürasyon dosyasındaki sıraya sadık kalıyoruz.
	// resources zaten [][]ResourceConfig tipinde olduğu için ekstra bir dönüşüm gerekmez.

	// Eğer düz bir liste gelirse ve biz bunu katmanlara ayırmak istiyorsak
	// burada işlem yapılabilir. Ancak şu anki config yapısı (YAML) zaten katmanlı (array of arrays)
	// veya gruplu geldiği için olduğu gibi dönebiliriz.

	// Örnek basit validasyon:
	if len(resources) == 0 {
		return nil, nil // Boş ise hata değil, boş liste dön
	}

	return resources, nil
}

// Bu yardımcı fonksiyon, ileride bağımlılık analizi yapıldığında kullanılabilir.
// Şu an ResourceConfig içinde "DependsOn" alanı olmadığı için pasif.
func validateDependencies(res ResourceConfig, allResources map[string]ResourceConfig) error {
	// Örn: if res.DependsOn != "" && allResources[res.DependsOn] == nil { error }
	return nil
}

// Flatten, katmanlı yapıyı düz listeye çevirir (İhtiyaç olursa)
func Flatten(layers [][]ResourceConfig) []ResourceConfig {
	var flat []ResourceConfig
	for _, layer := range layers {
		flat = append(flat, layer...)
	}
	return flat
}
