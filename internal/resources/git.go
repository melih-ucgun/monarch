package resources

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type GitResource struct {
	CanonicalID string `mapstructure:"-"`
	RepoURL     string `mapstructure:"repo"`
	Dest        string `mapstructure:"dest"`
	Branch      string `mapstructure:"branch"`
}

func (r *GitResource) ID() string {
	return r.CanonicalID
}

func (r *GitResource) Check() (bool, error) {
	// Dizin var mı?
	if _, err := os.Stat(r.Dest); os.IsNotExist(err) {
		return false, nil
	}
	// .git klasörü var mı? (Basit bir repo check)
	gitDir := fmt.Sprintf("%s/.git", r.Dest)
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false, nil
	}

	// TODO: Branch check eklenebilir (git rev-parse --abbrev-ref HEAD)
	return true, nil
}

func (r *GitResource) Apply() error {
	if _, err := os.Stat(r.Dest); os.IsNotExist(err) {
		// Dizin yok, Clone yap
		args := []string{"clone", r.RepoURL, r.Dest}
		if r.Branch != "" {
			args = append(args, "--branch", r.Branch)
		}

		cmd := exec.Command("git", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone hatası: %s", string(out))
		}
	} else {
		// Dizin var, Pull yap
		// Not: Eğer uncommitted changes varsa fail olabilir, force pull gerekebilir.
		cmd := exec.Command("git", "-C", r.Dest, "pull")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git pull hatası: %s", string(out))
		}
	}
	return nil
}

func (r *GitResource) Undo(ctx context.Context) error {
	return os.RemoveAll(r.Dest)
}

func (r *GitResource) Diff() (string, error) {
	return fmt.Sprintf("Git repo[%s] missing or outdated", r.Dest), nil
}
