package resources

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/config"
)

func New(r config.Resource, vars map[string]interface{}) (Resource, error) {
	// 1. Şablonlama işlemi
	fieldsToProcess := map[string]*string{
		"name":    &r.Name,
		"path":    &r.Path,
		"id":      &r.ID,
		"content": &r.Content,
	}

	for fieldName, fieldValue := range fieldsToProcess {
		if *fieldValue != "" {
			processed, err := config.ExecuteTemplate(*fieldValue, vars)
			if err != nil {
				return nil, fmt.Errorf("[%s] '%s' alanı şablon işleme hatası: %w", r.Name, fieldName, err)
			}
			*fieldValue = processed
		}
	}

	// 2. Canonical ID belirle
	canonicalID := r.Identify()

	// 3. Kaynak tipine göre nesneyi oluştur ve ID'yi enjekte et
	switch r.Type {
	case "file":
		return &FileResource{
			CanonicalID:  canonicalID,
			ResourceName: r.Name,
			Path:         r.Path,
			Content:      r.Content,
		}, nil

	case "package":
		return &PackageResource{
			CanonicalID: canonicalID,
			PackageName: r.Name,
			State:       r.State,
			Provider:    GetDefaultProvider(),
		}, nil

	case "service":
		return &ServiceResource{
			CanonicalID:  canonicalID,
			ServiceName:  r.Name,
			DesiredState: r.State,
			Enabled:      r.Enabled,
		}, nil

	case "git":
		return &GitResource{
			CanonicalID: canonicalID,
			URL:         r.URL,
			Path:        r.Path,
		}, nil

	case "exec":
		return &ExecResource{
			CanonicalID: canonicalID,
			Name:        r.Name,
			Command:     r.Command,
		}, nil

	case "noop":
		return nil, nil

	default:
		return nil, fmt.Errorf("bilinmeyen kaynak tipi: %s", r.Type)
	}
}
