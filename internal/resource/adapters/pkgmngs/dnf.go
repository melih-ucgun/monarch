package pkgmngs

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/core"
	"github.com/melih-ucgun/monarch/internal/resource"
)

type DnfAdapter struct {
	resource.BaseResource
	State string
}

func NewDnfAdapter(name string, state string) *DnfAdapter {
	if state == "" {
		state = "present"
	}
	return &DnfAdapter{
		BaseResource: resource.BaseResource{Name: name, Type: "package"},
		State:        state,
	}
}

func (r *DnfAdapter) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("package name is required for dnf")
	}
	return nil
}

func (r *DnfAdapter) Check(ctx *core.SystemContext) (bool, error) {
	installed := isInstalled("rpm", "-q", r.Name)
	if r.State == "absent" {
		return installed, nil
	}
	return !installed, nil
}

func (r *DnfAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, _ := r.Check(ctx)
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Package %s already %s", r.Name, r.State)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] dnf %s %s", r.State, r.Name)), nil
	}

	var args []string
	if r.State == "absent" {
		args = []string{"remove", "-y", r.Name}
	} else {
		args = []string{"install", "-y", r.Name}
	}

	out, err := runCommand("dnf", args...)
	if err != nil {
		return core.Failure(err, "Dnf failed: "+out), err
	}

	return core.SuccessChange(fmt.Sprintf("Dnf processed %s", r.Name)), nil
}
