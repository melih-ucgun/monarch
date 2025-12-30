package resources

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/config"
	"github.com/mitchellh/mapstructure"
)

// FactoryFunc, yapılandırmayı alıp Resource üreten fonksiyon tipidir.
type FactoryFunc func(cfg config.ResourceConfig) (Resource, error)

// registry, kaynak tiplerini oluşturucu fonksiyonlarla eşleştirir.
var registry = make(map[string]FactoryFunc)

// Register, yeni bir kaynak tipini sisteme kaydeder.
func Register(typeStr string, fn FactoryFunc) {
	registry[typeStr] = fn
}

// New, verilen konfigürasyona göre uygun Resource nesnesini oluşturur.
func New(cfg config.ResourceConfig, vars map[string]string) (Resource, error) {
	factoryFn, exists := registry[cfg.Type]
	if !exists {
		return nil, fmt.Errorf("bilinmeyen kaynak tipi: %s", cfg.Type)
	}
	return factoryFn(cfg)
}

// init fonksiyonu, uygulama başladığında kaynakları kaydeder.
func init() {
	Register("file", newFileResource)
	Register("exec", newExecResource)
	Register("package", newPackageResource)
	Register("service", newServiceResource)
	Register("git", newGitResource)
	Register("symlink", newSymlinkResource)
	Register("container", newContainerResource)
}

// decodeConfig, parametre map'ini struct'a çevirir.
func decodeConfig(input interface{}, result interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           result,
		WeaklyTypedInput: true, // string -> int, string -> bool dönüşümleri için esneklik
	})
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

// --- Factory Fonksiyonları (Varsayılan Değerler Burada Atanır) ---

func newFileResource(cfg config.ResourceConfig) (Resource, error) {
	res := &FileResource{
		CanonicalID: cfg.ID,
		Mode:        "0644", // Varsayılan dosya izni
	}
	if err := decodeConfig(cfg.Parameters, res); err != nil {
		return nil, err
	}
	return res, nil
}

func newExecResource(cfg config.ResourceConfig) (Resource, error) {
	res := &ExecResource{
		CanonicalID: cfg.ID,
	}
	if err := decodeConfig(cfg.Parameters, res); err != nil {
		return nil, err
	}
	return res, nil
}

func newPackageResource(cfg config.ResourceConfig) (Resource, error) {
	res := &PackageResource{
		CanonicalID: cfg.ID,
		ManagerName: "pacman",    // Varsayılan paket yöneticisi
		State:       "installed", // Varsayılan durum
	}
	if err := decodeConfig(cfg.Parameters, res); err != nil {
		return nil, err
	}
	return res, nil
}

func newServiceResource(cfg config.ResourceConfig) (Resource, error) {
	res := &ServiceResource{
		CanonicalID: cfg.ID,
		State:       "started", // Varsayılan durum
		Enabled:     true,      // Varsayılan olarak başlangıçta çalıştır
	}
	if err := decodeConfig(cfg.Parameters, res); err != nil {
		return nil, err
	}
	return res, nil
}

func newGitResource(cfg config.ResourceConfig) (Resource, error) {
	res := &GitResource{
		CanonicalID: cfg.ID,
	}
	if err := decodeConfig(cfg.Parameters, res); err != nil {
		return nil, err
	}
	return res, nil
}

func newSymlinkResource(cfg config.ResourceConfig) (Resource, error) {
	res := &SymlinkResource{
		CanonicalID: cfg.ID,
		Force:       false,
	}
	if err := decodeConfig(cfg.Parameters, res); err != nil {
		return nil, err
	}
	return res, nil
}

func newContainerResource(cfg config.ResourceConfig) (Resource, error) {
	res := &ContainerResource{
		CanonicalID: cfg.ID,
		State:       "running",
	}
	// mapstructure, []interface{} -> []string dönüşümünü otomatik yapar
	if err := decodeConfig(cfg.Parameters, res); err != nil {
		return nil, err
	}
	return res, nil
}
