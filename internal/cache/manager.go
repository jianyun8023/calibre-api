package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/pkg/content"
)

// Config holds cache configuration
type Config struct {
	Dir              string  `mapstructure:"dir"`
	MaxSizeGB        float64 `mapstructure:"max_size_gb"`
	CleanupThreshold float64 `mapstructure:"cleanup_threshold"`
}

// Manager manages EPUB file caching with size limits and automatic cleanup
type Manager struct {
	config     Config
	contentApi *content.Api
	mu         sync.RWMutex
}

// fileInfo holds file metadata for cleanup decisions
type fileInfo struct {
	path       string
	size       int64
	accessTime time.Time
}

// NewManager creates a new cache manager
func NewManager(config Config, contentApi *content.Api) (*Manager, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(config.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Manager{
		config:     config,
		contentApi: contentApi,
	}, nil
}

// GetOrExtractEpub gets the EPUB file path, downloading if necessary
func (cm *Manager) GetOrExtractEpub(bookID string) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	filename := filepath.Join(cm.config.Dir, bookID+".epub")

	// Check if file exists
	if _, err := os.Stat(filename); err == nil {
		// Update access time
		if err := cm.updateAccessTime(filename); err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to update access time for %s: %v\n", filename, err)
		}
		return filename, nil
	}

	// Check and cleanup before downloading
	if err := cm.checkAndCleanup(); err != nil {
		return "", fmt.Errorf("failed to cleanup cache: %w", err)
	}

	// Download the file
	_, reader, err := cm.contentApi.GetBook(bookID, "library")
	if err != nil {
		return "", fmt.Errorf("failed to get book: %w", err)
	}
	defer reader.Close()

	// Create the file
	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file: %w", err)
	}
	defer f.Close()

	// Copy content
	if _, err := io.Copy(f, reader); err != nil {
		os.Remove(filename) // Cleanup on error
		return "", fmt.Errorf("failed to write cache file: %w", err)
	}

	return filename, nil
}

// checkAndCleanup checks cache size and cleans up if necessary
func (cm *Manager) checkAndCleanup() error {
	currentSize, err := cm.getDirSize(cm.config.Dir)
	if err != nil {
		return fmt.Errorf("failed to get cache size: %w", err)
	}

	maxSizeBytes := int64(cm.config.MaxSizeGB * 1024 * 1024 * 1024)
	thresholdBytes := int64(float64(maxSizeBytes) * cm.config.CleanupThreshold)

	// Check if cleanup is needed
	if currentSize < thresholdBytes {
		return nil
	}

	// Get all files with their access times
	files, err := cm.getFilesByAccessTime(cm.config.Dir)
	if err != nil {
		return fmt.Errorf("failed to get files: %w", err)
	}

	// Sort by access time (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].accessTime.Before(files[j].accessTime)
	})

	// Delete files until we're below 70% of max size
	targetSize := int64(float64(maxSizeBytes) * 0.7)
	deletedSize := int64(0)

	for _, file := range files {
		if currentSize-deletedSize <= targetSize {
			break
		}

		if err := os.Remove(file.path); err != nil {
			fmt.Printf("Warning: failed to remove cache file %s: %v\n", file.path, err)
			continue
		}

		deletedSize += file.size
		fmt.Printf("Cleaned up cache file: %s (%.2f MB)\n", filepath.Base(file.path), float64(file.size)/(1024*1024))
	}

	return nil
}

// getDirSize calculates the total size of all files in a directory
func (cm *Manager) getDirSize(dir string) (int64, error) {
	var size int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// getFilesByAccessTime gets all files in directory with their access times
func (cm *Manager) getFilesByAccessTime(dir string) ([]fileInfo, error) {
	var files []fileInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Get access time
		atime := getAccessTime(info)

		files = append(files, fileInfo{
			path:       path,
			size:       info.Size(),
			accessTime: atime,
		})
	}

	return files, nil
}

// updateAccessTime updates the access time of a file
func (cm *Manager) updateAccessTime(path string) error {
	now := time.Now()
	return os.Chtimes(path, now, now)
}

// GetCacheStats returns current cache statistics
func (cm *Manager) GetCacheStats() (map[string]interface{}, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	currentSize, err := cm.getDirSize(cm.config.Dir)
	if err != nil {
		return nil, err
	}

	maxSizeBytes := int64(cm.config.MaxSizeGB * 1024 * 1024 * 1024)
	usagePercent := float64(currentSize) / float64(maxSizeBytes) * 100

	files, err := cm.getFilesByAccessTime(cm.config.Dir)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"current_size_mb": float64(currentSize) / (1024 * 1024),
		"max_size_mb":     float64(maxSizeBytes) / (1024 * 1024),
		"usage_percent":   usagePercent,
		"file_count":      len(files),
		"cache_dir":       cm.config.Dir,
	}, nil
}
