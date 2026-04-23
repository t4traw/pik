package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type DiffView struct {
	widget.BaseWidget
	scroll *container.Scroll
	body   *fyne.Container
}

func NewDiffView() *DiffView {
	d := &DiffView{
		body: container.NewVBox(),
	}
	d.scroll = container.NewScroll(d.body)
	d.ExtendBaseWidget(d)
	return d
}

func (d *DiffView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.scroll)
}

func (d *DiffView) SetEmpty(msg string) {
	d.body.RemoveAll()
	lbl := widget.NewLabel(msg)
	lbl.Alignment = fyne.TextAlignCenter
	d.body.Add(container.NewCenter(lbl))
	d.body.Refresh()
	d.scroll.ScrollToTop()
}

func (d *DiffView) SetDiff(raw string) {
	d.body.RemoveAll()
	if strings.TrimSpace(raw) == "" {
		d.SetEmpty("差分なし")
		return
	}
	lines := strings.Split(raw, "\n")
	// drop trailing empty line caused by final newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for _, ln := range lines {
		d.body.Add(renderDiffLine(ln))
	}
	d.body.Refresh()
	d.scroll.ScrollToTop()
}

func renderDiffLine(line string) fyne.CanvasObject {
	var bg color.Color = color.Transparent
	var fg color.Color = color.NRGBA{R: 0xe6, G: 0xe6, B: 0xe6, A: 0xff}

	switch {
	case strings.HasPrefix(line, "diff ") ||
		strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "--- ") ||
		strings.HasPrefix(line, "+++ ") ||
		strings.HasPrefix(line, "new file") ||
		strings.HasPrefix(line, "deleted file") ||
		strings.HasPrefix(line, "similarity") ||
		strings.HasPrefix(line, "rename "):
		bg = ColorHeader
		fg = color.NRGBA{R: 0x9c, G: 0xdc, B: 0xfe, A: 0xff}
	case strings.HasPrefix(line, "@@"):
		bg = ColorHunk
		fg = color.NRGBA{R: 0xc5, G: 0xa3, B: 0xff, A: 0xff}
	case strings.HasPrefix(line, "+"):
		bg = ColorAdd
		fg = color.NRGBA{R: 0xb5, G: 0xf2, B: 0xb5, A: 0xff}
	case strings.HasPrefix(line, "-"):
		bg = ColorDel
		fg = color.NRGBA{R: 0xf2, G: 0xb5, B: 0xb5, A: 0xff}
	}

	txt := canvas.NewText(line, fg)
	txt.TextStyle = fyne.TextStyle{Monospace: true}
	txt.TextSize = 12

	rect := canvas.NewRectangle(bg)
	// Pad the text horizontally by wrapping in a container with small padding
	padded := container.NewPadded(txt)
	return container.NewStack(rect, padded)
}
