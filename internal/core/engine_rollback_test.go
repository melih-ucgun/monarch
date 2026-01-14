package core_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/melih-ucgun/veto/internal/core"
)

func TestEngine_Rollback_Sequential_PartialFailure(t *testing.T) {
	ctx := core.NewSystemContext(false, nil, nil)
	updater := &MockStateUpdater{}
	engine := core.NewEngine(ctx, updater)

	// Setup: 1 OK, 2 OK, 3 FAIL
	res1 := &MockResource{Name: "pkg1", Type: "pkg", ApplyResult: core.SuccessChange("installed")}
	res2 := &MockResource{Name: "pkg2", Type: "pkg", ApplyResult: core.SuccessChange("installed")}
	res3 := &MockResource{Name: "pkg3", Type: "pkg", ApplyErr: errors.New("connection failed")}

	createFn := func(t, n string, p map[string]interface{}, c *core.SystemContext) (core.Resource, error) {
		switch n {
		case "pkg1":
			return res1, nil
		case "pkg2":
			return res2, nil
		case "pkg3":
			return res3, nil
		}
		return nil, errors.New("unknown")
	}

	items := []core.ConfigItem{
		{Name: "pkg1", Type: "pkg"},
		{Name: "pkg2", Type: "pkg"},
		{Name: "pkg3", Type: "pkg"},
	}

	err := engine.Run(items, createFn)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verification
	// pkg3 failed, so it shouldn't be reverted (it wasn't applied successfully)
	if res3.RevertCalled {
		t.Error("Revert should NOT be called for failed resource pkg3")
	}

	// pkg2 and pkg1 should be reverted
	if !res2.RevertCalled {
		t.Error("pkg2 was not reverted")
	}
	if !res1.RevertCalled {
		t.Error("pkg1 was not reverted")
	}

	// Verify State Updates for pkg1 and pkg2 are 'reverted'
	revertedCount := 0
	for _, u := range updater.Updates {
		if u.Status == "reverted" {
			revertedCount++
		}
	}
	if revertedCount < 2 {
		t.Errorf("Expected at least 2 resources marked as reverted, got %d", revertedCount)
	}
}

func TestEngine_Rollback_Parallel_PartialFailure(t *testing.T) {
	ctx := core.NewSystemContext(false, nil, nil)
	// We don't strictly need updater for logic check, but good for completeness
	updater := &MockStateUpdater{}
	engine := core.NewEngine(ctx, updater)

	// Layer 1: base_pkg (Success)
	// Layer 2: app_pkg1 (Success), app_pkg2 (Fail)

	resBase := &MockResource{Name: "base_pkg", ApplyResult: core.SuccessChange("ok")}
	resApp1 := &MockResource{Name: "app_pkg1", ApplyResult: core.SuccessChange("ok")}
	resApp2 := &MockResource{Name: "app_pkg2", ApplyErr: errors.New("fail")}

	createFn := func(t, n string, p map[string]interface{}, c *core.SystemContext) (core.Resource, error) {
		switch n {
		case "base_pkg":
			return resBase, nil
		case "app_pkg1":
			return resApp1, nil
		case "app_pkg2":
			return resApp2, nil
		}
		return nil, errors.New("unknown")
	}

	// Use DependsOn to force layers
	items := []core.ConfigItem{
		{Name: "base_pkg", Type: "pkg"},                                  // Layer 1
		{Name: "app_pkg1", Type: "pkg", DependsOn: []string{"base_pkg"}}, // Layer 2
		{Name: "app_pkg2", Type: "pkg", DependsOn: []string{"base_pkg"}}, // Layer 2
	}

	err := engine.Run(items, createFn)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "encountered 1 errors") {
		t.Logf("Error message: %v", err)
	}

	// Verify Revert Logic
	// Layer 2 failure (app_pkg2) should trigger revert of app_pkg1 (same layer success)
	if !resApp1.RevertCalled {
		t.Error("app_pkg1 (same layer as failure) was not reverted")
	}

	// Then, it should propagate to revert Layer 1 (base_pkg)
	if !resBase.RevertCalled {
		t.Error("base_pkg (previous layer) was not reverted")
	}
}
