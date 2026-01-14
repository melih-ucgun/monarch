package core_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/melih-ucgun/veto/internal/core"
)

func TestEngine_RunSequential_Rollback(t *testing.T) {
	ctx := core.NewSystemContext(false, nil, nil)
	// Mock State Updater
	updater := &MockStateUpdater{}

	t.Run("Sequential failure triggers rollback", func(t *testing.T) {
		engine := core.NewEngine(ctx, updater)

		// Resource A: Success
		resA := &MockResource{Name: "resA", Type: "test", ApplyResult: core.SuccessChange("ok")}
		// Resource B: Success
		resB := &MockResource{Name: "resB", Type: "test", ApplyResult: core.SuccessChange("ok")}
		// Resource C: Fail
		resC := &MockResource{Name: "resC", Type: "test", ApplyErr: errors.New("fail")}

		createFn := func(t, n string, p map[string]interface{}, c *core.SystemContext) (core.Resource, error) {
			switch n {
			case "resA":
				return resA, nil
			case "resB":
				return resB, nil
			case "resC":
				return resC, nil
			}
			return nil, errors.New("unknown")
		}

		// Items without dependencies -> Sequential Mode
		items := []core.ConfigItem{{Name: "resA"}, {Name: "resB"}, {Name: "resC"}}
		err := engine.Run(items, createFn)

		// 1. Check Error
		if err == nil {
			t.Error("Expected error, got nil")
		} else if !strings.Contains(err.Error(), "encountered 1 errors") {
			t.Errorf("Unexpected error message: %v", err)
		}

		// 2. Verify Rollback was called on previous resources (LIFO order ideally, but check existence first)
		if !resB.RevertCalled {
			t.Error("Rollback not triggered for resB")
		}
		if !resA.RevertCalled {
			t.Error("Rollback not triggered for resA")
		}

		// 3. Verify State Updates contain 'reverted' status
		foundReverted := false
		for _, u := range updater.Updates {
			if u.Status == "reverted" {
				foundReverted = true
				break
			}
		}
		if !foundReverted {
			t.Error("State not updated to 'reverted'")
		}
	})
}
