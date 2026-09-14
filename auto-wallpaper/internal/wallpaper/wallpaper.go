package wallpaper

import "fmt"

// ApplyHome sets the desktop wallpaper to path.
func ApplyHome(path string) error {
	return applyHome(path)
}

// ApplyLock sets the lock screen image to path (elevated).
func ApplyLock(path string) error {
	return applyLock(path)
}

// Supported reports whether wallpaper APIs are available on this OS.
func Supported() bool {
	return supported()
}

func errUnsupported(kind string) error {
	return fmt.Errorf("%s wallpaper apply only supported on Windows", kind)
}
