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
	_, err := r.run("restore", "--staged", "--", path)
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
	_, err := r.run("reset")
	return err
}

func (r *Repo) Commit(msg string) error {
	if strings.TrimSpace(msg) == "" {
		return errors.New("empty commit message")
	}
	_, err := r.run("commit", "-m", msg)
	return err
}

func (r *Repo) Branch() string {
	out, err := r.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(string(out))
}
