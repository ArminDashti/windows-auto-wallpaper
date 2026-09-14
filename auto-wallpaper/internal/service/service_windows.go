//go:build windows

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ArminDashti/auto-wallpaper/internal/config"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func isWindows() bool { return true }

type autoWallService struct {
	cfg config.Config
}

func (s *autoWallService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmds = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmds}

	done := make(chan struct{})
	go func() {
		_ = runWorker(s.cfg)
		close(done)
	}()

	for {
		select {
		case <-done:
			changes <- svc.Status{State: svc.StopPending}
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				return false, 0
			}
		}
	}
}

func runService(cfg config.Config) error {
	elog, err := eventlog.Open(cfg.ServiceName)
	if err == nil {
		defer elog.Close()
		_ = elog.Info(1, "Auto Wallpaper service starting")
	}
	return svc.Run(cfg.ServiceName, &autoWallService{cfg: cfg})
}

func (m *Manager) exePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(exe)
}

func (m *Manager) start() error {
	if err := os.MkdirAll(m.Cfg.DataDir, 0o755); err != nil {
		return err
	}
	if err := writeConfig(m.Cfg); err != nil {
		return err
	}
	exepath, err := m.exePath()
	if err != nil {
		return err
	}
	mgrConn, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect service manager: %w", err)
	}
	defer mgrConn.Disconnect()

	s, err := mgrConn.OpenService(m.Cfg.ServiceName)
	if err != nil {
		s, err = mgrConn.CreateService(
			m.Cfg.ServiceName,
			exepath,
			mgr.Config{
				DisplayName: m.Cfg.DisplayName,
				StartType:   mgr.StartAutomatic,
				Description: "Applies scheduled lock and home wallpapers from auto-wallpaper-api",
			},
			"service",
		)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}
	}
	defer s.Close()
	err = s.Start()
	if err != nil {
		// already running is ok-ish
		return nil
	}
	return nil
}

func (m *Manager) stop() error {
	mgrConn, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer mgrConn.Disconnect()
	s, err := mgrConn.OpenService(m.Cfg.ServiceName)
	if err != nil {
		return fmt.Errorf("service not installed: %w", err)
	}
	defer s.Close()
	status, err := s.Control(svc.Stop)
	if err != nil {
		return err
	}
	timeout := time.Now().Add(20 * time.Second)
	for status.State != svc.Stopped {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout waiting for stop")
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) uninstall() error {
	_ = m.stop()
	mgrConn, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer mgrConn.Disconnect()
	s, err := mgrConn.OpenService(m.Cfg.ServiceName)
	if err != nil {
		// not installed
	} else {
		err = s.Delete()
		s.Close()
		if err != nil {
			return err
		}
	}
	_ = os.RemoveAll(m.Cfg.CacheDir)
	_ = os.Remove(filepath.Join(m.Cfg.DataDir, "config.env"))
	return nil
}

func (m *Manager) status() string {
	mgrConn, err := mgr.Connect()
	if err != nil {
		return "service manager unavailable: " + err.Error()
	}
	defer mgrConn.Disconnect()
	s, err := mgrConn.OpenService(m.Cfg.ServiceName)
	if err != nil {
		return "not installed"
	}
	defer s.Close()
	st, err := s.Query()
	if err != nil {
		return "query failed: " + err.Error()
	}
	switch st.State {
	case svc.Running:
		return "running"
	case svc.Stopped:
		return "stopped"
	case svc.StartPending:
		return "start pending"
	case svc.StopPending:
		return "stop pending"
	default:
		return fmt.Sprintf("state=%d", st.State)
	}
}

func writeConfig(cfg config.Config) error {
	path := filepath.Join(cfg.DataDir, "config.env")
	body := fmt.Sprintf(
		"AUTOWALL_API_URL=%s\nAUTOWALL_USERNAME=%s\nAUTOWALL_PASSWORD=%s\nAUTOWALL_POLL_SECONDS=%d\n",
		cfg.APIURL, cfg.Username, cfg.Password, cfg.PollSeconds,
	)
	if cfg.Token != "" {
		body += "AUTOWALL_TOKEN=" + cfg.Token + "\n"
	}
	return os.WriteFile(path, []byte(body), 0o600)
}
