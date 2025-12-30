package pkgmngs

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/core"
	"github.com/melih-ucgun/monarch/internal/resource"
)

type ParuAdapter struct {
	resource.BaseResource
	State string
}

func NewParuAdapter(name string, state string) *ParuAdapter {
	if state == "" {
		state = "present"
	}
	return &ParuAdapter{
		BaseResource: resource.BaseResource{Name: name, Type: "package"},
		State:        state,
	}
}

func (r *ParuAdapter) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("package name is required for paru")
	}
	return nil
}

func (r *ParuAdapter) Check(ctx *core.SystemContext) (bool, error) {
	installed := isInstalled("paru", "-Qi", r.Name)
	if r.State == "absent" {
		return installed, nil
	}
	return !installed, nil
}

func (r *ParuAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, _ := r.Check(ctx)
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Package %s already %s", r.Name, r.State)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] paru %s %s", r.State, r.Name)), nil
	}

	var args []string
	if r.State == "absent" {
		args = []string{"-Rns", "--noconfirm", r.Name}
	} else {
		args = []string{"-S", "--noconfirm", "--needed", r.Name}
	}

	out, err := runCommand("paru", args...)
	if err != nil {
		return core.Failure(err, "Paru failed: "+out), err
	}

	return core.SuccessChange(fmt.Sprintf("Paru processed %s", r.Name)), nil
}
