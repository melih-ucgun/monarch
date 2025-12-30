package state

import "time"

// ResourceEntry, tek bir kaynağın sistemdeki durumunu temsil eder.
type ResourceEntry struct {
	ID          string                 `json:"id"` // Unique ID (Name + Type)
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	State       string                 `json:"state"`        // Desired state (present, absent)
	LastApplied time.Time              `json:"last_applied"` // Son uygulanma zamanı
	Status      string                 `json:"status"`       // success, failed
	Metadata    map[string]interface{} `json:"metadata"`     // Ekstra bilgiler (version, hash vb.)
}

// State, tüm sistemin o anki snapshot'ıdır.
type State struct {
	Version   string                   `json:"version"` // State dosya versiyonu
	LastRun   time.Time                `json:"last_run"`
	Resources map[string]ResourceEntry `json:"resources"`
}

func NewState() *State {
	return &State{
		Version:   "1.0",
		Resources: make(map[string]ResourceEntry),
	}
}
