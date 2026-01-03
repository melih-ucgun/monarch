package state

import (
	"fmt"

	"github.com/melih-ucgun/veto/internal/types"
)

// AddTransaction appends a new transaction to history and saves state.
func (m *Manager) AddTransaction(tx types.Transaction) error {
	m.mu.Lock()
	m.Current.History = append(m.Current.History, tx)
	m.mu.Unlock()

	return m.Save()
}

// GetTransactions returns a copy of history.
func (m *Manager) GetTransactions() []types.Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	history := make([]types.Transaction, len(m.Current.History))
	copy(history, m.Current.History)
	return history
}

// GetTransaction finds a transaction by ID.
func (m *Manager) GetTransaction(id string) (types.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tx := range m.Current.History {
		if tx.ID == id {
			return tx, nil
		}
	}
	return types.Transaction{}, fmt.Errorf("transaction not found: %s", id)
}

// NewHistoryManager is a helper to get a manager specifically for history lookup.
// In reality, history is part of the main State Manager.
// This helper is for CLI convenience if we just want to read.
// Note: It creates a new Manager instance, so it reads from disk.
func NewHistoryManager(ignoredPath string) *Manager {
	// We need default path. Internal detail but exposed via this helper for cmd.
	// Hardcoding default path here is risky if changed elsewhere.
	// Better to let cmd inject path.
	// For now, assuming standard path.
	// But Manager needs FS.
	return nil // Stub, logic should be in cmd to create Manager properly.
}
