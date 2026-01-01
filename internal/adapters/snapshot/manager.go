package snapshot

import "github.com/pterm/pterm"

// Manager handles the selection and execution of snapshot providers
type Manager struct {
	provider Provider
}

// NewManager detects the best available snapshot provider
// Priority: Snapper (if BTRFS and configured) > Timeshift > None
func NewManager(rootFSType string) *Manager {
	// 1. Try Snapper
	// Snapper is preferred for BTRFS because of atomic/fast snapshots
	snapper := NewSnapper()
	if rootFSType == "btrfs" && snapper.IsAvailable() {
		// Basit bir kontrol: snapper list komutu çalışıyor mu? (Config var mı?)
		// Detaylı kontrolü provider içinde yapmak daha iyi olabilir ama şimdilik availability yeterli.
		return &Manager{provider: snapper}
	}

	// 2. Try Timeshift
	timeshift := NewTimeshift()
	if timeshift.IsAvailable() {
		return &Manager{provider: timeshift}
	}

	// 3. Fallback to Snapper if configured on non-BTRFS (e.g. LVM thin provisioning support in snapper exists too)
	if snapper.IsAvailable() {
		return &Manager{provider: snapper}
	}

	return nil
}

func (m *Manager) IsAvailable() bool {
	return m != nil && m.provider != nil
}

func (m *Manager) ProviderName() string {
	if m.provider == nil {
		return "None"
	}
	return m.provider.Name()
}

func (m *Manager) CreatePreSnapshot(desc string) (string, error) {
	if m.provider == nil {
		return "", nil
	}
	pterm.Info.Printf("Creating system snapshot using %s...\n", m.provider.Name())
	return m.provider.CreatePreSnapshot(desc)
}

func (m *Manager) CreatePostSnapshot(id string, desc string) error {
	if m.provider == nil {
		return nil
	}
	return m.provider.CreatePostSnapshot(id, desc)
}
