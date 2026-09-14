package service

import "github.com/ArminDashti/auto-wallpaper/internal/config"

// Manager installs/controls the Windows service.
type Manager struct {
	Cfg config.Config
}

func New(cfg config.Config) *Manager {
	return &Manager{Cfg: cfg}
}

// Start installs if needed and starts the service.
func (m *Manager) Start() error { return m.start() }

// Stop stops the service.
func (m *Manager) Stop() error { return m.stop() }

// Restart stops then starts.
func (m *Manager) Restart() error {
	_ = m.Stop()
	return m.Start()
}

// Uninstall stops and removes the service.
func (m *Manager) Uninstall() error { return m.uninstall() }

// Status returns a short status string.
func (m *Manager) Status() string { return m.status() }

// IsWindows reports GOOS.
func IsWindows() bool { return isWindows() }

// RunService runs the Windows service main loop (called when started by SCM).
func RunService(cfg config.Config) error { return runService(cfg) }

// RunWorker runs the poll loop in foreground (also used by service).
func RunWorker(cfg config.Config) error { return runWorker(cfg) }
