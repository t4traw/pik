package ui

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Segment struct {
	Text     string
	Emphasis bool
}

// PairWordDiff computes a character-level diff between two lines and returns
// segment lists for the old (left) and new (right) sides. Segments with
// Emphasis=true are the parts that actually differ between the lines.
func PairWordDiff(oldText, newText string) (left, right []Segment) {
	dmp := diffmatchpatch.New()
	// Use character mode directly — it produces reasonable word-level
	// boundaries when combined with Coalesce.
	diffs := dmp.DiffMain(oldText, newText, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	for _, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			left = append(left, Segment{Text: d.Text, Emphasis: false})
			right = append(right, Segment{Text: d.Text, Emphasis: false})
		case diffmatchpatch.DiffDelete:
			left = append(left, Segment{Text: d.Text, Emphasis: true})
		case diffmatchpatch.DiffInsert:
			right = append(right, Segment{Text: d.Text, Emphasis: true})
		}
	}
	return left, right
}

// LinePair references lines in a Hunk by index so callers can look up both
// the DiffLine itself and any per-line precomputed data (e.g. tokens).
// Exactly one of ContextIdx / (RemoveIdx, AddIdx) is set; -1 means absent.
type LinePair struct {
	RemoveIdx  int
	AddIdx     int
	ContextIdx int
}

// PairHunkLines walks a hunk and pairs consecutive remove/add runs so word
// diff can be applied on a row-by-row basis for split display.
func PairHunkLines(lines []DiffLine) []LinePair {
	var out []LinePair
	i := 0
	for i < len(lines) {
		l := lines[i]
		switch l.Type {
		case LineContext:
			out = append(out, LinePair{RemoveIdx: -1, AddIdx: -1, ContextIdx: i})
			i++
		case LineRemove:
			rStart := i
			for i < len(lines) && lines[i].Type == LineRemove {
				i++
			}
			aStart := i
			for i < len(lines) && lines[i].Type == LineAdd {
				i++
			}
			nRem := aStart - rStart
			nAdd := i - aStart
			pair := nRem
			if nAdd < pair {
				pair = nAdd
			}
			for k := 0; k < pair; k++ {
				out = append(out, LinePair{RemoveIdx: rStart + k, AddIdx: aStart + k, ContextIdx: -1})
			}
			for k := pair; k < nRem; k++ {
				out = append(out, LinePair{RemoveIdx: rStart + k, AddIdx: -1, ContextIdx: -1})
			}
			for k := pair; k < nAdd; k++ {
				out = append(out, LinePair{RemoveIdx: -1, AddIdx: aStart + k, ContextIdx: -1})
			}
		case LineAdd:
			out = append(out, LinePair{RemoveIdx: -1, AddIdx: i, ContextIdx: -1})
			i++
		default:
			// Skip "No newline at EOF" markers.
			i++
		}
	}
	return out
}
