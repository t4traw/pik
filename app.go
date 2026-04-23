package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/t4traw/pik/internal/git"
	"github.com/t4traw/pik/internal/settings"
)

// App is bound to the Wails frontend. All exported methods are callable
// from JS/TS and their arguments/returns are JSON-marshaled automatically.
type App struct {
	ctx  context.Context
	repo *git.Repo

	mu        sync.Mutex
	undoStack []undoOp
	redoStack []undoOp
}

// undoOp is a single reversible action. Desc is shown in the UI; Undo rolls
// the repo back to the state before the action ran; Redo re-applies it.
type undoOp struct {
	Desc string
	Undo func() error
	Redo func() error
}

func NewApp(repo *git.Repo) *App {
	return &App{repo: repo}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ---- read-only info ----

type RepoInfo struct {
	Root   string `json:"root"`
	Branch string `json:"branch"`
}

func (a *App) Info() RepoInfo {
	return RepoInfo{Root: a.repo.Root, Branch: a.repo.Branch()}
}

func (a *App) Status() ([]git.FileStatus, error) {
	return a.repo.Status()
}

type DiffResult struct {
	Files []git.FileDiff `json:"files"`
	Raw   string         `json:"raw"`
}

// Diff returns parsed diff for the given file. For untracked files, the raw
// content is emitted as a synthetic add-only hunk (via git diff --no-index).
func (a *App) Diff(path string, staged bool, untracked bool) (DiffResult, error) {
	var raw string
	var err error
	if untracked {
		raw, err = a.repo.DiffUntracked(path)
	} else {
		raw, err = a.repo.Diff(path, staged)
	}
	if err != nil {
		return DiffResult{}, err
	}
	return DiffResult{Files: git.ParseUnifiedDiff(raw), Raw: raw}, nil
}

// ---- undo / redo infra ----

// UndoState lets the frontend decide whether to enable the undo/redo buttons
// and what description to show in a tooltip or toast.
type UndoState struct {
	CanUndo  bool   `json:"canUndo"`
	CanRedo  bool   `json:"canRedo"`
	UndoDesc string `json:"undoDesc"`
	RedoDesc string `json:"redoDesc"`
}

func (a *App) UndoState() UndoState {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := UndoState{CanUndo: len(a.undoStack) > 0, CanRedo: len(a.redoStack) > 0}
	if s.CanUndo {
		s.UndoDesc = a.undoStack[len(a.undoStack)-1].Desc
	}
	if s.CanRedo {
		s.RedoDesc = a.redoStack[len(a.redoStack)-1].Desc
	}
	return s
}

// Undo reverts the most recent action and returns its description (empty
// string means the stack was empty and nothing was done).
func (a *App) Undo() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.undoStack) == 0 {
		return "", nil
	}
	op := a.undoStack[len(a.undoStack)-1]
	a.undoStack = a.undoStack[:len(a.undoStack)-1]
	if err := op.Undo(); err != nil {
		// Re-push on failure so the user can retry / inspect.
		a.undoStack = append(a.undoStack, op)
		return "", err
	}
	a.redoStack = append(a.redoStack, op)
	return op.Desc, nil
}

func (a *App) Redo() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.redoStack) == 0 {
		return "", nil
	}
	op := a.redoStack[len(a.redoStack)-1]
	a.redoStack = a.redoStack[:len(a.redoStack)-1]
	if err := op.Redo(); err != nil {
		a.redoStack = append(a.redoStack, op)
		return "", err
	}
	a.undoStack = append(a.undoStack, op)
	return op.Desc, nil
}

// push appends a new entry and clears redo history (standard editor semantics:
// performing a new action invalidates everything you redid past).
func (a *App) push(op undoOp) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.undoStack = append(a.undoStack, op)
	a.redoStack = nil
}

// indexSnapshot returns an undoOp that restores the index to its current tree.
// Used as the Undo for any operation that only mutates the index.
func (a *App) indexSnapshot(desc string, apply func() error) error {
	treeBefore, err := a.repo.WriteTree()
	if err != nil {
		return err
	}
	if err := apply(); err != nil {
		return err
	}
	a.push(undoOp{
		Desc: desc,
		Undo: func() error { return a.repo.ReadTree(treeBefore) },
		Redo: apply,
	})
	return nil
}

// ---- mutations ----

func (a *App) Stage(path string) error {
	return a.indexSnapshot("ステージ: "+path, func() error { return a.repo.Stage(path) })
}

func (a *App) Unstage(path string) error {
	return a.indexSnapshot("アンステージ: "+path, func() error { return a.repo.Unstage(path) })
}

func (a *App) StageAll() error {
	return a.indexSnapshot("全ステージ", func() error { return a.repo.StageAll() })
}

func (a *App) UnstageAll() error {
	return a.indexSnapshot("全アンステージ", func() error { return a.repo.UnstageAll() })
}

// StageLines applies only the selected lines of the given hunks to the index.
func (a *App) StageLines(path string, hunks []git.PatchHunk) error {
	patch := git.BuildSubPatch(path, hunks)
	if patch == "" {
		return fmt.Errorf("no lines selected")
	}
	apply := func() error { return a.repo.ApplyPatch(patch, false) }
	reverse := func() error { return a.repo.ApplyPatch(patch, true) }
	if err := apply(); err != nil {
		return err
	}
	a.push(undoOp{Desc: "行ステージ: " + path, Undo: reverse, Redo: apply})
	return nil
}

// UnstageLines reverses a sub-patch against the index.
func (a *App) UnstageLines(path string, hunks []git.PatchHunk) error {
	patch := git.BuildSubPatch(path, hunks)
	if patch == "" {
		return fmt.Errorf("no lines selected")
	}
	apply := func() error { return a.repo.ApplyPatch(patch, true) }
	reverse := func() error { return a.repo.ApplyPatch(patch, false) }
	if err := apply(); err != nil {
		return err
	}
	a.push(undoOp{Desc: "行アンステージ: " + path, Undo: reverse, Redo: apply})
	return nil
}

// Discard reverts working-tree changes (or deletes the file, for untracked).
// We snapshot the file's current disk content first so undo can put it back.
func (a *App) Discard(path string, untracked bool) error {
	full := filepath.Join(a.repo.Root, path)
	content, err := os.ReadFile(full)
	if err != nil {
		return fmt.Errorf("snapshot %s: %w", path, err)
	}
	// File mode — preserve the executable bit on undo.
	info, _ := os.Stat(full)
	mode := os.FileMode(0644)
	if info != nil {
		mode = info.Mode().Perm()
	}
	if err := a.repo.Discard(path, untracked); err != nil {
		return err
	}
	a.push(undoOp{
		Desc: "破棄: " + path,
		Undo: func() error {
			if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
				return err
			}
			return os.WriteFile(full, content, mode)
		},
		Redo: func() error { return a.repo.Discard(path, untracked) },
	})
	return nil
}

// ---- AI commit message ----

// generateCommitPrompt — kept short on purpose. Project conventions live in
// CLAUDE.md; `claude -p` picks those up automatically via the repo cwd.
const generateCommitPrompt = `stdin のステージ済み git 差分を読んで、このリポジトリの CLAUDE.md に従った 1 行のコミットメッセージを出力してください。

出力ルール:
- コミットメッセージ本文のみを 1 行で出力
- 前置き・説明・コードブロック・引用符を一切含めない
- 先頭に "commit:" や "Message:" のようなラベルを付けない`

// GenerateCommitMessage shells out to the locally-installed `claude` CLI and
// asks it to propose a commit message for the current staged diff. Requires
// the user to have `claude` on PATH and be logged in.
func (a *App) GenerateCommitMessage() (string, error) {
	diff, err := a.repo.StagedDiff()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("ステージ済みの変更がありません")
	}

	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "-p", generateCommitPrompt)
	// cwd = repo root so Claude Code auto-loads the project's CLAUDE.md.
	cmd.Dir = a.repo.Root
	cmd.Stdin = strings.NewReader(diff)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("claude CLI が見つかりません。Claude Code をインストールしてください")
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("claude の応答がタイムアウトしました (60秒)")
		}
		return "", fmt.Errorf("claude 実行エラー: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	msg := strings.TrimSpace(stdout.String())
	// Occasionally the model wraps the line in backticks despite instructions.
	msg = strings.Trim(msg, "`")
	if idx := strings.IndexByte(msg, '\n'); idx >= 0 {
		msg = strings.TrimSpace(msg[:idx])
	}
	return msg, nil
}

// ---- settings ----

func (a *App) GetSettings() settings.Settings {
	s, _ := settings.Load()
	return s
}

func (a *App) UpdateSettings(s settings.Settings) (settings.Settings, error) {
	if err := settings.Save(s); err != nil {
		return settings.Defaults(), err
	}
	return settings.Sanitize(s), nil
}

// Commit creates a new commit and records the pre-commit HEAD so undo can
// soft-reset back. Empty pre-HEAD means this was the initial commit.
func (a *App) Commit(msg string) error {
	prev, err := a.repo.HEADSha()
	if err != nil {
		return err
	}
	if err := a.repo.Commit(msg); err != nil {
		return err
	}
	a.push(undoOp{
		Desc: "コミット: " + msg,
		Undo: func() error {
			// Refuse to rewind past a commit that's already on a remote —
			// undoing it locally would diverge from origin and require a
			// force-push to reconcile.
			cur, err := a.repo.HEADSha()
			if err != nil {
				return err
			}
			remotes, err := a.repo.RemoteBranchesContaining(cur)
			if err != nil {
				return err
			}
			if len(remotes) > 0 {
				return fmt.Errorf("push 済みのため取り消せません (%s)", strings.Join(remotes, ", "))
			}
			return a.repo.ResetSoft(prev)
		},
		Redo: func() error { return a.repo.Commit(msg) },
	})
	return nil
}
