//go:build windows

package main

import "golang.org/x/sys/windows/svc"

func isWindowsService() bool {
	ok, err := svc.IsWindowsService()
	return err == nil && ok
}
