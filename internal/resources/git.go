package resources

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type GitResource struct {
	RepoURL string
	Dest    string
	Branch  string
}

func (g *GitResource) ID() string {
	return fmt.Sprintf("git:%s", g.Dest)
}

func (g *GitResource) Check() (bool, error) {
	info, err := os.Stat(g.Dest)
	if os.IsNotExist(err) {
		return false, nil
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s bir dizin değil", g.Dest)
	}

	// Daha güçlü kontrol: İçeride .git var mı?
	gitDir := filepath.Join(g.Dest, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Dizin var ama git reposu değil.
		// Bu durumda Apply çalışmalı ama "git clone" hata verebilir (dizin boş değilse).
		// Eğer dizin boşsa sorun yok. Doluysa çakışma var demektir.

		// Şimdilik basitçe: "Git reposu değilse false dön" diyoruz.
		return false, nil
	}

	return true, nil
}

func (g *GitResource) Apply() error {
	cmd := exec.Command("git", "clone", g.RepoURL, g.Dest)
	if g.Branch != "" {
		cmd.Args = append(cmd.Args, "-b", g.Branch)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone hatası: %s", string(out))
	}
	return nil
}

func (g *GitResource) Diff() (string, error) {
	exists, _ := g.Check()
	if !exists {
		return fmt.Sprintf("+ git clone: %s -> %s", g.RepoURL, g.Dest), nil
	}
	return "", nil
}

func (g *GitResource) Undo(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if g.Dest == "/" || g.Dest == "" {
		return fmt.Errorf("tehlikeli silme işlemi engellendi: %s", g.Dest)
	}
	// Güvenlik için sadece .git klasörü varsa silmeye izin verilebilir
	// Şimdilik manuel bırakıyoruz veya kullanıcıya bırakıyoruz.
	// return os.RemoveAll(g.Dest)
	return nil
}
