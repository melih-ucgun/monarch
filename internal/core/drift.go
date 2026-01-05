package core

import (
	"fmt"
	"strings"
	"sync"
)

// DriftStatus represents the sync state of a resource
type DriftStatus string

const (
	StatusSynced  DriftStatus = "Synced"
	StatusDrifted DriftStatus = "Drifted"
	StatusError   DriftStatus = "Error"
	StatusUnknown DriftStatus = "Unknown"
)

// DriftResult holds the result of a single resource check
type DriftResult struct {
	Type    string
	Name    string
	Desired string // Desired State (e.g. "present", "running")
	Status  DriftStatus
	Detail  string // Error message or drift detail
	Diff    string // Optional diff (if supported)
}

// CheckDrift performs a live audit of the given configuration against the system.
// It executes checks in parallel layers using the existing DAG logic if possible,
// or just a simple parallel loop since checks are usually read-only.
// However, to keep it simple and safe, we can just run checks in parallel (read-only usually safe).
func CheckDrift(items []ConfigItem, createFn ResourceCreator, ctx *SystemContext) ([]DriftResult, error) {
	var results []DriftResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrency to avoid file descriptor exhaustion or overloading
	sem := make(chan struct{}, 20)

	for _, item := range items {
		wg.Add(1)
		go func(it ConfigItem) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			resType := it.Type
			resName := it.Name

			// Fix empty names (logic copied from Engine)
			if resName == "" {
				if n, ok := it.Params["name"].(string); ok {
					resName = n
				}
			}

			// Render templates (important for paths etc)
			// We clone params to avoid race conditions if params are shared (they shouldn't be deep shared usually)
			// But for safety, rendering modifies maps.
			// Ideally we deep copy params. For MVP, we assume ConfigItem extraction does a fresh map.
			// Rendering...
			// Since we don't have deep copy util handy, we assume it's safe or risk it for Check.
			// Actually Engine.Run receives []ConfigItem, each is a struct copy but Map is reference.
			// Let's rely on standard map behavior.
			// We MUST clone the map to avoid editing the original config in memory?
			// CheckDrift shouldn't mutate the config passed to it.
			// Deep copy params to avoid concurrent map access
			params := deepCopyMap(it.Params)

			// Render
			if err := renderParams(params, ctx); err != nil {
				mu.Lock()
				results = append(results, DriftResult{
					Type:    resType,
					Name:    resName,
					Desired: it.State,
					Status:  StatusError,
					Detail:  fmt.Sprintf("Template error: %v", err),
				})
				mu.Unlock()
				return
			}

			// Create
			res, err := createFn(resType, resName, params, ctx)
			if err != nil {
				mu.Lock()
				results = append(results, DriftResult{
					Type:    resType,
					Name:    resName,
					Desired: it.State,
					Status:  StatusError,
					Detail:  fmt.Sprintf("Creation error: %v", err),
				})
				mu.Unlock()
				return
			}

			// Validate
			if err := res.Validate(ctx); err != nil {
				mu.Lock()
				results = append(results, DriftResult{
					Type:    resType,
					Name:    resName,
					Desired: it.State,
					Status:  StatusError,
					Detail:  fmt.Sprintf("Validation error: %v", err),
				})
				mu.Unlock()
				return
			}

			// Check
			// Resource.Check returns (needsAction bool, err error)
			// needsAction == true => Drifted
			drifted, err := res.Check(ctx)

			status := StatusSynced
			detail := ""
			diff := ""

			if err != nil {
				status = StatusError
				detail = err.Error()
			} else if drifted {
				status = StatusDrifted
				// Try to get detail/diff
				if differ, ok := res.(Differ); ok {
					if d, err := differ.Diff(ctx); err == nil && d != "" {
						diff = d
						// Summarize diff? Or just say "Content differs"
						firstLine := strings.Split(d, "\n")[0]
						detail = fmt.Sprintf("Changes detected (%s...)", firstLine)
					}
				}
				if detail == "" {
					detail = "Resource state does not match configuration"
				}
			}
			result := DriftResult{
				Type:    resType,
				Name:    resName,
				Desired: it.State,
				Status:  status,
				Detail:  detail,
				Diff:    diff,
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()

		}(item)
	}

	wg.Wait()
	return results, nil
}

// deepCopyMap creates a deep copy of a map to ensure thread safety
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = deepCopyMap(val)
		case []interface{}:
			newSlice := make([]interface{}, len(val))
			for i, item := range val {
				if subMap, ok := item.(map[string]interface{}); ok {
					newSlice[i] = deepCopyMap(subMap)
				} else {
					newSlice[i] = item
				}
			}
			result[k] = newSlice
		default:
			result[k] = v
		}
	}
	return result
}
