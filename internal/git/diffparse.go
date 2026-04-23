package git

import (
	"strconv"
	"strings"
)

// DiffLine is a single parsed line of a unified diff.
type DiffLine struct {
	Op        LineOp `json:"op"`        // "context" | "add" | "remove"
	Text      string `json:"text"`
	OldLineNo int    `json:"oldLineNo"` // 0 = N/A (add-only)
	NewLineNo int    `json:"newLineNo"` // 0 = N/A (remove-only)
}

// Hunk is a hunk in a unified diff.
type Hunk struct {
	OldStart int        `json:"oldStart"`
	NewStart int        `json:"newStart"`
	Header   string     `json:"header"`
	Lines    []DiffLine `json:"lines"`
}

// FileDiff is all hunks + metadata for one file in a unified diff.
type FileDiff struct {
	OldPath  string   `json:"oldPath"`
	NewPath  string   `json:"newPath"`
	Hunks    []Hunk   `json:"hunks"`
	Binary   bool     `json:"binary"`
	Preamble []string `json:"preamble"`
}

// ParseUnifiedDiff parses unified diff text (possibly covering multiple files)
// into a structured list of FileDiff.
func ParseUnifiedDiff(raw string) []FileDiff {
	var result []FileDiff
	if strings.TrimSpace(raw) == "" {
		return result
	}
	lines := strings.Split(raw, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var cur *FileDiff
	var curHunk *Hunk
	var oldLine, newLine int

	flushHunk := func() {
		if curHunk != nil && cur != nil {
			cur.Hunks = append(cur.Hunks, *curHunk)
			curHunk = nil
		}
	}
	flushFile := func() {
		flushHunk()
		if cur != nil {
			result = append(result, *cur)
			cur = nil
		}
	}

	for _, ln := range lines {
		switch {
		case strings.HasPrefix(ln, "diff --git "):
			flushFile()
			cur = &FileDiff{}
			cur.Preamble = append(cur.Preamble, ln)
			parts := strings.Fields(ln)
			if len(parts) >= 4 {
				cur.OldPath = strings.TrimPrefix(parts[2], "a/")
				cur.NewPath = strings.TrimPrefix(parts[3], "b/")
			}
		case strings.HasPrefix(ln, "diff "):
			flushFile()
			cur = &FileDiff{}
			cur.Preamble = append(cur.Preamble, ln)
		case cur != nil && strings.HasPrefix(ln, "--- "):
			cur.Preamble = append(cur.Preamble, ln)
			p := strings.TrimPrefix(ln, "--- ")
			if p != "/dev/null" {
				cur.OldPath = strings.TrimPrefix(p, "a/")
			}
		case cur != nil && strings.HasPrefix(ln, "+++ "):
			cur.Preamble = append(cur.Preamble, ln)
			p := strings.TrimPrefix(ln, "+++ ")
			if p != "/dev/null" {
				cur.NewPath = strings.TrimPrefix(p, "b/")
			}
		case cur != nil && strings.HasPrefix(ln, "Binary files "):
			cur.Binary = true
			cur.Preamble = append(cur.Preamble, ln)
		case cur != nil && strings.HasPrefix(ln, "@@"):
			flushHunk()
			oldS, newS := parseHunkHeader(ln)
			curHunk = &Hunk{OldStart: oldS, NewStart: newS, Header: ln}
			oldLine = oldS
			newLine = newS
		case curHunk != nil && len(ln) > 0 && ln[0] == '+':
			curHunk.Lines = append(curHunk.Lines, DiffLine{
				Op:        OpAdd,
				Text:      ln[1:],
				NewLineNo: newLine,
			})
			newLine++
		case curHunk != nil && len(ln) > 0 && ln[0] == '-':
			curHunk.Lines = append(curHunk.Lines, DiffLine{
				Op:        OpRemove,
				Text:      ln[1:],
				OldLineNo: oldLine,
			})
			oldLine++
		case curHunk != nil && len(ln) > 0 && ln[0] == ' ':
			curHunk.Lines = append(curHunk.Lines, DiffLine{
				Op:        OpContext,
				Text:      ln[1:],
				OldLineNo: oldLine,
				NewLineNo: newLine,
			})
			oldLine++
			newLine++
		case cur != nil:
			cur.Preamble = append(cur.Preamble, ln)
		}
	}
	flushFile()
	return result
}

func parseHunkHeader(h string) (oldStart, newStart int) {
	inner := strings.TrimPrefix(h, "@@")
	if i := strings.Index(inner, "@@"); i >= 0 {
		inner = inner[:i]
	}
	inner = strings.TrimSpace(inner)
	for _, p := range strings.Fields(inner) {
		if strings.HasPrefix(p, "-") {
			oldStart = firstIntOf(p[1:])
		} else if strings.HasPrefix(p, "+") {
			newStart = firstIntOf(p[1:])
		}
	}
	return
}

func firstIntOf(s string) int {
	comma := strings.IndexByte(s, ',')
	if comma >= 0 {
		s = s[:comma]
	}
	n, _ := strconv.Atoi(s)
	return n
}
