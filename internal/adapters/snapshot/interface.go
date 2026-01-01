package snapshot

// Provider defines the interface for snapshot tools
type Provider interface {
	Name() string
	IsAvailable() bool
	// CreateSnapshot creates a single snapshot with description
	CreateSnapshot(description string) error
	// CreatePreSnapshot starts a transactional snapshot (returns transaction ID/Handle)
	CreatePreSnapshot(description string) (string, error)
	// CreatePostSnapshot completes a transactional snapshot using the ID from Pre
	CreatePostSnapshot(id string, description string) error
}
