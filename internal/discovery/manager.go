package discovery

import (
	"fmt"
	"strings"

	"github.com/melih-ucgun/veto/internal/core"
)

type PackageManagerType string

const (
	PkgMgrPacman PackageManagerType = "pacman"
	PkgMgrParu   PackageManagerType = "paru"
	PkgMgrYay    PackageManagerType = "yay"
	PkgMgrApt    PackageManagerType = "apt"
	PkgMgrDnf    PackageManagerType = "dnf"
	PkgMgrRpm    PackageManagerType = "rpm"
	PkgMgrBrew   PackageManagerType = "brew"
	PkgMgrUnknown PackageManagerType = ""
)

// DetectManager identifies the primary package manager available on the system
func DetectManager(ctx *core.SystemContext) PackageManagerType {
	if isCommandAvailable(ctx, "paru") {
		return PkgMgrParu
	}
	if isCommandAvailable(ctx, "yay") {
		return PkgMgrYay
	}
	if isCommandAvailable(ctx, "pacman") {
		return PkgMgrPacman
	}
	if isCommandAvailable(ctx, "apt") {
		return PkgMgrApt
	}
	if isCommandAvailable(ctx, "dnf") {
		return PkgMgrDnf
	}
	if isCommandAvailable(ctx, "rpm") {
		return PkgMgrRpm
	}
	if isCommandAvailable(ctx, "brew") {
		return PkgMgrBrew
	}
	return PkgMgrUnknown
}


// GetPackageFiles returns the list of files installed by a package
func GetPackageFiles(ctx *core.SystemContext, mgr PackageManagerType, pkgName string) ([]string, error) {
	var cmd string
	var args []string

	switch mgr {
	case PkgMgrPacman:
		cmd = "pacman"
		args = []string{"-Qlq", pkgName}
	case PkgMgrParu:
		cmd = "paru"
		args = []string{"-Qlq", pkgName}
	case PkgMgrYay:
		cmd = "yay"
		args = []string{"-Qlq", pkgName}
	case PkgMgrApt:
		cmd = "dpkg"
		args = []string{"-L", pkgName}
	case PkgMgrDnf, PkgMgrRpm:
		cmd = "rpm"
		args = []string{"-ql", pkgName}
	case PkgMgrBrew:
		cmd = "brew"
		args = []string{"list", "--verbose", pkgName}
	default:
		return nil, fmt.Errorf("unsupported package manager for file listing: %s", mgr)
	}

	fullCmd := cmd + " " + strings.Join(args, " ")
	output, err := ctx.Transport.Execute(ctx.Context, fullCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list files for %s: %w", pkgName, err)
	}

	lines := parseLines([]byte(output))
	var files []string
	for _, line := range lines {
		// Filter out directories (trailing slash) or empty lines if necessary
		// pacman -Qlq returns directories with trailing / usually, but we should verify.
		// apt/dpkg returns dirs too.
		// We generally only want files for config scanning.
		if line != "" && !strings.HasSuffix(line, "/") {
			files = append(files, line)
		}
	}

	return files, nil
}
