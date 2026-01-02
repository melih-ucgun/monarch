package core

import (
	"bytes"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// GenerateDiff generates a unified diff between current and desired content.
// GenerateDiff generates a unified diff between current and desired content.
func GenerateDiff(name, current, desired string) string {
	dmp := diffmatchpatch.New()

	// Perform line-level diff
	a, b, c := dmp.DiffLinesToChars(current, desired)
	diffs := dmp.DiffMain(a, b, false)
	result := dmp.DiffCharsToLines(diffs, c)

	// No clean up, we want raw diffs for visualization or simple cleanup
	// dmp.DiffCleanupSemantic(result)

	// Create a unified-like output or just return the patches?
	// For "Visual Diff", we might want to return the Diff list or a formatted string.
	// Let's return a unified diff string which mimics git diff.
	// Actually, dmp.DiffPrettyText gives a very basic +/- view.
	// For true git style, we might need to iterate.

	// Let's use dmp.DiffPrettyText for now as a base,
	// but the CLI will likely need to parse this or we return a raw struct?
	// The interface returns string. A visual CLI needs ANSI codes.
	// We can return the string with ANSI codes if we want the core to handle coloring,
	// BUT core should be agnostic usually.
	// However, `pterm` is used in core...

	// Let's stick to the plan: Return a string representing the diff.
	// If we use diffmatchpatch, we can iterate and build a string.

	var buff bytes.Buffer
	for _, diff := range result {
		text := diff.Text
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			// Add green lines
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				buff.WriteString("+ " + line + "\n")
			}
		case diffmatchpatch.DiffDelete:
			// Add red lines
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				buff.WriteString("- " + line + "\n")
			}
		case diffmatchpatch.DiffEqual:
			// Context (optional, let's keep it minimal or full?)
			// Git shows context.
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				buff.WriteString("  " + line + "\n")
			}
		}
	}
	return buff.String()
}
