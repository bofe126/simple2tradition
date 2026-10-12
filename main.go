package main

import (
	"embed"
	"log"
	"os"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Apps launched from Finder have no LANG, which makes the
	// pbpaste/pbcopy helpers (used by the clipboard library on
	// macOS) transcode non-ASCII text to '?'. Force UTF-8.
	if goruntime.GOOS == "darwin" && os.Getenv("LANG") == "" {
		os.Setenv("LANG", "en_US.UTF-8")
	}

	app := NewApp()

	// macOS has no tray, so the app menu carries Quit (Cmd+Q);
	// Edit menu enables the standard Cmd+C/V/X shortcuts in the
	// WebView.
	var appMenu *menu.Menu
	if goruntime.GOOS == "darwin" {
		appMenu = menu.NewMenu()
		appSub := appMenu.AddSubmenu("简繁")
		appSub.AddText("退出 简繁", keys.CmdOrCtrl("q"), func(*menu.CallbackData) {
			app.RequestQuit()
		})
		appMenu.Append(menu.EditMenu())
	}

	err := wails.Run(&options.App{
		Menu:             appMenu,
		Title:            "简繁",
		Width:            480,
		Height:           560,
		MinWidth:         360,
		MinHeight:        480,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 24, G: 26, B: 32, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		Bind:             []interface{}{app},
		// Start hidden when the user has "start minimized" enabled;
		// they can always click the tray icon to bring up the
		// window. We respect the saved setting at startup by
		// hiding in OnDomReady if needed (see app.domReady).
		// The tray only exists on Windows; elsewhere the window
		// must be visible or there is no way to reach the UI.
		StartHidden: goruntime.GOOS == "windows",
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})
	if err != nil {
		log.Fatal("启动失败:", err)
	}
}
