//go:build !windows

package wallpaper

func supported() bool { return false }

func applyHome(path string) error {
	_ = path
	return errUnsupported("home")
}

func applyLock(path string) error {
	_ = path
	return errUnsupported("lock")
}
