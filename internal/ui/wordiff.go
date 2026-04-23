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

// PairHunkLines walks a hunk and pairs consecutive remove/add runs so word
// diff can be applied. Returns a slice where each entry is either a single
// line (unpaired) or a pair of lines (remove + add, both with the same index).
type LinePair struct {
	Remove *DiffLine // nil if add-only
	Add    *DiffLine // nil if remove-only
	// For context lines, both are nil and Context holds the shared line.
	Context *DiffLine
}

func PairHunkLines(lines []DiffLine) []LinePair {
	var out []LinePair
	i := 0
	for i < len(lines) {
		l := lines[i]
		switch l.Type {
		case LineContext:
			ll := l
			out = append(out, LinePair{Context: &ll})
			i++
		case LineRemove:
			// Collect the run of removes, then the run of adds that follows.
			rStart := i
			for i < len(lines) && lines[i].Type == LineRemove {
				i++
			}
			removes := lines[rStart:i]
			aStart := i
			for i < len(lines) && lines[i].Type == LineAdd {
				i++
			}
			adds := lines[aStart:i]
			pairUp := minInt(len(removes), len(adds))
			for k := 0; k < pairUp; k++ {
				rr := removes[k]
				aa := adds[k]
				out = append(out, LinePair{Remove: &rr, Add: &aa})
			}
			for k := pairUp; k < len(removes); k++ {
				rr := removes[k]
				out = append(out, LinePair{Remove: &rr})
			}
			for k := pairUp; k < len(adds); k++ {
				aa := adds[k]
				out = append(out, LinePair{Add: &aa})
			}
		case LineAdd:
			// Add without a preceding remove — unpaired.
			ll := l
			out = append(out, LinePair{Add: &ll})
			i++
		default:
			// Skip "No newline at EOF" markers from pairing.
			i++
		}
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
