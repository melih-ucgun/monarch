package core

import (
	"fmt"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

// GenerateDiff generates a unified diff between current and desired content.
func GenerateDiff(name, current, desired string) string {
	if current == desired {
		return ""
	}

	edits := myers.ComputeEdits(span.URIFromPath(name), current, desired)
	return fmt.Sprint(gotextdiff.ToUnified(name+" (current)", name+" (desired)", current, edits))
}
