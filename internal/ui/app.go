package ui

import (
	"image/color"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/t4traw/pik/internal/git"
)

type App struct {
	fyneApp fyne.App
	win     fyne.Window
	repo    *git.Repo

	fileList   *FileList
	diffView   *DiffView
	commitMsg  *widget.Entry
	commitBtn  *widget.Button
	branchLbl  *canvas.Text
	statusText *widget.Label
	modeBtn    *widget.Button

	currentPath   string
	currentStaged bool

	// diffCache holds raw diff output keyed by (path, staged, untracked).
	// Invalidated on every refresh (stage/unstage/commit/discard).
	diffCache map[diffCacheKey]string
}

type diffCacheKey struct {
	path      string
	staged    bool
	untracked bool
}

func Run(repo *git.Repo) error {
	a := app.NewWithID("com.t4traw.pik")
	a.Settings().SetTheme(NewDarkTheme())

	w := a.NewWindow("pik — " + filepath.Base(repo.Root))
	w.Resize(fyne.NewSize(1100, 720))

	ap := &App{
		fyneApp:   a,
		win:       w,
		repo:      repo,
		diffCache: map[diffCacheKey]string{},
	}
	ap.build()
	ap.refresh()

	w.SetContent(ap.content())
	w.ShowAndRun()
	return nil
}

func (a *App) build() {
	a.fileList = NewFileList()
	a.fileList.OnSelect = func(f git.FileStatus, staged bool) {
		a.showDiff(f, staged)
	}
	a.fileList.OnStage = func(p string) { a.do(func() error { return a.repo.Stage(p) }) }
	a.fileList.OnUnstage = func(p string) { a.do(func() error { return a.repo.Unstage(p) }) }
	a.fileList.OnDiscard = func(p string, untracked bool) {
		msg := "変更を破棄しますか?\n" + p
		if untracked {
			msg = "未追跡ファイルを削除しますか?\n" + p
		}
		dialog.NewConfirm("確認", msg, func(ok bool) {
			if !ok {
				return
			}
			a.do(func() error { return a.repo.Discard(p, untracked) })
		}, a.win).Show()
	}
	a.fileList.OnStageAll = func() { a.do(func() error { return a.repo.StageAll() }) }
	a.fileList.OnResetAll = func() { a.do(func() error { return a.repo.UnstageAll() }) }

	a.diffView = NewDiffView()
	a.diffView.SetEmpty("ファイルを選択してね")

	a.commitMsg = widget.NewMultiLineEntry()
	a.commitMsg.SetPlaceHolder("コミットメッセージ (Ctrl/Cmd+Enter で確定)")
	a.commitMsg.Wrapping = fyne.TextWrapWord

	a.commitBtn = widget.NewButtonWithIcon("コミット", theme.ConfirmIcon(), a.commit)
	a.commitBtn.Importance = widget.HighImportance

	a.branchLbl = canvas.NewText("● "+a.repo.Branch(), color.NRGBA{R: 0x9c, G: 0xdc, B: 0xfe, A: 0xff})
	a.branchLbl.TextStyle = fyne.TextStyle{Bold: true}
	a.branchLbl.TextSize = 12

	a.modeBtn = widget.NewButtonWithIcon("Split", theme.ViewRestoreIcon(), a.toggleMode)
	a.modeBtn.Importance = widget.LowImportance

	a.statusText = widget.NewLabel("")

	// Keyboard shortcut: Cmd/Ctrl+Enter to commit
	a.win.Canvas().AddShortcut(&fyne.ShortcutSelectAll{}, func(fyne.Shortcut) {})
	a.win.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		// no-op; left as extension point
	})
}

func (a *App) content() fyne.CanvasObject {
	// Title bar
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), a.refresh)
	refreshBtn.Importance = widget.LowImportance

	rightBar := container.NewHBox(a.modeBtn, refreshBtn)
	title := container.NewBorder(nil, nil,
		container.NewPadded(a.branchLbl),
		rightBar,
		widget.NewLabel(a.repo.Root),
	)
	titleBG := canvas.NewRectangle(color.NRGBA{R: 0x25, G: 0x25, B: 0x26, A: 0xff})
	titleBar := container.NewStack(titleBG, container.NewPadded(title))

	// Commit area
	commitArea := container.NewBorder(
		widget.NewSeparator(),
		container.NewBorder(nil, nil, nil, a.commitBtn, nil),
		nil, nil,
		container.NewPadded(a.commitMsg),
	)

	// Left: file list + commit area
	left := container.NewBorder(
		nil, commitArea, nil, nil,
		a.fileList,
	)

	// Split
	split := container.NewHSplit(left, a.diffView)
	split.Offset = 0.32

	// Status bar
	statusBG := canvas.NewRectangle(color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff})
	statusBar := container.NewStack(statusBG, container.NewPadded(a.statusText))

	return container.NewBorder(titleBar, statusBar, nil, nil, split)
}

func (a *App) currentFiles() []git.FileStatus {
	files, err := a.repo.Status()
	if err != nil {
		a.statusText.SetText("status: " + err.Error())
		return nil
	}
	return files
}

func (a *App) refresh() {
	// Any state-changing op (stage/unstage/commit/discard) invalidates cached diffs.
	a.diffCache = map[diffCacheKey]string{}
	files := a.currentFiles()
	a.fileList.SetFiles(files)
	a.branchLbl.Text = "● " + a.repo.Branch()
	a.branchLbl.Refresh()

	// refresh diff for current selection
	sel, staged := a.fileList.Selected()
	if sel == "" {
		a.diffView.SetEmpty("ファイルを選択してね")
	} else {
		for _, f := range files {
			if f.Path == sel {
				a.showDiff(f, staged)
				a.statusText.SetText("")
				return
			}
		}
		a.diffView.SetEmpty("ファイルを選択してね")
	}
	a.statusText.SetText("")
}

func (a *App) showDiff(f git.FileStatus, staged bool) {
	// Same file + same mode? Nothing to do.
	if a.currentPath == f.Path && a.currentStaged == staged {
		return
	}
	a.currentPath = f.Path
	a.currentStaged = staged

	key := diffCacheKey{path: f.Path, staged: staged, untracked: f.Untracked}
	diff, hit := a.diffCache[key]
	if !hit {
		var err error
		if f.Untracked {
			diff, err = a.repo.DiffUntracked(f.Path)
		} else {
			diff, err = a.repo.Diff(f.Path, staged)
		}
		if err != nil {
			a.diffView.SetEmpty("diff取得エラー: " + err.Error())
			return
		}
		a.diffCache[key] = diff
	}
	a.diffView.SetDiff(diff, f.Path)
}

func (a *App) toggleMode() {
	if a.diffView.Mode() == ModeUnified {
		a.diffView.SetMode(ModeSplit)
		a.modeBtn.SetText("Unified")
	} else {
		a.diffView.SetMode(ModeUnified)
		a.modeBtn.SetText("Split")
	}
}

func (a *App) do(fn func() error) {
	if err := fn(); err != nil {
		dialog.ShowError(err, a.win)
		return
	}
	a.refresh()
}

func (a *App) commit() {
	msg := a.commitMsg.Text
	if err := a.repo.Commit(msg); err != nil {
		dialog.ShowError(err, a.win)
		return
	}
	a.commitMsg.SetText("")
	a.statusText.SetText("コミット完了")
	a.refresh()
}
