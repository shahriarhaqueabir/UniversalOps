package main

import (
	"embed"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/shahriarhaqueabir/AllOpsFull/internal/app"
)

//go:embed all:cmd/opsforall-gui/frontend/dist
var assets embed.FS

func main() {
	// Start pprof server for Phase 0 instrumentation
	go func() {
		log.Println("pprof server starting on :6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Printf("pprof server failed: %v", err)
		}
	}()

	application := app.NewApp()

	err := wails.Run(&options.App{
		Title:     "Universal-Ops Operations Platform",
		Width:     1400,
		Height:    900,
		MinWidth:  1024,
		MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  application.Startup,
		OnShutdown: application.Shutdown,
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			BackdropType:         windows.Mica,
			DisableWindowIcon:    false,

			Theme:                windows.Dark,
			CustomTheme:          &windows.ThemeSettings{},
			IsZoomControlEnabled: false,

			DisableFramelessWindowDecorations: false,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
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
			application.Timeline,
			application.Workflows,
			application.Reports,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
