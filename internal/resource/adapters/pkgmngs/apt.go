package pkg

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/core"
	"github.com/melih-ucgun/monarch/internal/resource"
)

type AptAdapter struct {
	resource.BaseResource
	State string
}

func NewAptAdapter(name string, state string) *AptAdapter {
	if state == "" {
		state = "present"
	}
	return &AptAdapter{
		BaseResource: resource.BaseResource{Name: name, Type: "package"},
		State:        state,
	}
}

func (r *AptAdapter) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("package name is required")
	}
	return nil
}

func (r *AptAdapter) Check(ctx *core.SystemContext) (bool, error) {
	// dpkg -s <package>
	installed := isInstalled("dpkg", "-s", r.Name)

	if r.State == "absent" {
		return installed, nil
	}
	return !installed, nil
}

func (r *AptAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, _ := r.Check(ctx)
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Package %s already %s", r.Name, r.State)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] apt-get %s %s", r.State, r.Name)), nil
	}

	var args []string
	if r.State == "absent" {
		args = []string{"remove", "-y", r.Name}
	} else {
		args = []string{"install", "-y", r.Name}
	}

	out, err := runCommand("apt-get", args...)
	if err != nil {
		return core.Failure(err, "Apt failed: "+out), err
	}

	return core.SuccessChange(fmt.Sprintf("Apt processed %s", r.Name)), nil
}
