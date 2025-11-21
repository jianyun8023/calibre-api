//go:build !windows
// +build !windows

package cache

import (
	"os"
	"syscall"
	"time"
)

// getAccessTime extracts the access time from FileInfo (Unix systems)
func getAccessTime(info os.FileInfo) time.Time {
	stat := info.Sys().(*syscall.Stat_t)
	// On Darwin (macOS), use Atimespec
	return time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
}
