//go:build darwin

// macOS menu-bar (status item) implementation. Uses fyne.io/systray's
// RunWithExternalLoop so the status item coexists with the NSApplication
// run loop that Wails owns.
package tray

/*
#cgo LDFLAGS: -framework Cocoa
void dispatchTrayStart(void);
void installReopenHandler(void);
void setDockIcon(const void* bytes, int len);
*/
import "C"

import (
	_ "embed"
	"log"
	"unsafe"

	"fyne.io/systray"
)

//go:embed icon_s2t.png
var iconS2T []byte

//go:embed icon_t2s.png
var iconT2S []byte

// SystemTray is the menu-bar status item + menu. Create via New,
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

	start func()
	stop  func()

	// Menu-bar state queued until onReady has run.
	started      bool
	pendingTitle string
}

// current is the tray instance whose start function the main-thread
// dispatch (goTrayStart) should invoke. The app only ever creates one.
var current *SystemTray

// New returns a SystemTray that defaults to "enabled" (matches
// config.Default().Enabled).
func New() *SystemTray {
	return &SystemTray{enabled: true}
}

// SetDirection flips the direction-aware branding: Dock icon and
// menu-bar title follow the active conversion direction.
func (s *SystemTray) SetDirection(dir string) {
	icon, title := iconS2T, "简→繁"
	if dir == "t2s" {
		icon, title = iconT2S, "繁→简"
	}
	C.setDockIcon(unsafe.Pointer(&icon[0]), C.int(len(icon)))
	if s.started {
		systray.SetTitle(title)
	} else {
		s.pendingTitle = title
	}
}

// SetEnabled updates the checkbox state to match the live config.
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

// Run registers the status item against the existing (Wails-owned)
// run loop. NSStatusItem may only be created on the main thread, so
// the actual start is dispatched there via dispatchTrayStart.
func (s *SystemTray) Run() {
	start, stop := systray.RunWithExternalLoop(s.onReady, s.onExit)
	s.start = start
	s.stop = stop
	current = s
	C.dispatchTrayStart()
	// Dock icon click → re-show the hidden window (fires OnShow).
	C.installReopenHandler()
}

// Close removes the status item. The App calls this from its
// shutdown hook.
func (s *SystemTray) Close() {
	if s.stop != nil {
		s.stop()
	}
}

func (s *SystemTray) onReady() {
	title := s.pendingTitle
	if title == "" {
		title = "简→繁"
	}
	systray.SetTitle(title)
	systray.SetTooltip("简繁 — 剪贴板简繁转换")
	s.started = true

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

	mQuit := systray.AddMenuItem("退出", "退出 简繁")
	go func() {
		for range mQuit.ClickedCh {
			if s.OnQuit != nil {
				s.OnQuit()
			}
		}
	}()

	log.Println("✓ 菜单栏图标已启动")
}

func (s *SystemTray) onExit() {
	// Nothing to clean up — systray owns the status item.
}
