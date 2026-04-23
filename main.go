package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/t4traw/pik/internal/git"
)

//go:embed all:frontend/dist
var assets embed.FS

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
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "pik: %v\n", err)
		os.Exit(1)
	}
}
