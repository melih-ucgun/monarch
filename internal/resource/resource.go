package resource

import (
	"github.com/melih-ucgun/monarch/internal/core"
)

// Resource, sistemdeki herhangi bir yönetilebilir birimi (paket, dosya, servis vb.) temsil eder.
// Tüm adaptörler (apt, pacman, systemd, file) bu sözleşmeye uymak zorundadır.
type Resource interface {
	// Apply, kaynağı hedeflenen duruma getirir.
	// Örneğin: Paket kurulu değilse kurar, servis çalışmıyorsa başlatır.
	Apply(ctx *core.SystemContext) (core.Result, error)

	// Check, kaynağın mevcut durumunu kontrol eder.
	// Değişiklik gerekip gerekmediğini (pending changes) döner.
	// Dry-run modu ve durum raporu (status) için kritiktir.
	Check(ctx *core.SystemContext) (bool, error)

	// Validate, YAML'dan okunan parametrelerin doğruluğunu kontrol eder.
	// Örneğin: "name" alanı boş mu, geçersiz karakter var mı?
	Validate() error

	// GetName, kaynağın insan tarafından okunabilir adını döner.
	GetName() string
}

// BaseResource, tüm kaynakların paylaştığı ortak alanları içerir.
// Adaptörler bu struct'ı 'embed' ederek kod tekrarından kurtulur.
type BaseResource struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// GetName, BaseResource üzerindeki Name alanını döner.
// Bu sayede her adaptör için tekrar GetName yazmana gerek kalmaz.
func (b *BaseResource) GetName() string {
	return b.Name
}
