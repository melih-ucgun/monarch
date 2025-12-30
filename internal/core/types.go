package core

// ResourceState, bir kaynağın olması gereken durumunu belirtir.
type ResourceState string

const (
	StatePresent ResourceState = "present" // Kaynak var olmalı
	StateAbsent  ResourceState = "absent"  // Kaynak silinmeli
	StateLatest  ResourceState = "latest"  // Kaynak (paketse) en güncel sürümde olmalı
)

// String, ResourceState'i string'e çevirir.
func (s ResourceState) String() string {
	return string(s)
}
