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

	// If raw diff bytes exceed this, syntax highlighting is skipped for
	// speed. Still shows colors for +/- lines and word-level emphasis.
	highlightByteLimit = 40 * 1024
)

type DiffView struct {
	widget.BaseWidget
	grid      *widget.TextGrid
	empty     *widget.Label
	container *fyne.Container

	mode     ViewMode
	filename string
	files    []FileDiff
	emptyMsg string
	skipHL   bool // skip chroma for this diff (large file)
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
	d.skipHL = false
	d.render()
}

func (d *DiffView) SetDiff(raw, filename string) {
	d.filename = filename
	d.skipHL = len(raw) > highlightByteLimit
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
	return widget.TextGridRow{
		Cells: runesWithStyle(" "+title, fileHeaderStyle),
		Style: fileHeaderRowStyle,
	}
}

func (d *DiffView) hunkHeaderRow(h Hunk) widget.TextGridRow {
	indent := strings.Repeat(" ", unifiedGutterChars())
	return widget.TextGridRow{
		Cells: runesWithStyle(indent+h.Header, hunkHeaderStyle),
		Style: hunkHeaderRowStyle,
	}
}

// buildUnifiedRows outputs lines for a hunk grouped by block:
// context lines pass through, then runs of consecutive removes/adds are emitted
// as "all removes, then all adds", with word-level emphasis on paired lines.
func (d *DiffView) buildUnifiedRows(h Hunk) []widget.TextGridRow {
	tokens := d.tokenizeHunk(h)
	var rows []widget.TextGridRow
	i := 0
	for i < len(h.Lines) {
		l := h.Lines[i]
		if l.Type == LineContext {
			rows = append(rows, d.unifiedContentRow(l.OldLineNo, l.NewLineNo, ctxBGSpec, l.Text, tokens[i], nil))
			i++
			continue
		}
		if l.Type == LineNoNewline {
			i++
			continue
		}
		rStart := i
		for i < len(h.Lines) && h.Lines[i].Type == LineRemove {
			i++
		}
		rEnd := i
		aStart := i
		for i < len(h.Lines) && h.Lines[i].Type == LineAdd {
			i++
		}
		aEnd := i

		nRem := rEnd - rStart
		nAdd := aEnd - aStart
		pairN := nRem
		if nAdd < pairN {
			pairN = nAdd
		}
		leftSegs := make([][]Segment, pairN)
		rightSegs := make([][]Segment, pairN)
		for k := 0; k < pairN; k++ {
			leftSegs[k], rightSegs[k] = PairWordDiff(h.Lines[rStart+k].Text, h.Lines[aStart+k].Text)
		}

		for k := 0; k < nRem; k++ {
			r := h.Lines[rStart+k]
			var segs []Segment
			if k < pairN {
				segs = leftSegs[k]
			}
			rows = append(rows, d.unifiedContentRow(r.OldLineNo, 0, delBGSpec, r.Text, tokens[rStart+k], segs))
		}
		for k := 0; k < nAdd; k++ {
			a := h.Lines[aStart+k]
			var segs []Segment
			if k < pairN {
				segs = rightSegs[k]
			}
			rows = append(rows, d.unifiedContentRow(0, a.NewLineNo, addBGSpec, a.Text, tokens[aStart+k], segs))
		}
	}
	return rows
}

// buildSplitRows outputs rows in side-by-side layout.
func (d *DiffView) buildSplitRows(h Hunk) []widget.TextGridRow {
	tokens := d.tokenizeHunk(h)

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
	pairs := PairHunkLines(h.Lines)
	for _, p := range pairs {
		var leftCells, rightCells []widget.TextGridCell
		switch {
		case p.ContextIdx >= 0:
			ctx := h.Lines[p.ContextIdx]
			leftCells = d.halfCells(ctx.OldLineNo, ctxBGSpec, ctx.Text, tokens[p.ContextIdx], nil, maxL)
			rightCells = d.halfCells(ctx.NewLineNo, ctxBGSpec, ctx.Text, tokens[p.ContextIdx], nil, 0)
		case p.RemoveIdx >= 0 && p.AddIdx >= 0:
			rem := h.Lines[p.RemoveIdx]
			add := h.Lines[p.AddIdx]
			ls, rs := PairWordDiff(rem.Text, add.Text)
			leftCells = d.halfCells(rem.OldLineNo, delBGSpec, rem.Text, tokens[p.RemoveIdx], ls, maxL)
			rightCells = d.halfCells(add.NewLineNo, addBGSpec, add.Text, tokens[p.AddIdx], rs, 0)
		case p.RemoveIdx >= 0:
			rem := h.Lines[p.RemoveIdx]
			leftCells = d.halfCells(rem.OldLineNo, delBGSpec, rem.Text, tokens[p.RemoveIdx], nil, maxL)
			rightCells = d.halfCells(0, ctxBGSpec, "", nil, nil, 0)
		case p.AddIdx >= 0:
			add := h.Lines[p.AddIdx]
			leftCells = d.halfCells(0, ctxBGSpec, "", nil, nil, maxL)
			rightCells = d.halfCells(add.NewLineNo, addBGSpec, add.Text, tokens[p.AddIdx], nil, 0)
		default:
			continue
		}
		cells := make([]widget.TextGridCell, 0, len(leftCells)+1+len(rightCells))
		cells = append(cells, leftCells...)
		cells = append(cells, splitSepCell)
		cells = append(cells, rightCells...)
		rows = append(rows, widget.TextGridRow{Cells: cells})
	}
	return rows
}

// tokenizeHunk batch-runs chroma across a hunk's old and new content once each.
// For huge diffs (d.skipHL), returns empty tokens so callers fall back to plain.
func (d *DiffView) tokenizeHunk(h Hunk) [][]Token {
	out := make([][]Token, len(h.Lines))
	if d.skipHL {
		return out
	}
	oldIdx := make([]int, len(h.Lines))
	newIdx := make([]int, len(h.Lines))
	for i := range oldIdx {
		oldIdx[i] = -1
		newIdx[i] = -1
	}
	var oldLines, newLines []string
	for i, l := range h.Lines {
		switch l.Type {
		case LineContext:
			oldIdx[i] = len(oldLines)
			newIdx[i] = len(newLines)
			oldLines = append(oldLines, l.Text)
			newLines = append(newLines, l.Text)
		case LineRemove:
			oldIdx[i] = len(oldLines)
			oldLines = append(oldLines, l.Text)
		case LineAdd:
			newIdx[i] = len(newLines)
			newLines = append(newLines, l.Text)
		}
	}
	oldTokens := Highlight(d.filename, strings.Join(oldLines, "\n"))
	newTokens := Highlight(d.filename, strings.Join(newLines, "\n"))

	for i, l := range h.Lines {
		switch l.Type {
		case LineContext, LineRemove:
			if oldIdx[i] >= 0 && oldIdx[i] < len(oldTokens) {
				out[i] = oldTokens[oldIdx[i]]
			}
		case LineAdd:
			if newIdx[i] >= 0 && newIdx[i] < len(newTokens) {
				out[i] = newTokens[newIdx[i]]
			}
		}
	}
	return out
}

// ----- cell builders ------------------------------------------------------

type bgSpec struct {
	line         color.Color
	stripeStyle  *widget.CustomTextGridStyle
	emph         color.Color
	rowStyle     widget.TextGridStyle
}

var (
	ctxBGSpec = &bgSpec{line: nil, stripeStyle: stripeTransStyle, emph: nil, rowStyle: nil}
	addBGSpec = &bgSpec{line: ColorAdd, stripeStyle: stripeAddStyle, emph: ColorAddEmph, rowStyle: &widget.CustomTextGridStyle{BGColor: ColorAdd}}
	delBGSpec = &bgSpec{line: ColorDel, stripeStyle: stripeDelStyle, emph: ColorDelEmph, rowStyle: &widget.CustomTextGridStyle{BGColor: ColorDel}}
)

// Shared style pointers (avoid per-cell allocation).
var (
	gutterStyle      = &widget.CustomTextGridStyle{FGColor: ColorGutterFG, BGColor: ColorGutterBG}
	stripeAddStyle   = &widget.CustomTextGridStyle{BGColor: ColorAddStripe}
	stripeDelStyle   = &widget.CustomTextGridStyle{BGColor: ColorDelStripe}
	stripeTransStyle = &widget.CustomTextGridStyle{BGColor: color.Transparent}

	fileHeaderStyle    = &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0x9c, G: 0xdc, B: 0xfe, A: 0xff}, BGColor: ColorHeader, TextStyle: fyne.TextStyle{Bold: true, Monospace: true}}
	fileHeaderRowStyle = &widget.CustomTextGridStyle{BGColor: ColorHeader}
	hunkHeaderStyle    = &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0xc5, G: 0xa3, B: 0xff, A: 0xff}, BGColor: ColorHunk}
	hunkHeaderRowStyle = &widget.CustomTextGridStyle{BGColor: ColorHunk}

	splitSepCell = widget.TextGridCell{Rune: '│', Style: &widget.CustomTextGridStyle{FGColor: color.NRGBA{R: 0x3c, G: 0x3c, B: 0x3c, A: 0xff}}}

	// Per-fgcolor style caches — populated on demand.
	plainStyleCache    = map[fgColorKey]*widget.CustomTextGridStyle{}
	emphAddStyleCache  = map[fgColorKey]*widget.CustomTextGridStyle{}
	emphDelStyleCache  = map[fgColorKey]*widget.CustomTextGridStyle{}
	emphCtxStyleCache  = map[fgColorKey]*widget.CustomTextGridStyle{}
)

type fgColorKey struct {
	R, G, B, A uint8
}

func fgKey(c color.Color) fgColorKey {
	if n, ok := c.(color.NRGBA); ok {
		return fgColorKey{n.R, n.G, n.B, n.A}
	}
	r, g, b, a := c.RGBA()
	return fgColorKey{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func plainStyle(fg color.Color) *widget.CustomTextGridStyle {
	k := fgKey(fg)
	s, ok := plainStyleCache[k]
	if !ok {
		s = &widget.CustomTextGridStyle{FGColor: fg}
		plainStyleCache[k] = s
	}
	return s
}

func emphStyle(fg color.Color, emphBG color.Color) *widget.CustomTextGridStyle {
	k := fgKey(fg)
	var cache map[fgColorKey]*widget.CustomTextGridStyle
	switch emphBG {
	case ColorAddEmph:
		cache = emphAddStyleCache
	case ColorDelEmph:
		cache = emphDelStyleCache
	default:
		cache = emphCtxStyleCache
	}
	s, ok := cache[k]
	if !ok {
		s = &widget.CustomTextGridStyle{FGColor: fg, BGColor: emphBG}
		cache[k] = s
	}
	return s
}

// unifiedContentRow builds a unified-view content row.
func (d *DiffView) unifiedContentRow(oldNo, newNo int, bg *bgSpec, text string, tokens []Token, segs []Segment) widget.TextGridRow {
	cells := make([]widget.TextGridCell, 0, 16+len(text))
	cells = append(cells, widget.TextGridCell{Rune: ' ', Style: bg.stripeStyle})
	cells = append(cells, gutterCells(padLeft(numOrBlank(oldNo), gutterNumW)+" "+padLeft(numOrBlank(newNo), gutterNumW)+" ")...)
	cells = append(cells, contentCells(d, text, tokens, segs, bg.emph)...)

	return widget.TextGridRow{Cells: cells, Style: bg.rowStyle}
}

// halfCells produces one half of a split-view row.
func (d *DiffView) halfCells(lineNo int, bg *bgSpec, text string, tokens []Token, segs []Segment, padTo int) []widget.TextGridCell {
	cells := make([]widget.TextGridCell, 0, 8+len(text))
	cells = append(cells, widget.TextGridCell{Rune: ' ', Style: bg.stripeStyle})
	cells = append(cells, gutterCells(padLeft(numOrBlank(lineNo), gutterNumW)+" ")...)
	content := contentCells(d, text, tokens, segs, bg.emph)
	cells = append(cells, content...)
	if padTo > 0 {
		runeCount := utf8.RuneCountInString(text)
		for i := runeCount; i < padTo; i++ {
			cells = append(cells, widget.TextGridCell{Rune: ' '})
		}
	}
	return cells
}

func gutterCells(text string) []widget.TextGridCell {
	cells := make([]widget.TextGridCell, 0, len(text))
	for _, r := range text {
		cells = append(cells, widget.TextGridCell{Rune: r, Style: gutterStyle})
	}
	return cells
}

// contentCells renders the text portion with syntax highlighting and optional
// word-diff emphasis via per-cell background.
func contentCells(d *DiffView, text string, tokens []Token, segs []Segment, emphBG color.Color) []widget.TextGridCell {
	if text == "" {
		return nil
	}
	// Skip syntax highlighting for huge diffs.
	if d.skipHL {
		tokens = []Token{{Text: text, Color: ColorDefaultFG}}
	} else if len(tokens) == 0 || tokensLen(tokens) != len(text) {
		lineToks := Highlight(d.filename, text)
		tokens = nil
		if len(lineToks) > 0 {
			tokens = lineToks[0]
		}
		if len(tokens) == 0 || tokensLen(tokens) != len(text) {
			tokens = []Token{{Text: text, Color: ColorDefaultFG}}
		}
	}
	mask := emphasisByteMask(segs, len(text))

	cells := make([]widget.TextGridCell, 0, utf8.RuneCountInString(text))
	pos := 0
	for _, tok := range tokens {
		if tok.Text == "" {
			continue
		}
		end := pos + len(tok.Text)
		pStyle := plainStyle(tok.Color)
		var eStyle *widget.CustomTextGridStyle
		if emphBG != nil {
			eStyle = emphStyle(tok.Color, emphBG)
		}
		for i := pos; i < end; {
			r, size := utf8.DecodeRuneInString(text[i:])
			em := i < len(mask) && mask[i]
			var st *widget.CustomTextGridStyle
			if em && eStyle != nil {
				st = eStyle
			} else {
				st = pStyle
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

func unifiedGutterChars() int { return 1 + gutterNumW + 1 + gutterNumW + 1 }

var _ = fmt.Sprintf
