//go:build !windows

package service

import (
	"fmt"

	"github.com/ArminDashti/auto-wallpaper/internal/config"
)

func isWindows() bool { return false }

func runService(cfg config.Config) error {
	_ = cfg
	return fmt.Errorf("Windows service only supported on Windows")
}

func (m *Manager) start() error {
	return fmt.Errorf("service start only supported on Windows")
}

func (m *Manager) stop() error {
	return fmt.Errorf("service stop only supported on Windows")
}

func (m *Manager) uninstall() error {
	return fmt.Errorf("uninstall only supported on Windows")
}

func (m *Manager) status() string {
	return "not windows"
}
