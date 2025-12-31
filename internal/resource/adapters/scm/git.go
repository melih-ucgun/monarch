package scm

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/melih-ucgun/monarch/internal/core"
)

type GitAdapter struct {
	core.BaseResource
	Repo   string
	Dest   string
	Branch string
	State  string // present (clone/pull), absent (delete)
}

func NewGitAdapter(name string, params map[string]interface{}) *GitAdapter {
	repo, _ := params["repo"].(string)
	dest, _ := params["dest"].(string)
	branch, _ := params["branch"].(string)
	state, _ := params["state"].(string)

	if state == "" {
		state = "present"
	}
	if branch == "" {
		branch = "main"
	}

	return &GitAdapter{
		BaseResource: core.BaseResource{Name: name, Type: "git"},
		Repo:         repo,
		Dest:         dest,
		Branch:       branch,
		State:        state,
	}
}

func (r *GitAdapter) Validate() error {
	if r.Repo == "" {
		return fmt.Errorf("git repository url is required")
	}
	if r.Dest == "" {
		return fmt.Errorf("git destination path is required")
	}
	return nil
}

func (r *GitAdapter) Check(ctx *core.SystemContext) (bool, error) {
	if r.State == "absent" {
		if _, err := os.Stat(r.Dest); !os.IsNotExist(err) {
			return true, nil // Folder exists, need to delete
		}
		return false, nil
	}

	// State == present
	if _, err := os.Stat(r.Dest); os.IsNotExist(err) {
		return true, nil // Folder doesn't exist, need to clone
	}

	// Folder exists, check if it's the right repo/branch?
	// For simplicity, we assume if it exists, we might need to pull.
	// But Check usually returns "needsAction". "Always pull" might be detected here or we assume if it exists we are good unless we want to enforce latest.
	// Let's assume if it exists, we return false (no change needed) unless we want to support "latest".
	// For now, let's keep it simple: if exists, we are good.
	// TODO: Implement "update" or strict checking.
	return false, nil
}

func (r *GitAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, err := r.Check(ctx)
	if err != nil {
		return core.Failure(err, "Check failed"), err
	}
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Git repo %s already in desired state at %s", r.Repo, r.Dest)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] Git %s %s to %s", r.State, r.Repo, r.Dest)), nil
	}

	if r.State == "absent" {
		if err := os.RemoveAll(r.Dest); err != nil {
			return core.Failure(err, "Failed to remove directory"), err
		}
		return core.SuccessChange(fmt.Sprintf("Removed git repo at %s", r.Dest)), nil
	}

	// State == present -> Clone
	cmd := exec.Command("git", "clone", "-b", r.Branch, r.Repo, r.Dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return core.Failure(err, fmt.Sprintf("Git clone failed: %s", string(out))), err
	}

	return core.SuccessChange(fmt.Sprintf("Cloned %s to %s", r.Repo, r.Dest)), nil
}
