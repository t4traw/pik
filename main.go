package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/t4traw/pik/internal/git"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// When invoked from a terminal (e.g. `pik` on the CLI), re-exec ourselves
	// via `open -na` so the shell prompt returns immediately instead of being
	// blocked while the GUI window is open. No-op when already running under
	// LaunchServices (no TTY) or when not inside a .app bundle (dev mode).
	if relaunchDetached() {
		return
	}

	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	repo, err := git.Open(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pik: %v\n", err)
		os.Exit(1)
	}

	app := NewApp(repo)

	err = wails.Run(&options.App{
		Title:  "pik — " + repo.Root,
		Width:  1200,
		Height: 780,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHidden(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "pik: %v\n", err)
		os.Exit(1)
	}
}

// relaunchDetached spawns a detached copy of ourselves via `open -na` and
// returns true, signalling the caller to exit. Returns false (and does
// nothing) when a relaunch isn't needed or possible.
func relaunchDetached() bool {
	// Escape hatch — set in `wails dev` or when debugging the binary directly.
	if os.Getenv("PIK_NO_RELAUNCH") != "" {
		return false
	}
	// LaunchServices always sets __CFBundleIdentifier when launching a .app
	// (including via `open -na`). If it's set we're the relaunched child —
	// returning here is what stops the open→spawn→open infinite loop that
	// shows up when the Homebrew shim `exec`s us with the parent shell's TTY
	// still attached (so the stdin/stdout TTY check below isn't enough).
	if os.Getenv("__CFBundleIdentifier") != "" {
		return false
	}
	// No controlling TTY on stdin → not invoked from a shell, nothing to do.
	stat, err := os.Stdin.Stat()
	if err != nil || stat.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	// Expect path of the form /…/pik.app/Contents/MacOS/pik. Walk up three
	// levels and require a .app bundle — otherwise we're running outside a
	// bundle (e.g. `go run .`) and there's nothing to hand off to `open`.
	appPath := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
	if !strings.HasSuffix(appPath, ".app") {
		return false
	}
	// Resolve the repo argument to an absolute path while we still have the
	// caller's cwd — `open` launches the app from LaunchServices, which does
	// not inherit it.
	cwd, _ := os.Getwd()
	dir := cwd
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(cwd, dir)
	}
	cmd := exec.Command("/usr/bin/open", "-na", appPath, "--args", dir)
	if err := cmd.Start(); err != nil {
		return false
	}
	return true
}
