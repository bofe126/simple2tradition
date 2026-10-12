// Package service contains the clipboard watcher that used to live
// in main.go. It is a small event-driven loop that emits a single
// Event struct on a buffered channel for every clipboard change it
// observes (or skips). The Wails App reads from Events() and forwards
// the events to the frontend via runtime.EventsEmit.
package service

import "time"

// EventKind identifies the kind of clipboard observation.
type EventKind string

const (
	KindConvert EventKind = "convert" // simplified text was converted and written back
	KindSkip    EventKind = "skip"    // clipboard unchanged or no CJK to convert
	KindError   EventKind = "error"   // clipboard read/write failed
)

// SkipReason explains why a KindSkip was emitted. Only meaningful
// when Kind == KindSkip.
type SkipReason string

const (
	SkipEmpty     SkipReason = "empty"     // clipboard was empty
	SkipUnchanged SkipReason = "unchanged" // same as last reading (dedup)
	SkipNonCJK    SkipReason = "non-cjk"   // no CJK characters in the text
	SkipPaused    SkipReason = "paused"    // service is paused via Settings.Enabled
)

// Event is the payload emitted on the service's Events channel and
// forwarded to the frontend. The fields are designed to be JSON-safe
// so the Wails runtime can serialise them without custom marshallers.
type Event struct {
	Kind       EventKind  `json:"kind"`
	When       time.Time  `json:"when"`
	Source     string     `json:"source,omitempty"`      // original clipboard text
	Target     string     `json:"target,omitempty"`      // converted text (only on KindConvert)
	Direction  string     `json:"direction,omitempty"`   // "s2t"/"t2s" (only on KindConvert)
	SkipReason SkipReason `json:"skip_reason,omitempty"` // only on KindSkip
	Error      string     `json:"error,omitempty"`       // only on KindError
}
