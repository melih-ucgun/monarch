package core

import (
	"fmt"
	"sync"
)

// ResourceFactory is a function that creates a new resource instance.
type ResourceFactory func(name string, params map[string]interface{}, ctx *SystemContext) (Resource, error)

var (
	resourceRegistry = make(map[string]ResourceFactory)
	registryMu       sync.RWMutex
)

// RegisterResource registers a resource factory for a given type name.
func RegisterResource(typeName string, factory ResourceFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	resourceRegistry[typeName] = factory
}

// CreateResource instantiates a resource of the given type.
func CreateResource(typeName string, name string, params map[string]interface{}, ctx *SystemContext) (Resource, error) {
	registryMu.RLock()
	factory, ok := resourceRegistry[typeName]
	registryMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown resource type: %s", typeName)
	}

	return factory(name, params, ctx)
}

// GetRegisteredTypes returns a list of all registered resource types.
func GetRegisteredTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	types := make([]string, 0, len(resourceRegistry))
	for t := range resourceRegistry {
		types = append(types, t)
	}
	return types
}
