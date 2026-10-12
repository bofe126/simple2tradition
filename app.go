package main

import (
	"context"
	"fmt"
	goruntime "runtime"
	"sync"
	"time"

	"simple2tradition/internal/config"
	"simple2tradition/internal/service"
	"simple2tradition/internal/tray"

	"github.com/atotto/clipboard"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Status is the payload returned by App.GetStatus. It is a small
// snapshot used by the frontend's Status tab to render without
// subscribing to every event.
type Status struct {
	Enabled       bool           `json:"enabled"`
	Running       bool           `json:"running"`
	Counter       uint64         `json:"counter"`
	ClipboardHead string         `json:"clipboard_head"`
	LastEvent     *service.Event `json:"last_event,omitempty"`
	LastUpdated   time.Time      `json:"last_updated"`
}

// App is the Wails-bound application. Its exported methods are
// auto-generated as JS bindings under frontend/wailsjs/go/main/App.js
// by `wails dev` / `wails build`.
type App struct {
	ctx context.Context

	cfg      *config.Store
	svc      *service.Service
	trayIcon *tray.SystemTray

	mu        sync.Mutex
	quitting  bool // true once the user picked "退出" from the tray
	lastEvent *service.Event
}

// NewApp returns an App with its config store loaded from disk.
// The service and tray are created in startup().
func NewApp() *App {
	store, err := config.NewStore()
	if err != nil {
		// If the config dir is somehow unwritable, fall back to
		// in-memory defaults so the app still launches. The user
		// can re-attempt by restarting with a fixed HOME/APPDATA.
		fmt.Println("⚠ 无法初始化配置存储，使用默认设置:", err)
		store = mustInMemoryStore()
	}
	return &App{cfg: store}
}

// startup is wired into options.App.OnStartup. It must be non-blocking:
// long-running work is dispatched to goroutines.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Adapter so the service can read live settings without a
	// direct dependency on *config.Store.
	prov := service.Config{
		Snapshot: func() service.Snapshot {
			s := a.cfg.Get()
			return service.Snapshot{
				PollIntervalMs: s.PollIntervalMs,
				Enabled:        s.Enabled,
				Direction:      s.Direction,
				WriteBack:      s.WriteBackToClip,
			}
		},
		Set: func(b bool) { a.cfg.SetEnabled(b) },
	}

	// Respect "start minimized" on every platform: macOS starts
	// visible (hide if requested), Windows starts hidden via
	// StartHidden (show unless requested).
	if a.cfg.Get().StartMinimized {
		wailsruntime.WindowHide(ctx)
	} else if goruntime.GOOS == "windows" {
		wailsruntime.WindowShow(ctx)
	}

	a.svc = service.New(prov, 50)

	// Initialize the tray first so the menu is available even if
	// the service is slow to start.
	a.initTray()

	// Service goroutine (blocking) and bridge goroutine (drains
	// events to the WebView).
	go a.svc.Start(ctx)
	go a.bridgeEvents()
}

// shutdown is wired into options.App.OnShutdown. Best-effort cleanup.
func (a *App) shutdown(ctx context.Context) {
	if a.trayIcon != nil {
		a.trayIcon.Close()
	}
	// service goroutine exits when its ctx is cancelled by Wails.
}

// beforeClose is wired into options.App.OnBeforeClose. Returns true
// to prevent the close (we hide to tray instead) unless the user
// explicitly chose Quit.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	a.mu.Lock()
	quitting := a.quitting
	a.mu.Unlock()
	if quitting {
		return false // allow the app to exit
	}
	// Closing the window hides it; the app keeps converting in the
	// background. Quit via the tray (Windows) or Cmd+Q (macOS).
	wailsruntime.WindowHide(ctx)
	return true
}

// bridgeEvents drains service events and re-emits them to the
// frontend via the Wails runtime. Also caches the most recent event
// so GetStatus has something to return without subscribing.
func (a *App) bridgeEvents() {
	if a.ctx == nil {
		return
	}
	for ev := range a.svc.Events() {
		a.mu.Lock()
		a.lastEvent = &ev
		a.mu.Unlock()

		// Emit a copy (already JSON-safe) under a namespaced name.
		wailsruntime.EventsEmit(a.ctx, "clipboard:"+string(ev.Kind), ev)
	}
}

// initTray wires the systray callbacks. The tray runs on its own
// goroutine via energye/systray.
func (a *App) initTray() {
	a.trayIcon = tray.New()
	cur := a.cfg.Get()
	a.trayIcon.SetEnabled(cur.Enabled)
	a.trayIcon.SetDirection(cur.Direction)

	a.trayIcon.OnShow = func() {
		wailsruntime.WindowShow(a.ctx)
	}
	a.trayIcon.OnSettings = func() {
		wailsruntime.WindowShow(a.ctx)
	}
	a.trayIcon.OnToggle = func(enabled bool) {
		a.cfg.SetEnabled(enabled)
		wailsruntime.EventsEmit(a.ctx, "settings:enabled", enabled)
	}
	a.trayIcon.OnQuit = func() {
		a.mu.Lock()
		a.quitting = true
		a.mu.Unlock()
		wailsruntime.Quit(a.ctx)
	}
	go a.trayIcon.Run()
}

// ======== Wails-bound methods (auto-exposed to JS) ========

// GetSettings returns the current persisted settings.
func (a *App) GetSettings() config.Settings {
	return a.cfg.Get()
}

// SaveSettings validates and persists the provided settings, then
// emits settings:changed so the UI updates without re-fetching.
func (a *App) SaveSettings(in config.Settings) (config.Settings, error) {
	saved, err := a.cfg.Replace(in)
	if err != nil {
		return config.Settings{}, err
	}
	// Reflect Enabled/Direction changes into the tray and Dock
	// branding immediately.
	if a.trayIcon != nil {
		a.trayIcon.SetEnabled(saved.Enabled)
		a.trayIcon.SetDirection(saved.Direction)
	}
	wailsruntime.EventsEmit(a.ctx, "settings:changed", saved)
	return saved, nil
}

// GetStatus returns a UI snapshot. Cheap to call.
func (a *App) GetStatus() Status {
	s := a.cfg.Get()
	head := ""
	if text, err := clipboard.ReadAll(); err == nil {
		head = truncate(text, 50)
	}
	a.mu.Lock()
	last := a.lastEvent
	a.mu.Unlock()
	return Status{
		Enabled:       s.Enabled,
		Running:       a.svc != nil,
		Counter:       a.svc.Counter(),
		ClipboardHead: head,
		LastEvent:     last,
		LastUpdated:   time.Now(),
	}
}

// ResetCounter zeros the lifetime conversion counter (the value is
// in-memory only; the next Start will reset it again).
func (a *App) ResetCounter() {
	if a.svc == nil {
		return
	}
	a.svc.ResetCounter()
	wailsruntime.EventsEmit(a.ctx, "counter:reset", nil)
}

// RequestQuit marks the app as quitting and exits. Wired to the
// macOS app menu's Cmd+Q (there is no tray on macOS), so the quit
// is not swallowed by beforeClose's hide-to-background behaviour.
func (a *App) RequestQuit() {
	a.mu.Lock()
	a.quitting = true
	a.mu.Unlock()
	wailsruntime.Quit(a.ctx)
}

// ShowWindow is a convenience bound method for the frontend (e.g. a
// "open settings" button that re-uses the same path as the tray).
func (a *App) ShowWindow() {
	wailsruntime.WindowShow(a.ctx)
}

// HideWindow lets the frontend programmatically hide the window.
func (a *App) HideWindow() {
	wailsruntime.WindowHide(a.ctx)
}

// truncate returns the first maxRunes of s with "..." appended if cut.
func truncate(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes]) + "..."
}

// mustInMemoryStore returns a Store preloaded with defaults and a
// temp file path. Used as a last-resort fallback when the real
// config dir is unwritable so the app can still launch.
func mustInMemoryStore() *config.Store {
	s, err := config.NewStore()
	if err != nil {
		// NewStore only fails on mkdir; in practice the user can
		// still tweak in-memory settings even if persistence is
		// broken. We swallow the error and return a zero-value
		// store to avoid panicking.
		fmt.Println("⚠ 回退到内存配置（不会持久化）")
	}
	if s == nil {
		return &config.Store{}
	}
	return s
}
