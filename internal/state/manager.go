package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager, state dosyasını okuma ve yazma işlemlerini yönetir.
// Thread-safe olması için Mutex kullanır.
type Manager struct {
	FilePath string
	Current  *State
	mu       sync.RWMutex
}

// NewManager yeni bir state yöneticisi oluşturur ve mevcut dosyayı yükler.
func NewManager(path string) (*Manager, error) {
	mgr := &Manager{
		FilePath: path,
		Current:  NewState(),
	}

	// Dosya varsa yükle, yoksa boş başla
	if err := mgr.Load(); err != nil {
		// Eğer dosya yoksa sorun değil, yeni oluşturacağız
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return mgr, nil
}

// Load, state dosyasını diskten okur.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.FilePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, m.Current)
}

// Save, mevcut durumu diske yazar.
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.Current.LastRun = time.Now()

	data, err := json.MarshalIndent(m.Current, "", "  ")
	if err != nil {
		return err
	}

	// Klasörün var olduğundan emin ol
	dir := filepath.Dir(m.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(m.FilePath, data, 0644)
}

// UpdateResource, belirli bir kaynağın durumunu günceller ve kaydeder.
func (m *Manager) UpdateResource(resType, name, targetState, status string) error {
	m.mu.Lock()
	id := fmt.Sprintf("%s:%s", resType, name)

	entry := ResourceEntry{
		ID:          id,
		Name:        name,
		Type:        resType,
		State:       targetState,
		Status:      status,
		LastApplied: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	m.Current.Resources[id] = entry
	m.mu.Unlock()

	// Her güncellemede diske yazmak güvenlidir (crash durumlarına karşı)
	return m.Save()
}
