package ui

import (
	"os"
	"runtime"
)

// Menlo is shipped as a .ttc collection which Fyne's font loader rejects
// ("collections not allowed"), so we use SF Mono (SFNSMono.ttf) instead — it
// is a plain .ttf that exists on every modern macOS install. Japanese glyphs
// missing from SF Mono fall back to the system font map registered by Fyne's
// loadSystemFonts().
const macMonoPath = "/System/Library/Fonts/SFNSMono.ttf"

// DetectAndSetMonoFont points Fyne at Menlo on macOS. Must be called before
// any theme init (i.e. before app.New). No-op if FYNE_FONT_MONOSPACE is
// already set or on non-macOS platforms.
func DetectAndSetMonoFont() {
	if os.Getenv("FYNE_FONT_MONOSPACE") != "" {
		return
	}
	if runtime.GOOS != "darwin" {
		return
	}
	if st, err := os.Stat(macMonoPath); err != nil || st.IsDir() {
		return
	}
	_ = os.Setenv("FYNE_FONT_MONOSPACE", macMonoPath)
}
