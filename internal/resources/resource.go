package resources

import "context"

// Resource, Monarch tarafından yönetilen her varlığın uyması gereken arayüzdür.
type Resource interface {
	ID() string
	Check() (bool, error)           // İstenen durumda mı?
	Apply() error                   // Durumu düzelt
	Undo(ctx context.Context) error // YAPILAN İŞLEMİ GERİ AL (Yeni)
	Diff() (string, error)          // Mevcut ve istenen durum arasındaki farkı metin olarak döner
}
