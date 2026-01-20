package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/melih-ucgun/veto/internal/core"
)

// DiscoveredConfig represents a found configuration file and its relationship to a package.
type DiscoveredConfig struct {
	Path      string
	PackageID string // The package that "owns" or is related to this config
}

// CommonConfigMap maps package/tool names to their standard config paths.
// Paths starting with "~" will be expanded using userHome.
var CommonConfigMap = map[string][]string{
	// Shells
	"zsh":  {"~/.zshrc", "~/.zshenv", "~/.zprofile"},
	"bash": {"~/.bashrc", "~/.bash_profile", "~/.bash_aliases"},
	"fish": {"~/.config/fish/config.fish"},

	// Editors
	"vim":    {"~/.vimrc"},
	"neovim": {"~/.config/nvim/init.lua", "~/.config/nvim/init.vim"},
	"nvim":   {"~/.config/nvim/init.lua", "~/.config/nvim/init.vim"}, // Alias
	"nano":   {"~/.nanorc"},
	"emacs":  {"~/.emacs", "~/.emacs.d/init.el"},
	"vscode": {"~/.config/Code/User/settings.json", "~/.config/Code/User/keybindings.json"},
	"code":   {"~/.config/Code/User/settings.json", "~/.config/Code/User/keybindings.json"},

	// Tools
	"git":       {"~/.gitconfig"},
	"ssh":       {"~/.ssh/config"},
	"tmux":      {"~/.tmux.conf"},
	"alacritty": {"~/.config/alacritty/alacritty.yml", "~/.config/alacritty/alacritty.toml"},
	"kitty":     {"~/.config/kitty/kitty.conf"},
	"starship":  {"~/.config/starship.toml"},
	"gh":        {"~/.config/gh/config.yml"},
	"i3":        {"~/.config/i3/config"},
	"sway":      {"~/.config/sway/config"},
	"hyprland":  {"~/.config/hypr/hyprland.conf"},
	"waybar":    {"~/.config/waybar/config", "~/.config/waybar/style.css"},
	"rofi":      {"~/.config/rofi/config.rasi"},
	"wofi":      {"~/.config/wofi/config"},

	// Services (System-wide)
	"nginx":    {"/etc/nginx/nginx.conf"},
	"docker":   {"/etc/docker/daemon.json"},
	"samba":    {"/etc/samba/smb.conf"},
	"sshd":     {"/etc/ssh/sshd_config"},
	"postgres": {"/var/lib/postgres/data/postgresql.conf"}, // Varies heavily
}

// DiscoverConfigs suggests config files based on selected packages using a hybrid approach:
// 1. Static Map (CommonConfigMap)
// 2. Package Manager Query (System Configs)
// 3. Name Heuristics (User Configs)
func DiscoverConfigs(ctx *core.SystemContext, packages []string) ([]DiscoveredConfig, error) {
	var foundConfigs []DiscoveredConfig
	seenPaths := make(map[string]bool)

	userHome := ctx.Cwd
	if home, err := os.UserHomeDir(); err == nil {
		userHome = home
	}

	// Detect Package Manager for System Query
	mgr := DetectManager(ctx)

	for _, pkg := range packages {
		pkgName := strings.ToLower(pkg)

		// Strategy 1: Static Map (Legacy/Reliable for finding "hidden" configs)
		if paths, ok := CommonConfigMap[pkgName]; ok {
			for _, p := range paths {
				absPath := expandPath(p, userHome)
				if fileExists(absPath) && !seenPaths[absPath] {
					foundConfigs = append(foundConfigs, DiscoveredConfig{
						Path:      absPath,
						PackageID: pkg,
					})
					seenPaths[absPath] = true
				}
			}
		}

		// Strategy 2: Package Manager Query (System Configs in /etc)
		// This finds configs that the package manager actually tracks.
		if mgr != PkgMgrUnknown {
			files, err := GetPackageFiles(ctx, mgr, pkg)
			if err == nil {
				for _, f := range files {
					// We are mostly interested in /etc for system configs
					if strings.HasPrefix(f, "/etc/") && !isDirectory(f) && !seenPaths[f] {
						// Filter out non-config files if possible (e.g. binaries in /etc? rare)
						// For now accept all /etc files, maybe filter mainly .conf, .yaml, .json, .xml etc?
						// Or just take all. Taking all might be too much. 
						// Heuristic: Must be a text file or have config extension?
						// Let's stick to key config extensions or specific common names to avoid trash.
						// Actually, pacman -Qlq returns ALL files.
						// A safer bet is: if it is in /etc/pkgName/... or /etc/pkgName.conf
						if strings.Contains(f, "/"+pkgName+"/") || strings.HasSuffix(f, "/"+pkgName+".conf") || strings.HasSuffix(f, ".conf") || strings.HasSuffix(f, ".yaml") || strings.HasSuffix(f, ".json") || strings.HasSuffix(f, ".toml") {
							foundConfigs = append(foundConfigs, DiscoveredConfig{
								Path:      f,
								PackageID: pkg,
							})
							seenPaths[f] = true
						}
					}
				}
			}
		}

		// Strategy 3: Name Heuristics (User Configs)
		// Try to guess ~/.config/<pkg>/...
		// Common patterns:
		// ~/.config/<pkg>/config
		// ~/.config/<pkg>/<pkg>.conf
		// ~/.<pkg>rc
		potentialPaths := []string{
			fmt.Sprintf("~/.config/%s/config", pkgName),
			fmt.Sprintf("~/.config/%s/config.toml", pkgName),
			fmt.Sprintf("~/.config/%s/config.yaml", pkgName),
			fmt.Sprintf("~/.config/%s/config.yml", pkgName),
			fmt.Sprintf("~/.config/%s/config.json", pkgName),
			fmt.Sprintf("~/.config/%s/%s.conf", pkgName, pkgName),
			fmt.Sprintf("~/.%src", pkgName),
		}

		for _, p := range potentialPaths {
			absPath := expandPath(p, userHome)
			if fileExists(absPath) && !seenPaths[absPath] {
				foundConfigs = append(foundConfigs, DiscoveredConfig{
					Path:      absPath,
					PackageID: pkg,
				})
				seenPaths[absPath] = true
			}
		}
	}

	return foundConfigs, nil
}

func expandPath(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
