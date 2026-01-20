package repository

import (
	"fmt"
	"strings"

	"github.com/melih-ucgun/veto/internal/core"
)

type DnfRepository struct {
	core.BaseResource
	Source string // URL or Copr name (user/project)
	State  string
	IsCopr bool
}

func NewDnfRepository(name string, params map[string]interface{}) *DnfRepository {
	state, _ := params["state"].(string)
	if state == "" {
		state = "present"
	}
	source, _ := params["source"].(string)
	if source == "" {
		source = name
	}
	
	// Detect Copr
	// If source looks like "group/project" and no http, assume Copr?
	// Or user must specify 'copr: true'?
	// Simplify: If it doesn't start with http/https/ftp, assume Copr for DNF?
	// Or check param.
	isCopr, _ := params["copr"].(bool)
	if !isCopr && !strings.HasPrefix(source, "http") {
		// Heuristic: DNF config-manager needs a URL.
		isCopr = true
	}

	return &DnfRepository{
		BaseResource: core.BaseResource{Name: name, Type: "repository"},
		Source:       source,
		State:        state,
		IsCopr:       isCopr,
	}
}

func (r *DnfRepository) Validate(ctx *core.SystemContext) error {
	if r.Source == "" {
		return fmt.Errorf("repository source is required")
	}
	return nil
}

func (r *DnfRepository) Check(ctx *core.SystemContext) (bool, error) {
	// check if /etc/yum.repos.d/ contains a file with this name?
	// dnf repolist enabled | grep match
	
	// For Copr: dnf copr list
	
	cmd := "dnf repolist enabled"
	out, err := runCommand(ctx, "sh", "-c", cmd)
	if err != nil {
		return false, err
	}

	// Very loose check
	exists := strings.Contains(out, r.Name) || strings.Contains(out, r.Source)
	
	if r.State == "absent" {
		return exists, nil
	}
	return !exists, nil
}

func (r *DnfRepository) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, _ := r.Check(ctx)
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Repository %s already %s", r.Source, r.State)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] dnf add repo %s", r.Source)), nil
	}

	var err error
	var out string

	if r.IsCopr {
		action := "enable"
		if r.State == "absent" {
			action = "disable" // or remove
		}
		// dnf copr enable -y user/project
		out, err = runCommand(ctx, "dnf", "copr", action, "-y", r.Source)
	} else {
		if r.State == "absent" {
			// Removing a regular repo usually means deleting the file using config-manager --disable?
			// or rm /etc/yum.repos.d/X.repo?
			// 'dnf config-manager --disable' is safer.
			out, err = runCommand(ctx, "dnf", "config-manager", "--set-disabled", r.Name)
		} else {
			out, err = runCommand(ctx, "dnf", "config-manager", "--add-repo", r.Source)
		}
	}

	if err != nil {
		return core.Failure(err, "Failed to manage repo: "+out), err
	}

	return core.SuccessChange("Repository updated"), nil
}
