package file

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/melih-ucgun/veto/internal/core"
)

type SymlinkAdapter struct {
	core.BaseResource
	Link   string // Linkin oluşacağı yer
	Target string // Linkin hedefi
	State  string
}

func NewSymlinkAdapter(name string, params map[string]interface{}) *SymlinkAdapter {
	link, _ := params["path"].(string)
	if link == "" {
		link = name
	}

	target, _ := params["target"].(string)
	state, _ := params["state"].(string)
	if state == "" {
		state = "present"
	}

	return &SymlinkAdapter{
		BaseResource: core.BaseResource{Name: name, Type: "symlink"},
		Link:         link,
		Target:       target,
		State:        state,
	}
}

func (r *SymlinkAdapter) Validate() error {
	if r.Link == "" || r.Target == "" {
		return fmt.Errorf("symlink requires both 'path' (link) and 'target'")
	}
	return nil
}

func (r *SymlinkAdapter) Check(ctx *core.SystemContext) (bool, error) {
	info, err := os.Lstat(r.Link)

	if r.State == "absent" {
		return !os.IsNotExist(err), nil
	}

	if os.IsNotExist(err) {
		return true, nil
	}

	// Link mi?
	if info.Mode()&os.ModeSymlink == 0 {
		return true, nil // Dosya var ama link değil, düzeltilmeli
	}

	// Hedef doğru mu?
	currentDest, err := os.Readlink(r.Link)
	if err != nil {
		return true, err
	}

	// Absolute path kontrolü yapmak daha sağlıklı olabilir
	return currentDest != r.Target, nil
}

func (r *SymlinkAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, _ := r.Check(ctx)
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Symlink %s -> %s correct", r.Link, r.Target)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] Link %s -> %s", r.Link, r.Target)), nil
	}

	// Varsa sil (force update)
	os.Remove(r.Link)

	if r.State == "absent" {
		return core.SuccessChange("Symlink removed"), nil
	}

	// Klasörü oluştur
	os.MkdirAll(filepath.Dir(r.Link), 0755)

	if err := os.Symlink(r.Target, r.Link); err != nil {
		return core.Failure(err, "Failed to create symlink"), err
	}

	return core.SuccessChange(fmt.Sprintf("Symlink created: %s -> %s", r.Link, r.Target)), nil
}
