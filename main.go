package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/tpsapp/applauncher3/pkg/cli"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	opts, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "AppLauncher3 CLI Error: %v\n", err)
		os.Exit(2)
	}

	app, err := NewApp(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initialization Error: %v\n", err)
		os.Exit(1)
	}

	startHidden := false
	if opts.StartMinimized {
		startHidden = true
	}

	err = wails.Run(&options.App{
		Title:             "AppLauncher3",
		Width:             960,
		Height:            680,
		MinWidth:          720,
		MinHeight:         520,
		StartHidden:       startHidden,
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 24, G: 24, B: 27, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			Icon:                nil,
			WindowIsTranslucent: false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime Error: %v\n", err)
		os.Exit(1)
	}
}
