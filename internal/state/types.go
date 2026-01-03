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

// TransactionChange represents a single change within a transaction.
type TransactionChange struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Action     string `json:"action"`
	Target     string `json:"target,omitempty"`
	BackupPath string `json:"backup_path,omitempty"`
}

// Transaction represents a session of changes (e.g. one apply run).
type Transaction struct {
	ID        string              `json:"id"`
	Timestamp time.Time           `json:"timestamp"`
	Status    string              `json:"status"` // success, failed, reverted
	Changes   []TransactionChange `json:"changes"`
}

// State, tüm sistemin o anki snapshot'ıdır.
type State struct {
	Version   string                   `json:"version"` // State dosya versiyonu
	LastRun   time.Time                `json:"last_run"`
	Resources map[string]ResourceEntry `json:"resources"`
	History   []Transaction            `json:"history,omitempty"` // Log of actions
}

func NewState() *State {
	return &State{
		Version:   "1.0",
		Resources: make(map[string]ResourceEntry),
	}
}
