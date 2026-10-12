//go:build !windows && !darwin

// Stub so the package compiles on non-Windows platforms (used by
// `go vet ./...` on Mac/Linux dev machines and by cross-compilation
// in CI). The actual tray implementation lives in tray.go.
package tray

// SystemTray is a no-op stand-in for non-Windows builds. It exposes
// the same callback fields as the Windows version so callers can
// wire them up uniformly; the callbacks are simply never invoked.
type SystemTray struct {
	OnShow     func()
	OnSettings func()
	OnToggle   func(enabled bool)
	OnQuit     func()
}

func New() *SystemTray                        { return &SystemTray{} }
func (s *SystemTray) Run()                    {}
func (s *SystemTray) Close()                  {}
func (s *SystemTray) SetEnabled(v bool)       {}
func (s *SystemTray) SetDirection(dir string) {}
