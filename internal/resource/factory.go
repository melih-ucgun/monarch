package resource

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/adapters/bundle"
	"github.com/melih-ucgun/monarch/internal/adapters/file"
	"github.com/melih-ucgun/monarch/internal/adapters/git"
	"github.com/melih-ucgun/monarch/internal/adapters/identity"
	"github.com/melih-ucgun/monarch/internal/adapters/pkg"
	"github.com/melih-ucgun/monarch/internal/adapters/service"
	"github.com/melih-ucgun/monarch/internal/adapters/shell"
	"github.com/melih-ucgun/monarch/internal/core"
)

// Deprecated fonksiyon placeholder
func CreateResource(resType string, name string, state string, ctx *core.SystemContext) (core.Resource, error) {
	return nil, fmt.Errorf("use CreateResourceWithParams")
}

// CreateResourceWithParams artık core.Resource döndürüyor
func CreateResourceWithParams(resType string, name string, params map[string]interface{}, ctx *core.SystemContext) (core.Resource, error) {

	stateParam, _ := params["state"].(string)
	if stateParam == "" {
		stateParam = "present"
	}

	switch resType {
	// Package Managers
	case "pacman":
		return pkg.NewPacmanAdapter(name, stateParam), nil
	case "apt":
		return pkg.NewAptAdapter(name, stateParam), nil
	case "dnf":
		return pkg.NewDnfAdapter(name, stateParam), nil
	case "brew":
		return pkg.NewBrewAdapter(name, stateParam), nil
	case "apk":
		return pkg.NewApkAdapter(name, stateParam), nil
	case "flatpak":
		return pkg.NewFlatpakAdapter(name, stateParam), nil
	case "snap":
		return pkg.NewSnapAdapter(name, stateParam), nil
	case "zypper":
		return pkg.NewZypperAdapter(name, stateParam), nil
	case "yum":
		return pkg.NewYumAdapter(name, stateParam), nil
	case "paru":
		return pkg.NewParuAdapter(name, stateParam), nil
	case "yay":
		return pkg.NewYayAdapter(name, stateParam), nil
	case "package", "pkg":
		return detectPackageManager(name, stateParam, ctx)

	// Filesystem
	case "file":
		params["state"] = stateParam
		return file.NewFileAdapter(name, params), nil
	case "symlink":
		params["state"] = stateParam
		return file.NewSymlinkAdapter(name, params), nil
	case "archive", "extract":
		return file.NewArchiveAdapter(name, params), nil
	case "download":
		return file.NewDownloadAdapter(name, params), nil
	case "template":
		return file.NewTemplateAdapter(name, params), nil
	case "line_in_file", "lineinfile":
		params["state"] = stateParam
		return file.NewLineInFileAdapter(name, params), nil

	// Identity
	case "user":
		params["state"] = stateParam
		return identity.NewUserAdapter(name, params), nil
	case "group":
		params["state"] = stateParam
		return identity.NewGroupAdapter(name, params), nil

	// Others
	case "git":
		params["state"] = stateParam
		return git.NewGitAdapter(name, params), nil
	case "service", "systemd":
		params["state"] = stateParam
		return service.NewServiceAdapter(name, params, ctx), nil
	case "exec", "shell", "cmd":
		return shell.NewExecAdapter(name, params), nil

	// Bundle
	case "bundle":
		// Recursively pass this factory function
		return bundle.NewBundleAdapter(name, params, CreateResourceWithParams), nil

	default:
		return nil, fmt.Errorf("unknown resource type: %s", resType)
	}
}

func detectPackageManager(name, state string, ctx *core.SystemContext) (core.Resource, error) {
	switch ctx.Distro {
	case "arch", "cachyos", "manjaro", "endeavouros":
		return pkg.NewPacmanAdapter(name, state), nil
	case "ubuntu", "debian", "pop", "mint", "kali":
		return pkg.NewAptAdapter(name, state), nil
	case "fedora", "rhel", "centos", "almalinux":
		return pkg.NewDnfAdapter(name, state), nil
	case "alpine":
		return pkg.NewApkAdapter(name, state), nil
	case "opensuse", "sles":
		return pkg.NewZypperAdapter(name, state), nil
	case "darwin":
		return pkg.NewBrewAdapter(name, state), nil
	default:
		return nil, fmt.Errorf("automatic package manager detection failed")
	}
}
