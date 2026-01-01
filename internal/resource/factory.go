package resource

import (
	"github.com/melih-ucgun/veto/internal/core"
)

// CreateResourceWithParams uses the central registry to instantiate resources.
func CreateResourceWithParams(resType string, name string, params map[string]interface{}, ctx *core.SystemContext) (core.Resource, error) {
	if params == nil {
		params = make(map[string]interface{})
	}

	// For backward compatibility or convenience, ensure state is set if possible
	if _, ok := params["state"]; !ok {
		params["state"] = "present"
	}

	return core.CreateResource(resType, name, params, ctx)
}
