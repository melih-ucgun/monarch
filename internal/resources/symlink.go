package resources

import (
	"context"
	"fmt"
	"os"
)

type SymlinkResource struct {
	Target string // Hedef dosya (gerçek dosya)
	Link   string // Linkin oluşturulacağı yer
	Force  bool
}

func (s *SymlinkResource) ID() string {
	return fmt.Sprintf("symlink:%s", s.Link)
}

func (s *SymlinkResource) Check() (bool, error) {
	info, err := os.Lstat(s.Link)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return false, fmt.Errorf("%s mevcut fakat bir symlink değil", s.Link)
	}

	dest, err := os.Readlink(s.Link)
	if err != nil {
		return false, err
	}

	if dest != s.Target {
		return false, nil
	}

	return true, nil
}

func (s *SymlinkResource) Apply() error {
	if s.Force {
		_ = os.Remove(s.Link)
	}

	if err := os.Symlink(s.Target, s.Link); err != nil {
		return fmt.Errorf("symlink oluşturulamadı: %w", err)
	}
	return nil
}

func (s *SymlinkResource) Diff() (string, error) {
	exists, err := s.Check()
	if err != nil {
		return "", err
	}
	if !exists {
		return fmt.Sprintf("+ symlink: %s -> %s", s.Link, s.Target), nil
	}
	return "", nil
}

func (s *SymlinkResource) Undo(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// return os.Remove(s.Link)
	return nil
}
