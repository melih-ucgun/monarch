package resources

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/config"
)

func New(cfg config.ResourceConfig, vars map[string]string) (Resource, error) {
	switch cfg.Type {
	case "file":
		mode := "0644"
		if m, ok := cfg.Parameters["mode"].(string); ok {
			mode = m
		}
		content := ""
		if c, ok := cfg.Parameters["content"].(string); ok {
			content = c
		}
		path := ""
		if p, ok := cfg.Parameters["path"].(string); ok {
			path = p
		}

		return &FileResource{
			CanonicalID: cfg.ID,
			Path:        path,
			Content:     content,
			Mode:        mode,
		}, nil

	case "exec":
		cmd := ""
		if c, ok := cfg.Parameters["command"].(string); ok {
			cmd = c
		}
		unless := ""
		if u, ok := cfg.Parameters["unless"].(string); ok {
			unless = u
		}
		runAs := ""
		if r, ok := cfg.Parameters["run_as"].(string); ok {
			runAs = r
		}

		return &ExecResource{
			CanonicalID: cfg.ID,
			Command:     cmd,
			Unless:      unless,
			RunAsUser:   runAs,
		}, nil

	case "package":
		name := ""
		if n, ok := cfg.Parameters["name"].(string); ok {
			name = n
		}

		mgrName := "pacman"
		if m, ok := cfg.Parameters["manager"].(string); ok {
			mgrName = m
		}

		state := "installed"
		if s, ok := cfg.Parameters["state"].(string); ok {
			state = s
		}

		return &PackageResource{
			CanonicalID: cfg.ID,
			Name:        name,
			ManagerName: mgrName,
			State:       state,
		}, nil

	case "service":
		name := ""
		if n, ok := cfg.Parameters["name"].(string); ok {
			name = n
		}

		state := "started"
		if s, ok := cfg.Parameters["state"].(string); ok {
			state = s
		}

		enabled := true
		if e, ok := cfg.Parameters["enabled"].(bool); ok {
			enabled = e
		}

		return &ServiceResource{
			CanonicalID: cfg.ID,
			Name:        name,
			State:       state,
			Enabled:     enabled,
		}, nil

	case "git":
		repo := ""
		if r, ok := cfg.Parameters["repo"].(string); ok {
			repo = r
		}
		dest := ""
		if d, ok := cfg.Parameters["dest"].(string); ok {
			dest = d
		}
		branch := ""
		if b, ok := cfg.Parameters["branch"].(string); ok {
			branch = b
		}

		return &GitResource{
			RepoURL: repo,
			Dest:    dest,
			Branch:  branch,
		}, nil

	case "symlink":
		target := ""
		if t, ok := cfg.Parameters["target"].(string); ok {
			target = t
		}
		link := ""
		if l, ok := cfg.Parameters["link"].(string); ok {
			link = l
		}
		force := false
		if f, ok := cfg.Parameters["force"].(bool); ok {
			force = f
		}

		return &SymlinkResource{
			Target: target,
			Link:   link,
			Force:  force,
		}, nil

	case "container":
		name := ""
		if n, ok := cfg.Parameters["name"].(string); ok {
			name = n
		}
		image := ""
		if i, ok := cfg.Parameters["image"].(string); ok {
			image = i
		}
		state := "running"
		if s, ok := cfg.Parameters["state"].(string); ok {
			state = s
		}

		// Portları []interface{}'den []string'e çevir
		var ports []string
		if pList, ok := cfg.Parameters["ports"].([]interface{}); ok {
			for _, p := range pList {
				if pStr, ok := p.(string); ok {
					ports = append(ports, pStr)
				}
			}
		}

		return &ContainerResource{
			Name:  name,
			Image: image,
			State: state,
			Ports: ports,
		}, nil

	default:
		return nil, fmt.Errorf("bilinmeyen kaynak tipi: %s", cfg.Type)
	}
}
