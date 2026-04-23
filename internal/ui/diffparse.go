package ui

import (
	"strconv"
	"strings"
)

type LineType int

const (
	LineContext LineType = iota
	LineAdd
	LineRemove
	LineNoNewline
)

type DiffLine struct {
	Type      LineType
	Text      string // without leading +/-/space
	OldLineNo int    // 0 means N/A
	NewLineNo int    // 0 means N/A
}

type Hunk struct {
	OldStart int
	NewStart int
	Header   string
	Lines    []DiffLine
}

type FileDiff struct {
	OldPath  string
	NewPath  string
	Preamble []string
	Hunks    []Hunk
	Binary   bool
}

// ParseUnifiedDiff parses a git diff output (possibly with multiple files)
// into a list of FileDiff entries.
func ParseUnifiedDiff(raw string) []FileDiff {
	var result []FileDiff
	if strings.TrimSpace(raw) == "" {
		return result
	}
	lines := strings.Split(raw, "\n")
	// Strip trailing empty
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
			// parse paths: diff --git a/path b/path
			parts := strings.Fields(ln)
			if len(parts) >= 4 {
				cur.OldPath = strings.TrimPrefix(parts[2], "a/")
				cur.NewPath = strings.TrimPrefix(parts[3], "b/")
			}
		case strings.HasPrefix(ln, "diff "):
			// non-git diff (e.g. from --no-index)
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
			curHunk = &Hunk{
				OldStart: oldS,
				NewStart: newS,
				Header:   ln,
			}
			oldLine = oldS
			newLine = newS
		case curHunk != nil && len(ln) > 0 && ln[0] == '+':
			dl := DiffLine{Type: LineAdd, Text: ln[1:], NewLineNo: newLine}
			curHunk.Lines = append(curHunk.Lines, dl)
			newLine++
		case curHunk != nil && len(ln) > 0 && ln[0] == '-':
			dl := DiffLine{Type: LineRemove, Text: ln[1:], OldLineNo: oldLine}
			curHunk.Lines = append(curHunk.Lines, dl)
			oldLine++
		case curHunk != nil && len(ln) > 0 && ln[0] == ' ':
			dl := DiffLine{Type: LineContext, Text: ln[1:], OldLineNo: oldLine, NewLineNo: newLine}
			curHunk.Lines = append(curHunk.Lines, dl)
			oldLine++
			newLine++
		case curHunk != nil && strings.HasPrefix(ln, `\ `):
			// "\ No newline at end of file"
			curHunk.Lines = append(curHunk.Lines, DiffLine{Type: LineNoNewline, Text: ln[2:]})
		case cur != nil:
			// preamble extras (index..., new file mode, etc.)
			cur.Preamble = append(cur.Preamble, ln)
		}
	}
	flushFile()
	return result
}

// parseHunkHeader parses "@@ -oldStart,oldLines +newStart,newLines @@ ...".
func parseHunkHeader(h string) (oldStart, newStart int) {
	// Strip leading "@@ " and trailing " @@..."
	inner := strings.TrimPrefix(h, "@@")
	if i := strings.Index(inner, "@@"); i >= 0 {
		inner = inner[:i]
	}
	inner = strings.TrimSpace(inner)
	// e.g. "-10,3 +12,4"
	parts := strings.Fields(inner)
	for _, p := range parts {
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
