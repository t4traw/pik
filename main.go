package main

import (
	"fmt"
	"os"

	"github.com/t4traw/pik/internal/git"
	"github.com/t4traw/pik/internal/ui"
)

func main() {
	// Before any Fyne theme init: if the user has a CJK-capable monospace font
	// installed, point Fyne at it so Japanese renders correctly.
	ui.DetectAndSetMonoFont()

	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	repo, err := git.Open(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pik: %v\n", err)
		os.Exit(1)
	}
	if err := ui.Run(repo); err != nil {
		fmt.Fprintf(os.Stderr, "pik: %v\n", err)
		os.Exit(1)
	}
}
