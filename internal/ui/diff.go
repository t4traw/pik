package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ViewMode int

const (
	ModeUnified ViewMode = iota
	ModeSplit
)

const (
	gutterNumW = 5 // chars for a line number
)

type DiffView struct {
	widget.BaseWidget
	grid     *widget.TextGrid
	empty    *widget.Label
	container *fyne.Container

	mode     ViewMode
	filename string
	files    []FileDiff
	emptyMsg string
}

func NewDiffView() *DiffView {
	d := &DiffView{}
	d.grid = widget.NewTextGrid()
	d.grid.Scroll = fyne.ScrollBoth
	d.empty = widget.NewLabel("")
	d.empty.Alignment = fyne.TextAlignCenter
	d.empty.Hide()
	d.container = container.NewStack(d.grid, container.NewCenter(d.empty))
	d.ExtendBaseWidget(d)
	return d
}

func (d *DiffView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.container)
}

func (d *DiffView) Mode() ViewMode { return d.mode }

func (d *DiffView) SetMode(m ViewMode) {
	if d.mode == m {
		return
	}
	d.mode = m
	d.render()
}

func (d *DiffView) SetEmpty(msg string) {
	d.emptyMsg = msg
	d.files = nil
	d.filename = ""
	d.render()
}

func (d *DiffView) SetDiff(raw, filename string) {
	d.filename = filename
	d.files = ParseUnifiedDiff(raw)
	d.emptyMsg = ""
	if len(d.files) == 0 {
		d.SetEmpty("差分なし")
		return
	}
	d.render()
}

// render rebuilds the grid rows from the current state.
func (d *DiffView) render() {
	if d.emptyMsg != "" {
		d.grid.Rows = nil
		d.grid.Refresh()
		d.empty.SetText(d.emptyMsg)
		d.empty.Show()
		d.grid.Hide()
		return
	}
	d.empty.Hide()
	d.grid.Show()

	var rows []widget.TextGridRow
	for i, f := range d.files {
		if i > 0 {
			rows = append(rows, widget.TextGridRow{})
		}
		rows = append(rows, d.fileHeaderRow(f))
		if f.Binary {
			rows = append(rows, plainRow("(バイナリファイル)", ColorDefaultFG, nil))
			continue
		}
		for _, h := range f.Hunks {
			rows = append(rows, d.hunkHeaderRow(h))
			switch d.mode {
			case ModeSplit:
				rows = append(rows, d.buildSplitRows(h)...)
			default:
				rows = append(rows, d.buildUnifiedRows(h)...)
			}
		}
	}
	d.grid.Rows = rows
	d.grid.ScrollToTop()
	d.grid.Refresh()
}

// ----- row builders ------------------------------------------------------

func (d *DiffView) fileHeaderRow(f FileDiff) widget.TextGridRow {
	title := f.NewPath
	if title == "" {
		title = f.OldPath
	}
	if title == "" && len(f.Preamble) > 0 {
		title = f.Preamble[0]
	}
	rowBG := ColorHeader
	st := &widget.CustomTextGridStyle{
		FGColor:   color.NRGBA{R: 0x9c, G: 0xdc, B: 0xfe, A: 0xff},
		BGColor:   rowBG,
		TextStyle: fyne.TextStyle{Bold: true, Monospace: true},
	}
	return widget.TextGridRow{
		Cells: runesWithStyle(" "+title, st),
		Style: &widget.CustomTextGridStyle{BGColor: rowBG},
	}
}

func (d *DiffView) hunkHeaderRow(h Hunk) widget.TextGridRow {
	rowBG := ColorHunk
	st := &widget.CustomTextGridStyle{
		FGColor: color.NRGBA{R: 0xc5, G: 0xa3, B: 0xff, A: 0xff},
		BGColor: rowBG,
	}
	// Indent so the @@ aligns roughly with the content column.
	indent := strings.Repeat(" ", unifiedGutterChars())
	return widget.TextGridRow{
		Cells: runesWithStyle(indent+h.Header, st),
		Style: &widget.CustomTextGridStyle{BGColor: rowBG},
	}
}

// buildUnifiedRows outputs lines for a hunk grouped by block:
// context lines pass through, then runs of consecutive removes/adds are emitted
// as "all removes, then all adds", with word-level emphasis on paired lines.
func (d *DiffView) buildUnifiedRows(h Hunk) []widget.TextGridRow {
	var rows []widget.TextGridRow
	i := 0
	for i < len(h.Lines) {
		l := h.Lines[i]
		if l.Type == LineContext {
			rows = append(rows, d.unifiedContentRow(l.OldLineNo, l.NewLineNo, nil, color.Transparent, l.Text, nil, nil))
			i++
			continue
		}
		if l.Type == LineNoNewline {
			i++
			continue
		}
		// Collect the run of removes.
		rStart := i
		for i < len(h.Lines) && h.Lines[i].Type == LineRemove {
			i++
		}
		removes := h.Lines[rStart:i]
		// Collect the following run of adds.
		aStart := i
		for i < len(h.Lines) && h.Lines[i].Type == LineAdd {
			i++
		}
		adds := h.Lines[aStart:i]

		pairN := len(removes)
		if len(adds) < pairN {
			pairN = len(adds)
		}
		// Precompute word diff for paired lines.
		leftSegs := make([][]Segment, pairN)
		rightSegs := make([][]Segment, pairN)
		for k := 0; k < pairN; k++ {
			leftSegs[k], rightSegs[k] = PairWordDiff(removes[k].Text, adds[k].Text)
		}

		// Emit all removes first.
		for k, r := range removes {
			var segs []Segment
			if k < pairN {
				segs = leftSegs[k]
			}
			rows = append(rows, d.unifiedContentRow(r.OldLineNo, 0,
				&bgSpec{line: ColorDel, stripe: ColorDelStripe, emph: ColorDelEmph},
				color.Transparent, r.Text, segs, nil))
		}
		// Then all adds.
		for k, a := range adds {
			var segs []Segment
			if k < pairN {
				segs = rightSegs[k]
			}
			rows = append(rows, d.unifiedContentRow(0, a.NewLineNo,
				&bgSpec{line: ColorAdd, stripe: ColorAddStripe, emph: ColorAddEmph},
				color.Transparent, a.Text, segs, nil))
		}
	}
	return rows
}

// buildSplitRows outputs rows in side-by-side layout.
func (d *DiffView) buildSplitRows(h Hunk) []widget.TextGridRow {
	// Determine left half width: max content length of old-side lines across
	// this hunk (so the separator aligns).
	maxL := 0
	for _, l := range h.Lines {
		if l.Type == LineAdd {
			continue
		}
		if utf8.RuneCountInString(l.Text) > maxL {
			maxL = utf8.RuneCountInString(l.Text)
		}
	}

	var rows []widget.TextGridRow
	pairs := pairForSplit(h.Lines)
	for _, p := range pairs {
		var leftCells, rightCells []widget.TextGridCell
		switch {
		case p.Context != nil:
			ctx := p.Context
			leftCells = d.halfCells(ctx.OldLineNo, nil, color.Transparent, ctx.Text, nil, nil, maxL)
			rightCells = d.halfCells(ctx.NewLineNo, nil, color.Transparent, ctx.Text, nil, nil, 0)
		case p.Remove != nil && p.Add != nil:
			ls, rs := PairWordDiff(p.Remove.Text, p.Add.Text)
			leftCells = d.halfCells(p.Remove.OldLineNo,
				&bgSpec{line: ColorDel, stripe: ColorDelStripe, emph: ColorDelEmph},
				color.Transparent, p.Remove.Text, ls, nil, maxL)
			rightCells = d.halfCells(p.Add.NewLineNo,
				&bgSpec{line: ColorAdd, stripe: ColorAddStripe, emph: ColorAddEmph},
				color.Transparent, p.Add.Text, rs, nil, 0)
		case p.Remove != nil:
			leftCells = d.halfCells(p.Remove.OldLineNo,
				&bgSpec{line: ColorDel, stripe: ColorDelStripe, emph: ColorDelEmph},
				color.Transparent, p.Remove.Text, nil, nil, maxL)
			rightCells = d.halfCells(0, nil, color.Transparent, "", nil, nil, 0)
		case p.Add != nil:
			leftCells = d.halfCells(0, nil, color.Transparent, "", nil, nil, maxL)
			rightCells = d.halfCells(p.Add.NewLineNo,
				&bgSpec{line: ColorAdd, stripe: ColorAddStripe, emph: ColorAddEmph},
				color.Transparent, p.Add.Text, nil, nil, 0)
		default:
			continue
		}
		sep := widget.TextGridCell{Rune: '│', Style: &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0x3c, G: 0x3c, B: 0x3c, A: 0xff}}}
		cells := make([]widget.TextGridCell, 0, len(leftCells)+1+len(rightCells))
		cells = append(cells, leftCells...)
		cells = append(cells, sep)
		cells = append(cells, rightCells...)
		rows = append(rows, widget.TextGridRow{Cells: cells})
	}
	return rows
}

// pairForSplit walks a hunk and pairs remove/add runs for split display.
func pairForSplit(lines []DiffLine) []LinePair {
	return PairHunkLines(lines)
}

// ----- cell builders ------------------------------------------------------

type bgSpec struct {
	line   color.Color // full-line bg (subtle)
	stripe color.Color // left-edge stripe bg
	emph   color.Color // word-diff emphasis bg
}

// unifiedContentRow builds a unified-view content row.
// bg==nil means "context" (no bg, transparent stripe).
func (d *DiffView) unifiedContentRow(oldNo, newNo int, bg *bgSpec, _ color.Color, text string, segs []Segment, _ *struct{}) widget.TextGridRow {
	var stripeBG, lineBG, emphBG color.Color
	if bg != nil {
		stripeBG = bg.stripe
		lineBG = bg.line
		emphBG = bg.emph
	} else {
		stripeBG = color.Transparent
		lineBG = color.Transparent
		emphBG = color.Transparent
	}

	cells := make([]widget.TextGridCell, 0, 16+len(text))
	cells = append(cells, stripeCell(stripeBG))
	cells = append(cells, gutterCells(padLeft(numOrBlank(oldNo), gutterNumW)+" "+padLeft(numOrBlank(newNo), gutterNumW)+" ")...)
	cells = append(cells, contentCells(d.filename, text, segs, emphBG)...)

	return widget.TextGridRow{
		Cells: cells,
		Style: rowBGStyle(lineBG),
	}
}

// halfCells produces one half (left or right) of a split-view row.
// padTo: if > 0, right-pad the content cells with spaces so the column aligns.
func (d *DiffView) halfCells(lineNo int, bg *bgSpec, _ color.Color, text string, segs []Segment, _ *struct{}, padTo int) []widget.TextGridCell {
	var stripeBG, emphBG color.Color
	if bg != nil {
		stripeBG = bg.stripe
		emphBG = bg.emph
	} else {
		stripeBG = color.Transparent
		emphBG = color.Transparent
	}

	cells := make([]widget.TextGridCell, 0, 8+len(text))
	cells = append(cells, stripeCell(stripeBG))
	cells = append(cells, gutterCells(padLeft(numOrBlank(lineNo), gutterNumW)+" ")...)
	content := contentCells(d.filename, text, segs, emphBG)
	cells = append(cells, content...)
	if padTo > 0 {
		runeCount := utf8.RuneCountInString(text)
		for i := runeCount; i < padTo; i++ {
			cells = append(cells, widget.TextGridCell{Rune: ' '})
		}
	}
	return cells
}

func stripeCell(bg color.Color) widget.TextGridCell {
	return widget.TextGridCell{Rune: ' ', Style: &widget.CustomTextGridStyle{BGColor: bg}}
}

func gutterCells(text string) []widget.TextGridCell {
	st := &widget.CustomTextGridStyle{FGColor: ColorGutterFG, BGColor: ColorGutterBG}
	cells := make([]widget.TextGridCell, 0, len(text))
	for _, r := range text {
		cells = append(cells, widget.TextGridCell{Rune: r, Style: st})
	}
	return cells
}

// contentCells renders the text portion with syntax highlighting and optional
// word-diff emphasis via per-cell background.
func contentCells(filename, text string, segs []Segment, emphBG color.Color) []widget.TextGridCell {
	if text == "" {
		return nil
	}
	lineToks := Highlight(filename, text)
	var tokens []Token
	if len(lineToks) > 0 {
		tokens = lineToks[0]
	}
	if len(tokens) == 0 || tokensLen(tokens) != len(text) {
		tokens = []Token{{Text: text, Color: ColorDefaultFG}}
	}
	mask := emphasisByteMask(segs, len(text))

	cells := make([]widget.TextGridCell, 0, utf8.RuneCountInString(text))
	pos := 0
	for _, tok := range tokens {
		if tok.Text == "" {
			continue
		}
		end := pos + len(tok.Text)
		for i := pos; i < end; {
			r, size := utf8.DecodeRuneInString(text[i:])
			em := i < len(mask) && mask[i]
			var st *widget.CustomTextGridStyle
			if em {
				st = &widget.CustomTextGridStyle{FGColor: tok.Color, BGColor: emphBG}
			} else {
				st = &widget.CustomTextGridStyle{FGColor: tok.Color}
			}
			cells = append(cells, widget.TextGridCell{Rune: r, Style: st})
			i += size
		}
		pos = end
	}
	return cells
}

func emphasisByteMask(segs []Segment, n int) []bool {
	if len(segs) == 0 {
		return nil
	}
	mask := make([]bool, n)
	pos := 0
	for _, s := range segs {
		end := pos + len(s.Text)
		if end > n {
			end = n
		}
		for i := pos; i < end; i++ {
			mask[i] = s.Emphasis
		}
		pos += len(s.Text)
	}
	return mask
}

func rowBGStyle(c color.Color) widget.TextGridStyle {
	if c == nil {
		return nil
	}
	if _, ok := c.(color.Alpha); ok {
		return nil
	}
	if t, ok := c.(color.NRGBA); ok && t.A == 0 {
		return nil
	}
	return &widget.CustomTextGridStyle{BGColor: c}
}

func plainRow(text string, fg color.Color, bg color.Color) widget.TextGridRow {
	st := &widget.CustomTextGridStyle{FGColor: fg, BGColor: bg}
	return widget.TextGridRow{Cells: runesWithStyle(text, st)}
}

func runesWithStyle(text string, st widget.TextGridStyle) []widget.TextGridCell {
	cells := make([]widget.TextGridCell, 0, len(text))
	for _, r := range text {
		cells = append(cells, widget.TextGridCell{Rune: r, Style: st})
	}
	return cells
}

// ----- helpers ------------------------------------------------------------

func numOrBlank(n int) string {
	if n <= 0 {
		return ""
	}
	return strconv.Itoa(n)
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

func tokensLen(toks []Token) int {
	n := 0
	for _, t := range toks {
		n += len(t.Text)
	}
	return n
}

func unifiedGutterChars() int { return 1 + gutterNumW + 1 + gutterNumW + 1 } // stripe + old + sp + new + sp

// avoid unused import warnings
var _ = fmt.Sprintf
