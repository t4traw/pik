package ui

import (
	"fmt"
	"image/color"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/t4traw/pik/internal/git"
)

type FileList struct {
	widget.BaseWidget
	body   *fyne.Container
	scroll *container.Scroll

	OnSelect   func(f git.FileStatus, staged bool)
	OnStage    func(path string)
	OnUnstage  func(path string)
	OnDiscard  func(path string, untracked bool)
	OnStageAll func()
	OnResetAll func()

	selectedPath   string
	selectedStaged bool

	// Per-row selection-bg refs keyed by (path, staged) so we can update
	// highlight in-place without rebuilding the whole list.
	rowBG map[rowKey]*canvas.Rectangle
}

type rowKey struct {
	path   string
	staged bool
}

var (
	selectedBG = color.NRGBA{R: 0x09, G: 0x4f, B: 0x82, A: 0xff}
	clearBG    = color.Transparent
)

func NewFileList() *FileList {
	fl := &FileList{body: container.NewVBox(), rowBG: map[rowKey]*canvas.Rectangle{}}
	fl.scroll = container.NewScroll(fl.body)
	fl.ExtendBaseWidget(fl)
	return fl
}

func (fl *FileList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(fl.scroll)
}

func (fl *FileList) Selected() (string, bool) {
	return fl.selectedPath, fl.selectedStaged
}

func (fl *FileList) SetFiles(files []git.FileStatus) {
	fl.body.RemoveAll()
	fl.rowBG = map[rowKey]*canvas.Rectangle{}

	var staged, unstaged []git.FileStatus
	for _, f := range files {
		if f.Staged {
			staged = append(staged, f)
		}
		if f.Unstaged || f.Untracked {
			unstaged = append(unstaged, f)
		}
	}

	// --- STAGED ---
	stagedHdr := fl.sectionHeader(
		fmt.Sprintf("STAGED CHANGES (%d)", len(staged)),
		"すべてアンステージ",
		theme.ContentUndoIcon(),
		func() {
			if fl.OnResetAll != nil {
				fl.OnResetAll()
			}
		},
		len(staged) > 0,
	)
	fl.body.Add(stagedHdr)
	for _, f := range staged {
		fl.body.Add(fl.fileRow(f, true))
	}

	// --- CHANGES ---
	changesHdr := fl.sectionHeader(
		fmt.Sprintf("CHANGES (%d)", len(unstaged)),
		"すべてステージ",
		theme.ContentAddIcon(),
		func() {
			if fl.OnStageAll != nil {
				fl.OnStageAll()
			}
		},
		len(unstaged) > 0,
	)
	fl.body.Add(changesHdr)
	for _, f := range unstaged {
		fl.body.Add(fl.fileRow(f, false))
	}

	if len(staged) == 0 && len(unstaged) == 0 {
		lbl := widget.NewLabel("変更なし")
		lbl.Alignment = fyne.TextAlignCenter
		fl.body.Add(container.NewPadded(lbl))
	}

	fl.body.Refresh()
}

// selectRow updates the selection highlight in place without rebuilding rows.
func (fl *FileList) selectRow(path string, staged bool) {
	if fl.selectedPath == path && fl.selectedStaged == staged {
		return
	}
	if r, ok := fl.rowBG[rowKey{fl.selectedPath, fl.selectedStaged}]; ok {
		r.FillColor = clearBG
		r.Refresh()
	}
	fl.selectedPath = path
	fl.selectedStaged = staged
	if r, ok := fl.rowBG[rowKey{path, staged}]; ok {
		r.FillColor = selectedBG
		r.Refresh()
	}
}

func (fl *FileList) sectionHeader(title, tooltip string, icon fyne.Resource, onTap func(), enabled bool) fyne.CanvasObject {
	lbl := canvas.NewText(title, color.NRGBA{R: 0xcc, G: 0xcc, B: 0xcc, A: 0xff})
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	lbl.TextSize = 11

	bg := canvas.NewRectangle(color.NRGBA{R: 0x2a, G: 0x2a, B: 0x2a, A: 0xff})

	var row fyne.CanvasObject
	if onTap != nil && enabled {
		btn := widget.NewButtonWithIcon("", icon, onTap)
		btn.Importance = widget.LowImportance
		row = container.NewBorder(nil, nil, container.NewPadded(lbl), btn, nil)
	} else {
		row = container.NewBorder(nil, nil, container.NewPadded(lbl), nil, nil)
	}
	return container.NewStack(bg, row)
}

func (fl *FileList) fileRow(f git.FileStatus, staged bool) fyne.CanvasObject {
	letter, col := statusBadge(f, staged)

	badge := canvas.NewText(string(letter), col)
	badge.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	badge.TextSize = 12

	base := filepath.Base(f.Path)
	dir := filepath.Dir(f.Path)
	name := canvas.NewText(base, color.NRGBA{R: 0xe6, G: 0xe6, B: 0xe6, A: 0xff})
	name.TextSize = 13

	var right fyne.CanvasObject
	if dir != "." && dir != "" {
		d := canvas.NewText(dir, color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xff})
		d.TextSize = 11
		right = container.NewHBox(name, d)
	} else {
		right = name
	}

	var actions []fyne.CanvasObject
	if staged {
		u := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
			if fl.OnUnstage != nil {
				fl.OnUnstage(f.Path)
			}
		})
		u.Importance = widget.LowImportance
		actions = append(actions, u)
	} else {
		if !f.Untracked {
			dbtn := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
				if fl.OnDiscard != nil {
					fl.OnDiscard(f.Path, false)
				}
			})
			dbtn.Importance = widget.LowImportance
			actions = append(actions, dbtn)
		} else {
			dbtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				if fl.OnDiscard != nil {
					fl.OnDiscard(f.Path, true)
				}
			})
			dbtn.Importance = widget.LowImportance
			actions = append(actions, dbtn)
		}
		sbtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			if fl.OnStage != nil {
				fl.OnStage(f.Path)
			}
		})
		sbtn.Importance = widget.LowImportance
		actions = append(actions, sbtn)
	}
	btnBox := container.NewHBox(actions...)

	content := container.NewBorder(nil, nil,
		container.NewPadded(badge),
		btnBox,
		right,
	)

	var bgCol color.Color = clearBG
	if fl.selectedPath == f.Path && fl.selectedStaged == staged {
		bgCol = selectedBG
	}
	bg := canvas.NewRectangle(bgCol)
	fl.rowBG[rowKey{f.Path, staged}] = bg

	fCopy := f
	row := &clickableRow{
		content: container.NewStack(bg, content),
		onTap: func() {
			fl.selectRow(fCopy.Path, staged)
			if fl.OnSelect != nil {
				fl.OnSelect(fCopy, staged)
			}
		},
	}
	row.ExtendBaseWidget(row)
	return row
}

func statusBadge(f git.FileStatus, staged bool) (rune, color.Color) {
	var b byte
	if staged {
		b = f.IndexStatus
	} else {
		b = f.WorkStatus
		if f.Untracked {
			b = 'U'
		}
	}
	switch b {
	case 'M':
		return 'M', color.NRGBA{R: 0xe2, G: 0xc5, B: 0x41, A: 0xff}
	case 'A':
		return 'A', color.NRGBA{R: 0x81, G: 0xd8, B: 0x66, A: 0xff}
	case 'D':
		return 'D', color.NRGBA{R: 0xf0, G: 0x6b, B: 0x6b, A: 0xff}
	case 'R':
		return 'R', color.NRGBA{R: 0x9c, G: 0xdc, B: 0xfe, A: 0xff}
	case 'C':
		return 'C', color.NRGBA{R: 0x9c, G: 0xdc, B: 0xfe, A: 0xff}
	case '?', 'U':
		return 'U', color.NRGBA{R: 0x81, G: 0xd8, B: 0x66, A: 0xff}
	}
	return rune(b), color.NRGBA{R: 0xaa, G: 0xaa, B: 0xaa, A: 0xff}
}

// clickableRow — tapping area wrapping arbitrary content.
type clickableRow struct {
	widget.BaseWidget
	content fyne.CanvasObject
	onTap   func()
}

func (c *clickableRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}

func (c *clickableRow) Tapped(_ *fyne.PointEvent) {
	if c.onTap != nil {
		c.onTap()
	}
}

func (c *clickableRow) TappedSecondary(_ *fyne.PointEvent) {}
