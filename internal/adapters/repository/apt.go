package repository

import (
	"fmt"
	"strings"

	"github.com/melih-ucgun/veto/internal/core"
)

type AptRepository struct {
	core.BaseResource
	Source string // The PPA or Repo line
	KeyURL string
	State  string
}

func NewAptRepository(name string, params map[string]interface{}) *AptRepository {
	state, _ := params["state"].(string)
	if state == "" {
		state = "present"
	}
	source, _ := params["source"].(string)
	if source == "" {
		// Use name as source if not provided
		source = name
	}
	keyURL, _ := params["key_url"].(string)

	return &AptRepository{
		BaseResource: core.BaseResource{Name: name, Type: "repository"},
		Source:       source,
		KeyURL:       keyURL,
		State:        state,
	}
}

func (r *AptRepository) Validate(ctx *core.SystemContext) error {
	if r.Source == "" {
		return fmt.Errorf("repository source is required")
	}
	return nil
}

func (r *AptRepository) Check(ctx *core.SystemContext) (bool, error) {
	// Check if repo exists in sources list
	// apt-cache policy doesn't list repos directly in a parseable way easily for existence of PPA.
	// Better: grep /etc/apt/sources.list and /etc/apt/sources.list.d/*
	
	// Simplify: Check if we can find a file in sources.list.d that matches the name?
	// Or simplistic approach: Trust 'add-apt-repository' to handle idempotency?
	// add-apt-repository usually adds duplicates if not careful.
	// Let's try to detect.
	
	// Only run check on Linux
	if ctx.OS != "linux" {
		return false, nil
	}

	// Clean source for searching
	s := r.Source
	if strings.HasPrefix(s, "ppa:") {
		s = strings.TrimPrefix(s, "ppa:")
	}

	// grep -r "source" /etc/apt/sources.list /etc/apt/sources.list.d/
	cmd := fmt.Sprintf("grep -r \"%s\" /etc/apt/sources.list /etc/apt/sources.list.d/", s)
	_, err := runCommand(ctx, "sh", "-c", cmd)
	
	exists := (err == nil)
	
	if r.State == "absent" {
		return exists, nil
	}
	return !exists, nil
}

func (r *AptRepository) Apply(ctx *core.SystemContext) (core.Result, error) {
	needsAction, _ := r.Check(ctx)
	if !needsAction {
		return core.SuccessNoChange(fmt.Sprintf("Repository %s already %s", r.Source, r.State)), nil
	}

	if ctx.DryRun {
		return core.SuccessChange(fmt.Sprintf("[DryRun] add-apt-repository %s %s", r.State, r.Source)), nil
	}

	if r.State == "absent" {
		out, err := runCommand(ctx, "add-apt-repository", "--remove", "-y", r.Source)
		if err != nil {
			return core.Failure(err, "Failed to remove repo: "+out), err
		}
		// Update apt
		runCommand(ctx, "apt-get", "update")
		return core.SuccessChange("Repository removed"), nil
	}

	// Install Key if provided
	if r.KeyURL != "" {
		// New way: signed-by. But add-apt-repository handles keys? Not from URL usually.
		// Legacy safe way: wget -qO - https://... | sudo apt-key add -
		// Only if needed.
		keyCmd := fmt.Sprintf("wget -qO - %s | apt-key add -", r.KeyURL)
		if out, err := runCommand(ctx, "sh", "-c", keyCmd); err != nil {
			return core.Failure(err, "Failed to add key: "+out), err
		}
	}

	out, err := runCommand(ctx, "add-apt-repository", "-y", r.Source)
	if err != nil {
		return core.Failure(err, "Failed to add repo: "+out), err
	}

	// Update apt
	runCommand(ctx, "apt-get", "update")

	return core.SuccessChange("Repository added"), nil
}

// Helper for command execution (duplicated from other adapters, should be shared utils)
func runCommand(ctx *core.SystemContext, name string, args ...string) (string, error) {
	return ctx.Transport.Execute(name, args...)
}
