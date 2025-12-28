package resources

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/config"
)

func New(r config.Resource, vars map[string]interface{}) (Resource, error) {
	fieldsToProcess := map[string]*string{
		"name": &r.Name, "path": &r.Path, "content": &r.Content,
		"image": &r.Image, "target": &r.Target, "mode": &r.Mode,
		"owner": &r.Owner, "group": &r.Group, "command": &r.Command,
		"creates": &r.Creates, "only_if": &r.OnlyIf, "unless": &r.Unless,
	}

	for _, val := range fieldsToProcess {
		if *val != "" {
			if processed, err := config.ExecuteTemplate(*val, vars); err == nil {
				*val = processed
			}
		}
	}

	id := r.Identify()
	switch r.Type {
	case "file":
		return &FileResource{
			CanonicalID: id, ResourceName: r.Name, Path: r.Path,
			Content: r.Content, Mode: r.Mode, Owner: r.Owner, Group: r.Group,
		}, nil
	case "exec":
		return &ExecResource{
			CanonicalID: id, Name: r.Name, Command: r.Command,
			Creates: r.Creates, OnlyIf: r.OnlyIf, Unless: r.Unless,
		}, nil
	case "package":
		return &PackageResource{CanonicalID: id, PackageName: r.Name, State: r.State, Provider: GetDefaultProvider()}, nil
	case "service":
		return &ServiceResource{CanonicalID: id, ServiceName: r.Name, DesiredState: r.State, Enabled: r.Enabled}, nil
	case "container":
		return &ContainerResource{
			CanonicalID: id, Name: r.Name, Image: r.Image, State: r.State,
			Ports: r.Ports, Env: r.Env, Volumes: r.Volumes, Engine: GetContainerEngine(),
		}, nil
	case "symlink":
		return &SymlinkResource{CanonicalID: id, Path: r.Path, Target: r.Target}, nil
	case "git":
		return &GitResource{CanonicalID: id, URL: r.URL, Path: r.Path}, nil
	default:
		return nil, fmt.Errorf("bilinmeyen kaynak: %s", r.Type)
	}
}
