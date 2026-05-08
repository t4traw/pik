package git

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ConflictRegion is one <<<<<<< / ======= / >>>>>>> block in a conflicted file.
// Lines are stored without their trailing newline; the renderer is expected to
// add separators back. StartLine/EndLine are 1-based, inclusive, and reference
// the working-tree file as it exists right now (i.e. after the conflict).
type ConflictRegion struct {
	StartLine   int      `json:"startLine"`
	EndLine     int      `json:"endLine"`
	OursLabel   string   `json:"oursLabel"`
	TheirsLabel string   `json:"theirsLabel"`
	OursLines   []string `json:"oursLines"`
	TheirsLines []string `json:"theirsLines"`
	BaseLines   []string `json:"baseLines"` // populated only with merge.conflictstyle=diff3
}

// ConflictFile bundles all conflict regions in a single working-tree file plus
// the surrounding "calm" lines, so the frontend can rebuild the resolved
// content without re-reading the file.
type ConflictFile struct {
	Path    string           `json:"path"`
	Regions []ConflictRegion `json:"regions"`
	// Lines holds every line of the file in order, including the conflict
	// markers themselves. Frontend reconstructs the resolved file by replacing
	// each region [StartLine..EndLine] with the user's selection.
	Lines  []string `json:"lines"`
	Binary bool     `json:"binary"`
}

// RebaseState reports whether a rebase or merge is in flight. Used by the UI
// to decide whether to show the "continue / abort" footer.
type RebaseState struct {
	Rebasing bool   `json:"rebasing"` // .git/rebase-merge or .git/rebase-apply
	Merging  bool   `json:"merging"`  // .git/MERGE_HEAD
	Step     int    `json:"step"`     // current step number (rebase only)
	Total    int    `json:"total"`    // total steps (rebase only)
	Head     string `json:"head"`     // branch we were on when rebase started
}

// State inspects .git/* sentinels to determine the current operation mode.
func (r *Repo) State() (RebaseState, error) {
	st := RebaseState{}
	rmDir := filepath.Join(r.Root, ".git", "rebase-merge")
	raDir := filepath.Join(r.Root, ".git", "rebase-apply")
	if _, err := os.Stat(rmDir); err == nil {
		st.Rebasing = true
		st.Step = readIntFile(filepath.Join(rmDir, "msgnum"))
		st.Total = readIntFile(filepath.Join(rmDir, "end"))
		st.Head = strings.TrimPrefix(readTrimmed(filepath.Join(rmDir, "head-name")), "refs/heads/")
	} else if _, err := os.Stat(raDir); err == nil {
		st.Rebasing = true
		st.Step = readIntFile(filepath.Join(raDir, "next"))
		st.Total = readIntFile(filepath.Join(raDir, "last"))
		st.Head = strings.TrimPrefix(readTrimmed(filepath.Join(raDir, "head-name")), "refs/heads/")
	}
	if _, err := os.Stat(filepath.Join(r.Root, ".git", "MERGE_HEAD")); err == nil {
		st.Merging = true
	}
	return st, nil
}

// Conflicts lists every working-tree file that git considers conflicted. We
// reuse Status() so the same parsing rules apply and there's only one source
// of truth for "what's conflicted".
func (r *Repo) Conflicts() ([]string, error) {
	files, err := r.Status()
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, f := range files {
		if f.Conflicted {
			paths = append(paths, f.Path)
		}
	}
	return paths, nil
}

// ParseConflictFile reads the working-tree file and splits it into the regions
// the UI needs. Binary files are reported with Binary=true and zero regions —
// the caller should offer file-level ours/theirs only.
func (r *Repo) ParseConflictFile(path string) (ConflictFile, error) {
	full := filepath.Join(r.Root, path)
	data, err := os.ReadFile(full)
	if err != nil {
		return ConflictFile{Path: path}, err
	}
	if isBinary(data) {
		return ConflictFile{Path: path, Binary: true}, nil
	}
	cf := ConflictFile{Path: path}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	// Default 64K buffer is too small for large generated files.
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		cf.Lines = append(cf.Lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return cf, err
	}
	cf.Regions, err = parseConflictRegions(cf.Lines)
	if err != nil {
		return cf, err
	}
	return cf, nil
}

// parseConflictRegions walks the file once, tracking which side of the
// conflict marker block we're in. Marker prefixes per git docs:
//
//	<<<<<<<  start ours
//	|||||||  base (only with merge.conflictstyle=diff3)
//	=======  separator
//	>>>>>>>  end theirs
//
// We require markers to start at column 0 and have the standard 7-character
// prefix; that matches what every git plumbing/porcelain emits.
func parseConflictRegions(lines []string) ([]ConflictRegion, error) {
	const (
		stateNormal = iota
		stateOurs
		stateBase
		stateTheirs
	)
	state := stateNormal
	var regions []ConflictRegion
	var cur ConflictRegion
	for i, line := range lines {
		ln := i + 1
		switch {
		case strings.HasPrefix(line, "<<<<<<<"):
			if state != stateNormal {
				return nil, fmt.Errorf("unexpected <<<<<<< at line %d", ln)
			}
			cur = ConflictRegion{StartLine: ln, OursLabel: strings.TrimSpace(line[7:])}
			state = stateOurs
		case strings.HasPrefix(line, "|||||||") && state == stateOurs:
			state = stateBase
		case strings.HasPrefix(line, "=======") && (state == stateOurs || state == stateBase):
			state = stateTheirs
		case strings.HasPrefix(line, ">>>>>>>") && state == stateTheirs:
			cur.EndLine = ln
			cur.TheirsLabel = strings.TrimSpace(line[7:])
			regions = append(regions, cur)
			cur = ConflictRegion{}
			state = stateNormal
		default:
			switch state {
			case stateOurs:
				cur.OursLines = append(cur.OursLines, line)
			case stateBase:
				cur.BaseLines = append(cur.BaseLines, line)
			case stateTheirs:
				cur.TheirsLines = append(cur.TheirsLines, line)
			}
		}
	}
	if state != stateNormal {
		return nil, errors.New("unterminated conflict block")
	}
	return regions, nil
}

// WriteResolved writes the given content to the file and stages it. The
// caller is responsible for composing the final content from the user's
// per-region selections — this function just persists and marks resolved.
func (r *Repo) WriteResolved(path, content string) error {
	full := filepath.Join(r.Root, path)
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		return err
	}
	_, err := r.run("add", "--", path)
	return err
}

// ResolveOurs / ResolveTheirs are file-level shortcuts. Useful for binary
// files (where hunk-level resolution is impossible) and for "I just want to
// take this whole side" cases.
func (r *Repo) ResolveOurs(path string) error {
	if _, err := r.run("checkout", "--ours", "--", path); err != nil {
		return err
	}
	_, err := r.run("add", "--", path)
	return err
}

func (r *Repo) ResolveTheirs(path string) error {
	if _, err := r.run("checkout", "--theirs", "--", path); err != nil {
		return err
	}
	_, err := r.run("add", "--", path)
	return err
}

// ContinueRebase resumes a rebase after the user has resolved all conflicts.
// We pre-set GIT_EDITOR to true so git doesn't open an interactive editor for
// the resumed commit message — pik is a GUI app with no TTY.
func (r *Repo) ContinueRebase() error {
	cmd := r.runCmd("rebase", "--continue")
	cmd.Env = append(os.Environ(), "GIT_EDITOR=true")
	return cmd.Run()
}

// AbortRebase rolls the working tree back to the state before the rebase
// started. Safe even mid-conflict.
func (r *Repo) AbortRebase() error {
	_, err := r.run("rebase", "--abort")
	return err
}

// runCmd is like run but returns the *exec.Cmd so callers can tweak env/stdio
// before invoking. Only used for cases (like rebase --continue) that need to
// suppress the editor.
func (r *Repo) runCmd(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Root
	return cmd
}

// readIntFile reads a small text file containing an integer. Returns 0 on
// any error — these files are advisory (rebase progress display) so a stale
// value just shows "0/0" briefly.
func readIntFile(p string) int {
	s := readTrimmed(p)
	n := 0
	fmt.Sscanf(s, "%d", &n)
	return n
}

func readTrimmed(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// isBinary heuristic: a NUL byte in the first 8KB. Matches what git uses
// internally for diff classification.
func isBinary(data []byte) bool {
	n := len(data)
	if n > 8000 {
		n = 8000
	}
	for i := 0; i < n; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}
