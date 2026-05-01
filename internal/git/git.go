package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type FileStatus struct {
	Path        string
	IndexStatus byte
	WorkStatus  byte
	Staged      bool
	Unstaged    bool
	Untracked   bool
	Conflicted  bool
}

type Repo struct {
	Root string
}

func Open(dir string) (*Repo, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = abs
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.New("not a git repository")
	}
	root := strings.TrimSpace(string(out))
	return &Repo{Root: root}, nil
}

func (r *Repo) run(args ...string) ([]byte, error) {
	return r.runAllow(nil, args...)
}

// runAllow runs git and accepts the given non-zero exit codes as success
// (returning stdout). Any other non-zero exit is treated as error.
func (r *Repo) runAllow(okExits []int, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return stdout.Bytes(), nil
	}
	if ee, ok := err.(*exec.ExitError); ok {
		code := ee.ExitCode()
		for _, c := range okExits {
			if c == code {
				return stdout.Bytes(), nil
			}
		}
	}
	return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
}

func (r *Repo) Status() ([]FileStatus, error) {
	out, err := r.run("status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return nil, err
	}
	var files []FileStatus
	entries := strings.Split(string(out), "\x00")
	for i := 0; i < len(entries); i++ {
		e := entries[i]
		if len(e) < 3 {
			continue
		}
		x, y := e[0], e[1]
		path := e[3:]
		// Renames carry an extra path in the next field
		if x == 'R' || x == 'C' {
			if i+1 < len(entries) {
				i++
			}
		}
		fs := FileStatus{
			Path:        path,
			IndexStatus: x,
			WorkStatus:  y,
		}
		switch {
		case x == '?' && y == '?':
			fs.Untracked = true
			fs.Unstaged = true
		case x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D'):
			fs.Conflicted = true
		default:
			if x != ' ' && x != '?' {
				fs.Staged = true
			}
			if y != ' ' && y != '?' {
				fs.Unstaged = true
			}
		}
		files = append(files, fs)
	}
	return files, nil
}

func (r *Repo) Diff(path string, staged bool) (string, error) {
	args := []string{"diff", "--no-color"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--", path)
	out, err := r.run(args...)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// StagedDiff returns the entire staged diff across all files, suitable for
// feeding to an external tool (e.g. an LLM for commit-message generation).
func (r *Repo) StagedDiff() (string, error) {
	out, err := r.run("diff", "--cached", "--no-color")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (r *Repo) DiffUntracked(path string) (string, error) {
	// git diff --no-index exits with status 1 when files differ — which is the
	// whole point here — so we allow it.
	out, err := r.runAllow([]int{1}, "diff", "--no-color", "--no-index", "--", "/dev/null", path)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (r *Repo) Stage(path string) error {
	_, err := r.run("add", "--", path)
	return err
}

func (r *Repo) Unstage(path string) error {
	if r.hasHEAD() {
		_, err := r.run("restore", "--staged", "--", path)
		return err
	}
	// No commits yet — remove the entry from the index without touching the
	// working tree. The file goes back to untracked.
	_, err := r.run("rm", "--cached", "--", path)
	return err
}

func (r *Repo) Discard(path string, untracked bool) error {
	if untracked {
		_, err := r.run("clean", "-f", "--", path)
		return err
	}
	_, err := r.run("checkout", "--", path)
	return err
}

func (r *Repo) StageAll() error {
	_, err := r.run("add", "-A")
	return err
}

func (r *Repo) UnstageAll() error {
	if r.hasHEAD() {
		_, err := r.run("reset")
		return err
	}
	// No HEAD — empty the index instead.
	_, err := r.run("rm", "-r", "--cached", "--", ".")
	return err
}

func (r *Repo) hasHEAD() bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", "HEAD")
	cmd.Dir = r.Root
	return cmd.Run() == nil
}

func (r *Repo) Commit(msg string) error {
	if strings.TrimSpace(msg) == "" {
		return errors.New("empty commit message")
	}
	_, err := r.run("commit", "-m", msg)
	return err
}

// ApplyPatch applies the given unified diff to the index (--cached).
// Set reverse=true to unstage (useful for line-level unstaging).
func (r *Repo) ApplyPatch(patch string, reverse bool) error {
	args := []string{"apply", "--cached", "--recount", "--whitespace=nowarn", "--unidiff-zero"}
	if reverse {
		args = append(args, "--reverse")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Root
	cmd.Stdin = strings.NewReader(patch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git apply failed: %w: %s", err, stderr.String())
	}
	return nil
}

func (r *Repo) Branch() string {
	out, err := r.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(string(out))
}

// WriteTree snapshots the current index into a tree object and returns its SHA.
// Used as an undo anchor for index-mutating operations.
func (r *Repo) WriteTree() (string, error) {
	out, err := r.run("write-tree")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ReadTree restores the index to the given tree SHA. Working tree is untouched.
func (r *Repo) ReadTree(sha string) error {
	_, err := r.run("read-tree", sha)
	return err
}

// HEADSha returns the current HEAD commit SHA, or "" when there are no commits
// yet (fresh repo). The empty case is not treated as an error.
func (r *Repo) HEADSha() (string, error) {
	if !r.hasHEAD() {
		return "", nil
	}
	out, err := r.run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RemoteBranchesContaining returns the list of remote-tracking branches that
// contain the given commit. An empty result means the commit exists only
// locally — safe to rewrite. Lines like "origin/HEAD -> origin/main" are
// skipped since they're aliases, not real branches.
func (r *Repo) RemoteBranchesContaining(sha string) ([]string, error) {
	out, err := r.run("branch", "-r", "--contains", sha)
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, line := range strings.Split(string(out), "\n") {
		b := strings.TrimSpace(line)
		if b == "" || strings.Contains(b, "->") {
			continue
		}
		branches = append(branches, b)
	}
	return branches, nil
}

// ResetSoft moves HEAD to the given SHA without touching the index or working
// tree. If sha is empty, HEAD is detached (used to undo the initial commit).
func (r *Repo) ResetSoft(sha string) error {
	if sha == "" {
		_, err := r.run("update-ref", "-d", "HEAD")
		return err
	}
	_, err := r.run("reset", "--soft", sha)
	return err
}

// Fetch updates remote-tracking branches without touching local state.
// Safe to call on any repo (no-op if no remotes are configured).
func (r *Repo) Fetch() error {
	if !r.hasRemote() {
		return nil
	}
	_, err := r.run("fetch", "--all", "--prune")
	return err
}

// AheadBehind returns how many commits HEAD is ahead/behind its upstream.
// hasUpstream is false when the current branch has no @{u} configured —
// in that case ahead/behind are zero and the caller should offer
// `push -u` to set upstream on first push.
func (r *Repo) AheadBehind() (ahead, behind int, hasUpstream bool, err error) {
	out, runErr := r.runAllow([]int{128}, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	if runErr != nil {
		// Exit 128 = no upstream configured. Anything else is a real error.
		return 0, 0, false, runErr
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		// No upstream → git prints nothing (we allowed exit 128).
		return 0, 0, false, nil
	}
	fmt.Sscanf(parts[0], "%d", &behind)
	fmt.Sscanf(parts[1], "%d", &ahead)
	return ahead, behind, true, nil
}

// Push pushes the current branch to its upstream. If no upstream is set,
// uses --set-upstream so the next push doesn't need to be told again.
func (r *Repo) Push() error {
	if _, _, hasUp, _ := r.AheadBehind(); !hasUp {
		_, err := r.run("push", "--set-upstream", "origin", "HEAD")
		return err
	}
	_, err := r.run("push")
	return err
}

// PullFFOnly fast-forwards the current branch to its upstream. Refuses to
// merge — if the branches have diverged, the user must resolve in a terminal.
func (r *Repo) PullFFOnly() error {
	_, err := r.run("pull", "--ff-only")
	return err
}

func (r *Repo) hasRemote() bool {
	out, err := r.run("remote")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}
