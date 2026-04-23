package main

import (
	"fmt"
	"os"

	"github.com/t4traw/pik/internal/git"
	"github.com/t4traw/pik/internal/ui"
)

func main() {
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
