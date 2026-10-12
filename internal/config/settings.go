// Package config defines persisted user settings for simple2tradition.
//
// Settings are stored as JSON in the platform's user config directory:
//   - Windows: %APPDATA%\simple2tradition\settings.json
//   - Other (fallback): $HOME/.simple2tradition/settings.json
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Conversion directions.
const (
	DirS2T = "s2t" // simplified → traditional
	DirT2S = "t2s" // traditional → simplified
)

// Settings is the in-memory representation of the persisted config.
type Settings struct {
	PollIntervalMs   int    `json:"poll_interval_ms"`
	Enabled          bool   `json:"enabled"`
	Direction        string `json:"direction"` // DirS2T or DirT2S
	StartWithWindows bool   `json:"start_with_windows"`
	StartMinimized   bool   `json:"start_minimized"`
	WriteBackToClip  bool   `json:"write_back_to_clipboard"`
}

// Default returns the factory defaults used on first run or after a
// corrupt/missing config file.
func Default() Settings {
	return Settings{
		PollIntervalMs:   500,
		Enabled:          true,
		Direction:        DirS2T,
		StartWithWindows: false,
		StartMinimized:   false,
		WriteBackToClip:  true,
	}
}

// Validate clamps numeric fields to safe ranges. Always succeeds;
// returns nil because clamping is the primary repair strategy.
func (s *Settings) Validate() error {
	if s.PollIntervalMs < 100 {
		s.PollIntervalMs = 100
	}
	if s.PollIntervalMs > 5000 {
		s.PollIntervalMs = 5000
	}
	if s.Direction != DirT2S {
		s.Direction = DirS2T
	}
	return nil
}

// ConfigDir returns the directory where settings.json lives. Created
// on demand by callers. Safe to call repeatedly.
func ConfigDir() string {
	if app := os.Getenv("APPDATA"); app != "" {
		return filepath.Join(app, "simple2tradition")
	}
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".simple2tradition")
	}
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}

// settingsPath is the resolved path to settings.json. Exposed for tests.
func settingsPath() string {
	return filepath.Join(ConfigDir(), "settings.json")
}

// Store wraps Settings with disk persistence and a mutex so it can be
// safely accessed from goroutines and from the Wails-bound methods
// (which are called on the WebView's message loop).
type Store struct {
	mu   sync.RWMutex
	data Settings
	path string
}

// NewStore loads settings from disk (or defaults if the file is
// missing/corrupt) and returns a ready-to-use Store.
func NewStore() (*Store, error) {
	s := &Store{path: settingsPath()}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return nil, err
	}
	s.data = loadFromDisk(s.path)
	s.data.Validate()
	return s, nil
}

// Get returns a copy of the current settings (safe for the caller to
// keep or mutate without affecting the store).
func (s *Store) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// Replace swaps the in-memory settings with the provided value (after
// validation) and writes the result to disk atomically. Returns the
// validated value (which may differ from the input after clamping).
func (s *Store) Replace(in Settings) (Settings, error) {
	if in.PollIntervalMs == 0 {
		// Treat zero as "leave alone" for backwards compat with
		// clients that send partial updates; otherwise clamp.
		cur := s.Get()
		in.PollIntervalMs = cur.PollIntervalMs
	}
	if err := in.Validate(); err != nil {
		return Settings{}, err
	}
	s.mu.Lock()
	s.data = in
	s.mu.Unlock()
	if err := s.save(); err != nil {
		return Settings{}, err
	}
	return in, nil
}

// SetEnabled updates only the Enabled flag in-place (used by the
// tray toggle without re-serialising the rest of the config). Does
// not write to disk — the next Replace will persist it.
func (s *Store) SetEnabled(v bool) {
	s.mu.Lock()
	s.data.Enabled = v
	s.mu.Unlock()
}

// save writes the current settings to disk atomically (temp file +
// rename) so a crash mid-write cannot leave a half-written config.
func (s *Store) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// loadFromDisk reads the file at path and returns Settings, falling
// back to Default() on any error (missing, permission, corrupt JSON).
func loadFromDisk(path string) Settings {
	data, err := os.ReadFile(path)
	if err != nil {
		return Default()
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Default()
	}
	return s
}

// ErrInvalid is returned when a Replace call provides out-of-range
// values that cannot be clamped to a valid state.
var ErrInvalid = errors.New("invalid settings")
