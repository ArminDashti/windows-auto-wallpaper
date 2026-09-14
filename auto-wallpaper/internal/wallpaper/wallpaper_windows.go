//go:build windows

package wallpaper

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

func supported() bool { return true }

func applyHome(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	user, err := syscall.UTF16PtrFromString(abs)
	if err != nil {
		return err
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("SystemParametersInfoW")
	const SPI_SETDESKWALLPAPER = 0x0014
	const SPIF_UPDATEINIFILE = 0x01
	const SPIF_SENDCHANGE = 0x02
	r, _, e := proc.Call(
		uintptr(SPI_SETDESKWALLPAPER),
		0,
		uintptr(unsafe.Pointer(user)),
		uintptr(SPIF_UPDATEINIFILE|SPIF_SENDCHANGE),
	)
	if r == 0 {
		if e != nil {
			return fmt.Errorf("SystemParametersInfoW: %w", e)
		}
		return fmt.Errorf("SystemParametersInfoW failed")
	}
	return nil
}

func applyLock(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	key, _, err := registry.CreateKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\PersonalizationCSP`,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("open PersonalizationCSP: %w", err)
	}
	defer key.Close()
	if err := key.SetStringValue("LockScreenImagePath", abs); err != nil {
		return err
	}
	if err := key.SetStringValue("LockScreenImageUrl", abs); err != nil {
		return err
	}
	return key.SetDWordValue("LockScreenImageStatus", 1)
}
