package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/app"
)

//go:embed all:cmd/hawkward-gui/frontend/dist
var assets embed.FS

func main() {
	application := app.NewApp()

	err := wails.Run(&options.App{
		Title:     "Hawkward Operations Platform",
		Width:     1400,
		Height:    900,
		MinWidth:  1024,
		MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  application.Startup,
		OnShutdown: application.Shutdown,
		Bind: []interface{}{
			application,
			application.SysOps,
			application.NetOps,
			application.SecOps,
			application.DevOps,
			application.AIOps,
			application.Dash,
			application.PipelineAPI,
			application.AlertAPI,
			application.Logs,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
