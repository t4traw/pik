package main

import (
	"context"
	"fmt"

	"github.com/t4traw/pik/internal/git"
)

// App is bound to the Wails frontend. All exported methods are callable
// from JS/TS and their arguments/returns are JSON-marshaled automatically.
type App struct {
	ctx  context.Context
	repo *git.Repo
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

// ---- mutations ----

func (a *App) Stage(path string) error         { return a.repo.Stage(path) }
func (a *App) Unstage(path string) error       { return a.repo.Unstage(path) }
func (a *App) StageAll() error                 { return a.repo.StageAll() }
func (a *App) UnstageAll() error               { return a.repo.UnstageAll() }
func (a *App) Discard(path string, untracked bool) error {
	return a.repo.Discard(path, untracked)
}

// StageLines applies only the selected lines of the given hunks to the index.
// On the unstaged side → call this. Pass the (sub-)hunks the user picked.
func (a *App) StageLines(path string, hunks []git.PatchHunk) error {
	patch := git.BuildSubPatch(path, hunks)
	if patch == "" {
		return fmt.Errorf("no lines selected")
	}
	return a.repo.ApplyPatch(patch, false)
}

// UnstageLines reverses a sub-patch against the index — i.e. takes the
// currently-staged diff of the file and unstages only the selected portions.
func (a *App) UnstageLines(path string, hunks []git.PatchHunk) error {
	patch := git.BuildSubPatch(path, hunks)
	if patch == "" {
		return fmt.Errorf("no lines selected")
	}
	return a.repo.ApplyPatch(patch, true)
}

func (a *App) Commit(msg string) error { return a.repo.Commit(msg) }
