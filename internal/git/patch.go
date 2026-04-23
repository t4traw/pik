package git

import (
	"fmt"
	"strings"
)

// LineOp describes the role of a diff line in a hunk.
type LineOp string

const (
	OpContext LineOp = "context"
	OpAdd     LineOp = "add"
	OpRemove  LineOp = "remove"
)

// PatchLine is a single line of a hunk with a selection flag.
type PatchLine struct {
	Op       LineOp `json:"op"`
	Text     string `json:"text"`
	Selected bool   `json:"selected"`
}

// PatchHunk is a subset of a unified diff hunk the frontend has selected from.
type PatchHunk struct {
	OldStart int         `json:"oldStart"`
	NewStart int         `json:"newStart"`
	Lines    []PatchLine `json:"lines"`
}

// BuildSubPatch produces a unified diff patch that, when applied with
// `git apply --cached --recount`, stages exactly the selected lines of
// the input hunks.
//
// The rules for reducing a full hunk to a "selected-only" sub-patch are:
//   - Context lines always stay as context (' ').
//   - Add lines: if selected, remain '+'. If unselected, dropped entirely
//     (they still haven't entered the index).
//   - Remove lines: if selected, remain '-'. If unselected, become context
//     (' ') — that line is preserved in the index for now.
//
// Empty hunks (no real changes) are skipped. If the result has no hunks,
// returns an empty string — caller should skip the apply.
func BuildSubPatch(path string, hunks []PatchHunk) string {
	var outHunks []string
	for _, h := range hunks {
		rendered := renderSubHunk(h)
		if rendered == "" {
			continue
		}
		outHunks = append(outHunks, rendered)
	}
	if len(outHunks) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "diff --git a/%s b/%s\n", path, path)
	fmt.Fprintf(&b, "--- a/%s\n", path)
	fmt.Fprintf(&b, "+++ b/%s\n", path)
	for _, h := range outHunks {
		b.WriteString(h)
	}
	return b.String()
}

func renderSubHunk(h PatchHunk) string {
	var body strings.Builder
	hasChange := false
	oldLines := 0
	newLines := 0
	for _, l := range h.Lines {
		switch l.Op {
		case OpContext:
			body.WriteString(" " + l.Text + "\n")
			oldLines++
			newLines++
		case OpAdd:
			if l.Selected {
				body.WriteString("+" + l.Text + "\n")
				newLines++
				hasChange = true
			}
			// Unselected add: drop (not yet in index, not in patch).
		case OpRemove:
			if l.Selected {
				body.WriteString("-" + l.Text + "\n")
				oldLines++
				hasChange = true
			} else {
				// Unselected remove: treat as context so the line stays.
				body.WriteString(" " + l.Text + "\n")
				oldLines++
				newLines++
			}
		}
	}
	if !hasChange {
		return ""
	}
	oldStart := h.OldStart
	if oldStart < 1 {
		oldStart = 1
	}
	newStart := h.NewStart
	if newStart < 1 {
		newStart = 1
	}
	header := fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", oldStart, oldLines, newStart, newLines)
	return header + body.String()
}
