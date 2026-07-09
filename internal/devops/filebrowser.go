package devops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileEntry represents a file or directory entry.
type FileEntry struct {
	Name     string
	Path     string
	Size     string
	RawSize  int64
	IsDir    bool
	IsBinary bool
	Mode     os.FileMode
	ModTime  time.Time
}

// isPathSafe checks that the path does not escape the sandbox.
// Absolute paths are allowed (the frontend file browser sends them),
// but directory traversal is blocked.
func isPathSafe(path string) error {
	if path == "" || path == "." {
		return nil
	}
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		// For relative paths, block traversal
		if strings.HasPrefix(clean, "..") || strings.Contains(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("directory traversal is not allowed")
		}
	}
	return nil
}

// ListDir lists the contents of a directory.
func ListDir(path string) ([]FileEntry, error) {
	if path == "" || path == "." {
		var err error
		path, err = os.UserHomeDir()
		if err != nil {
			path = os.TempDir()
		}
	}
	if err := isPathSafe(path); err != nil {
		return nil, err
	}
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
			Name:     entry.Name(),
			Path:     filepath.Join(path, entry.Name()),
			Size:     formatSize(info.Size()),
			RawSize:  info.Size(),
			IsDir:    info.IsDir(),
			IsBinary: isBinaryFile(filepath.Join(path, entry.Name()), info),
			Mode:     info.Mode(),
			ModTime:  info.ModTime(),
		})
	}

	return files, nil
}

// isBinaryFile checks if a file is binary by looking at its extension and a small sample.
func isBinaryFile(path string, info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}
	// Common binary extensions
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".obj": true, ".o": true, ".a": true, ".lib": true,
		".zip": true, ".tar": true, ".gz": true, ".7z": true, ".rar": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".pdf": true,
		".db": true, ".sqlite": true, ".dat": true,
	}
	if binaryExts[ext] {
		return true
	}

	// Read small sample if file isn't huge
	if info.Size() == 0 {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return true // treat unreadable as binary for safety
	}
	defer f.Close()

	sample := make([]byte, 512)
	n, _ := f.Read(sample)
	for i := 0; i < n; i++ {
		if sample[i] == 0 {
			return true
		}
	}
	return false
}

// ReadFile reads the first 1MB of a file's contents if it's not binary.
func ReadFile(path string) (string, error) {
	if err := isPathSafe(path); err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if isBinaryFile(path, info) {
		return "// [Hawkward] Binary file content hidden for safety.", nil
	}

	// Limit read to 1MB
	maxSize := int64(1 * 1024 * 1024)
	readSize := info.Size()
	if readSize > maxSize {
		readSize = maxSize
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data := make([]byte, readSize)
	_, err = f.Read(data)
	if err != nil {
		return "", err
	}

	content := string(data)
	if info.Size() > maxSize {
		content += "\n\n// [Hawkward] File truncated - viewing first 1MB."
	}
	return content, nil
}

// WriteFile writes data to a file.
func WriteFile(path string, data string) error {
	if err := isPathSafe(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(data), 0600)
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
