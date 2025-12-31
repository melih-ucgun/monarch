package pkg

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/core"
)

type BrewAdapter struct {
	core.BaseResource
	State string
}

func NewBrewAdapter(name string, state string) *BrewAdapter {
	if state == "" {
		state = "present"
	}
	return &BrewAdapter{
		BaseResource: core.BaseResource{Name: name, Type: "package"},
		State:        state,
	}
}

func (r *BrewAdapter) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("package name is required for brew")
	}
	return nil
}

func (r *BrewAdapter) Check(ctx *core.SystemContext) (bool, error) {
	// brew list <package>
	installed := isInstalled("brew", "list", r.Name)

	if r.State == "absent" {
		return installed, nil
	}
	return !installed, nil
}

func (r *BrewAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, _ := r.Check(ctx)
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Package %s already %s", r.Name, r.State)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] brew %s %s", r.State, r.Name)), nil
	}

	var args []string
	if r.State == "absent" {
		args = []string{"uninstall", r.Name}
	} else {
		args = []string{"install", r.Name}
	}

	out, err := runCommand("brew", args...)
	if err != nil {
		return core.Failure(err, "Brew failed: "+out), err
	}

	return core.SuccessChange(fmt.Sprintf("Brew processed %s", r.Name)), nil
}
