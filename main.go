package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/shahriarhaqueabir/UniversalOps/internal/app"
)

//go:embed all:cmd/opsforall-gui/frontend/dist
var assets embed.FS

func main() {
	// Check OS-level prerequisites before initialising the UI.
	// On Windows this verifies the WebView2 Runtime is installed.
	checkWindowsPrereqs()

	// pprof server removed — violates 100% local / zero telemetry north star.
	// Re-enable with -tags debug if needed for profiling sessions.

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
			BackdropType:         backdropType(),
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
