package devops

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileEntry represents a file or directory entry.
type FileEntry struct {
	Name    string
	Path    string
	Size    string
	IsDir   bool
	Mode    os.FileMode
	ModTime time.Time
}

// ListDir lists the contents of a directory.
func ListDir(path string) ([]FileEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []FileEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileEntry{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			Size:    formatSize(info.Size()),
			IsDir:   info.IsDir(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
		})
	}

	return files, nil
}

// ReadFile reads the entire contents of a file.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatSize converts bytes to a human-readable string.
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
