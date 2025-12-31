package bundle

import (
	"fmt"
	"strings"

	"github.com/melih-ucgun/monarch/internal/core"
)

// ResourceCreator is a function type to break dependency cycle.
// It matches the signature of resource.CreateResourceWithParams (mostly).
type ResourceCreator func(resType, name string, params map[string]interface{}, ctx *core.SystemContext) (core.Resource, error)

type BundleAdapter struct {
	core.BaseResource
	Items   []map[string]interface{}
	Creator ResourceCreator
}

func NewBundleAdapter(name string, params map[string]interface{}, creator ResourceCreator) *BundleAdapter {
	items := []map[string]interface{}{}
	if resList, ok := params["resources"].([]interface{}); ok {
		for _, rawItem := range resList {
			if itemMap, ok := rawItem.(map[string]interface{}); ok {
				items = append(items, itemMap)
			}
		}
	} else if resList, ok := params["resources"].([]map[string]interface{}); ok {
		items = resList
	}

	return &BundleAdapter{
		BaseResource: core.BaseResource{Name: name, Type: "bundle"},
		Items:        items,
		Creator:      creator,
	}
}

func (r *BundleAdapter) Validate() error {
	if len(r.Items) == 0 {
		return fmt.Errorf("bundle '%s' has no resources", r.Name)
	}
	return nil
}

func (r *BundleAdapter) Check(ctx *core.SystemContext) (bool, error) {
	// Bundle always needs "Apply" to run its children checks/applies.
	// We can't easily check aggregate state without running checks on all children.
	// For simplicity, we say "true" (needs action) so Apply is called,
	// and Apply manages the children idempotent logic.
	return true, nil
}

func (r *BundleAdapter) Apply(ctx *core.SystemContext) (core.Result, error) {
	changes := []string{}

	// Iterate and apply children sequentially
	for i, item := range r.Items {
		// Extract core fields
		// YAML keys might be flexible, but we assume standard structure
		typeVal, _ := item["type"].(string)
		nameVal, _ := item["name"].(string)

		// "when" logic is handled by Engine usually.
		// Here, we don't have the engine's "When" evaluator available easily
		// unless we duplicate that logic or pass an evaluator.
		// For now, we skip "when" inside bundles or assume basic execution.
		// NOTE: Params to child is 'item' itself.

		if typeVal == "" || nameVal == "" {
			return core.Failure(nil, fmt.Sprintf("Bundle item #%d missing type or name", i)), fmt.Errorf("invalid config")
		}

		// Create Resource
		child, err := r.Creator(typeVal, nameVal, item, ctx)
		if err != nil {
			return core.Failure(err, fmt.Sprintf("Failed to create resource %s: %v", nameVal, err)), err
		}

		// Validate
		if err := child.Validate(); err != nil {
			return core.Failure(err, fmt.Sprintf("Invalid resource %s: %v", nameVal, err)), err
		}

		// Apply Child
		// Logic similar to Engine's execution: Check then Apply
		output, err := child.Apply(ctx)
		if err != nil {
			errMsg := fmt.Sprintf("Child %s failed: %v", nameVal, err)
			return core.Failure(err, errMsg), err
		}

		if output.Changed {
			changes = append(changes, fmt.Sprintf("[%s] %s", nameVal, output.Message))
		}
		// If not changed, we silently continue, or maybe debug log?
	}

	if len(changes) > 0 {
		return core.SuccessChange(fmt.Sprintf("Bundle applied: %s", strings.Join(changes, "; "))), nil
	}

	return core.SuccessNoChange("All bundle resources in desired state"), nil
}

func (r *BundleAdapter) Revert(ctx *core.SystemContext) error {
	// Revert in reverse order
	for i := len(r.Items) - 1; i >= 0; i-- {
		item := r.Items[i]
		typeVal, _ := item["type"].(string)
		nameVal, _ := item["name"].(string)

		child, err := r.Creator(typeVal, nameVal, item, ctx)
		if err != nil {
			continue // Skip invalid
		}

		// We cast to Reverter interface if possible
		if reverter, ok := child.(interface {
			Revert(*core.SystemContext) error
		}); ok {
			reverter.Revert(ctx)
		}
	}
	return nil
}
