//go:build windows

// Package tray wraps github.com/getlantern/systray with a small
// callback-oriented API suited to the simple2tradition App. It
// contains the Windows-only implementation; tray_other.go provides
// a no-op stub for non-Windows builds so `go vet ./...` and
// cross-compilation still work.
package tray

import (
	_ "embed"
	"log"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var iconData []byte

// SystemTray is the system-tray icon + menu. Create via New,
// configure callbacks, then call Run in its own goroutine.
type SystemTray struct {
	enabled bool

	// Callbacks wired by the App.
	OnShow     func()
	OnSettings func()
	OnToggle   func(enabled bool)
	OnQuit     func()

	// Menu item kept so we can flip the checkbox.
	mToggle *systray.MenuItem
}

// New returns a SystemTray that defaults to "enabled" (matches
// config.Default().Enabled). Callbacks are nil-safe; the App
// wires them before calling Run.
func New() *SystemTray {
	return &SystemTray{enabled: true}
}

// SetDirection is a no-op on Windows: the tray icon is static and
// the direction lives in the app window UI.
func (s *SystemTray) SetDirection(dir string) {}

// SetEnabled updates the checkbox state to match the live config.
// Safe to call from any goroutine after the menu is constructed.
func (s *SystemTray) SetEnabled(v bool) {
	s.enabled = v
	if s.mToggle != nil {
		if v {
			s.mToggle.Check()
		} else {
			s.mToggle.Uncheck()
		}
	}
}

// Run blocks on the systray event loop until systray.Quit is called.
// Intended to be invoked via `go tray.Run()` from the App's startup.
func (s *SystemTray) Run() {
	systray.Run(s.onReady, s.onExit)
}

// Close signals systray to exit. The App calls this from its
// shutdown hook.
func (s *SystemTray) Close() {
	systray.Quit()
}

func (s *SystemTray) onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("简繁")
	systray.SetTooltip("简繁 — 剪贴板简繁转换")

	// Note: getlantern/systray shows the menu on BOTH left- and
	// right-click of the tray icon (its wndProc routes both
	// WM_LBUTTONUP and WM_RBUTTONUP to showMenu). That is a
	// deliberate design choice in the library. We therefore put
	// "显示主窗口" as the very first item so users who click the
	// icon to bring up the window get one obvious top-level entry.

	mShow := systray.AddMenuItem("显示主窗口", "打开 简繁 主窗口")
	go func() {
		for range mShow.ClickedCh {
			if s.OnShow != nil {
				s.OnShow()
			}
		}
	}()

	s.mToggle = systray.AddMenuItemCheckbox(
		"启用/暂停",
		"暂停后将不再修改剪贴板内容",
		s.enabled,
	)
	go func() {
		for range s.mToggle.ClickedCh {
			s.enabled = !s.enabled
			if s.enabled {
				s.mToggle.Check()
			} else {
				s.mToggle.Uncheck()
			}
			if s.OnToggle != nil {
				s.OnToggle(s.enabled)
			}
		}
	}()

	systray.AddSeparator()

	mSettings := systray.AddMenuItem("设置…", "打开设置")
	go func() {
		for range mSettings.ClickedCh {
			if s.OnSettings != nil {
				s.OnSettings()
			}
		}
	}()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("退出", "退出 简繁")
	go func() {
		for range mQuit.ClickedCh {
			if s.OnQuit != nil {
				s.OnQuit()
			}
			systray.Quit()
		}
	}()

	log.Println("✓ 系统托盘已启动")
}

func (s *SystemTray) onExit() {
	// Nothing to clean up — systray owns the icon.
}
