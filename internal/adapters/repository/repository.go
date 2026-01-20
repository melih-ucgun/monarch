package repository

import (
	"fmt"
	"strings"

	"github.com/melih-ucgun/veto/internal/core"
)

// Init registers the repository resource.
func init() {
	core.RegisterResource("repository", NewRepositoryResource)
}

func NewRepositoryResource(name string, params map[string]interface{}, ctx *core.SystemContext) (core.Resource, error) {
	// 1. Check if manager is explicitly requested
	manager, _ := params["manager"].(string)

	// 2. If not, auto-detect based on context
	if manager == "" {
		switch ctx.Distro {
		case "ubuntu", "debian", "pop", "mint", "kali":
			manager = "apt"
		case "fedora", "rhel", "centos", "almalinux":
			manager = "dnf"
		case "arch", "manjaro", "endeavouros":
			manager = "pacman"
		default:
			// Fallback based on available binary?
			if isCommandAvailable(ctx, "apt-get") {
				manager = "apt"
			} else if isCommandAvailable(ctx, "dnf") {
				manager = "dnf"
			} else {
				return nil, fmt.Errorf("unsupported distro '%s' for repository resource, please specify 'manager'", ctx.Distro)
			}
		}
	}

	switch manager {
	case "apt":
		return NewAptRepository(name, params), nil
	case "dnf":
		return NewDnfRepository(name, params), nil
	case "pacman":
		return nil, fmt.Errorf("repository management for pacman is not yet implemented (manual editing required)")
	default:
		return nil, fmt.Errorf("unknown repository manager: %s", manager)
	}
}

func isCommandAvailable(ctx *core.SystemContext, cmd string) bool {
	// Simple check, assumes 'which' or similar exists, or we just rely on distro for now.
	// Since we don't have a robust check in this scope easily without running command.
	// We'll skip for now and trust distro mapping.
	return false
}
