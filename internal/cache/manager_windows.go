//go:build windows
// +build windows

package cache

import (
	"os"
	"syscall"
	"time"
)

// getAccessTime extracts the access time from FileInfo (Windows)
func getAccessTime(info os.FileInfo) time.Time {
	stat := info.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, stat.LastAccessTime.Nanoseconds())
}
