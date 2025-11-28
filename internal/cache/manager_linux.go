//go:build linux
// +build linux

package cache

import (
	"os"
	"syscall"
	"time"
)

// getAccessTime extracts the access time from FileInfo (Linux)
func getAccessTime(info os.FileInfo) time.Time {
	stat := info.Sys().(*syscall.Stat_t)
	// On Linux, use Atim
	return time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
}
