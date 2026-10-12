package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/siongui/gojianfan"
)

// Snapshot is the subset of live settings the polling loop needs
// each tick.
type Snapshot struct {
	PollIntervalMs int
	Enabled        bool
	Direction      string // "s2t" or "t2s"
	WriteBack      bool   // false = preview-only, don't touch the clipboard
}

// SettingsProvider is the read/write surface the service needs from
// the config layer. Decoupling the service from *config.Store lets
// tests inject a fake without touching disk.
type SettingsProvider interface {
	Get() Snapshot
	SetEnabled(bool)
}

// Config adapts an arbitrary snapshot getter to the SettingsProvider
// interface. Callers supply a Snapshot closure so this package does
// not need to import the config package.
type Config struct {
	Snapshot func() Snapshot
	Set      func(enabled bool)
}

func (c Config) Get() Snapshot     { return c.Snapshot() }
func (c Config) SetEnabled(b bool) { c.Set(b) }

// Service polls the OS clipboard at the configured interval and
// converts simplified Chinese to traditional Chinese. One Service
// runs a single goroutine; multiple readers may consume from Events.
type Service struct {
	cfg     SettingsProvider
	events  chan Event
	counter uint64 // total successful conversions since Start

	mu      sync.Mutex
	running bool
}

// New returns a Service ready to start. eventsBuf is the channel
// buffer; 50 is plenty for a desktop UI consumer and prevents the
// goroutine from blocking if the WebView is briefly slow.
func New(cfg SettingsProvider, eventsBuf int) *Service {
	if eventsBuf <= 0 {
		eventsBuf = 50
	}
	return &Service{
		cfg:    cfg,
		events: make(chan Event, eventsBuf),
	}
}

// Events returns the read-only event channel. The channel is closed
// when the service context is cancelled.
func (s *Service) Events() <-chan Event { return s.events }

// Counter returns the total number of successful conversions since
// the last Start. Thread-safe.
func (s *Service) Counter() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.counter
}

// ResetCounter zeros the conversion counter. Used by the UI's
// "Reset" button.
func (s *Service) ResetCounter() {
	s.mu.Lock()
	s.counter = 0
	s.mu.Unlock()
}

// Start blocks until ctx is cancelled. Spawned as a goroutine by
// the App on startup.
func (s *Service) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	defer close(s.events)
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	var lastText string
	// Start with a 500ms default; immediately overwritten from cfg.
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			// Check current settings; this is cheap (mutex read).
			snap := s.cfg.Get()

			text, err := clipboard.ReadAll()
			if err != nil {
				// Skip transient clipboard-busy conditions (e.g. the
				// clipboard is briefly held by another app). These
				// are not real errors and would otherwise spam the
				// event stream every poll cycle.
				if isTransientClipboardError(err) {
					continue
				}
				s.emit(Event{
					Kind:  KindError,
					When:  time.Now(),
					Error: err.Error(),
				})
				continue
			}

			if !snap.Enabled {
				// Only emit a SkipPaused the first time the text
				// changes while paused; otherwise stay quiet to
				// avoid log spam.
				if text != lastText {
					lastText = text
					s.emit(Event{
						Kind:       KindSkip,
						When:       time.Now(),
						Source:     text,
						SkipReason: SkipPaused,
					})
				}
				continue
			}

			if text == "" {
				continue
			}
			if text == lastText {
				continue
			}
			lastText = text

			if !containsCJK(text) {
				s.emit(Event{
					Kind:       KindSkip,
					When:       time.Now(),
					Source:     text,
					SkipReason: SkipNonCJK,
				})
				continue
			}

			var converted string
			if snap.Direction == "t2s" {
				converted = gojianfan.T2S(text)
			} else {
				converted = gojianfan.S2T(text)
			}
			if converted == text {
				// Already in the target script; nothing to do.
				continue
			}

			if snap.WriteBack {
				if err := clipboard.WriteAll(converted); err != nil {
					if isTransientClipboardError(err) {
						// Same transient handling as ReadAll above.
						continue
					}
					s.emit(Event{
						Kind:   KindError,
						When:   time.Now(),
						Source: text,
						Error:  err.Error(),
					})
					continue
				}
				// Remember the written text so the next poll doesn't
				// re-process our own output.
				lastText = converted
			}

			s.mu.Lock()
			s.counter++
			s.mu.Unlock()

			s.emit(Event{
				Kind:      KindConvert,
				When:      time.Now(),
				Source:    text,
				Target:    converted,
				Direction: snap.Direction,
			})

			// Reset the ticker to honour the (possibly changed)
			// poll interval from the live config. Resetting is
			// cheaper than tearing down a goroutine.
			if d := time.Duration(snap.PollIntervalMs) * time.Millisecond; d > 0 {
				ticker.Reset(d)
			}
		}
	}
}

// emit drops the event if the channel buffer is full (so the
// clipboard goroutine never blocks), but logs a counter to help
// detect buffer overflow in development.
func (s *Service) emit(e Event) {
	select {
	case s.events <- e:
	default:
		// Buffer full: drop the event silently. The UI will catch
		// up on the next non-dropped event.
	}
}

// containsCJK reports whether s contains at least one character in
// the CJK Unified Ideographs block. This is a fast pre-check; the
// real conversion handles simplified-vs-traditional disambiguation.
func containsCJK(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

// isTransientClipboardError reports whether err is a clipboard-busy
// condition that should be retried on the next poll rather than
// surfaced to the user. The Windows clipboard APIs sometimes report
// ERROR_SUCCESS ("The operation completed successfully") as an error
// when the clipboard is briefly held by another process; same for
// ERROR_ACCESS_DENIED. Both cases are normal during heavy clipboard
// use and should not flood the history with error entries.
func isTransientClipboardError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "completed successfully"),
		strings.Contains(msg, "Access is denied"),
		strings.Contains(msg, "拒绝访问"):
		return true
	}
	return false
}
