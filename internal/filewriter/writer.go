package filewriter

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	// MaxSecretSize is the maximum allowed size for secret content (1MB)
	MaxSecretSize = 1 * 1024 * 1024
)

// FileConfig holds file writing configuration
type FileConfig struct {
	Path  string
	Mode  os.FileMode
	Owner int
	Group int
}

// Writer handles atomic file writing
type Writer struct{}

// NewWriter creates a new file writer
func NewWriter() *Writer {
	return &Writer{}
}

// WriteFile writes content to a file atomically
func (w *Writer) WriteFile(config FileConfig, content string) error {
	// Validate content size
	if len(content) > MaxSecretSize {
		return fmt.Errorf("content size %d exceeds maximum allowed size %d", len(content), MaxSecretSize)
	}

	// Validate path for security
	if err := validatePath(config.Path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists and validate it's not a symlink or special file
	if err := validateFileType(config.Path); err != nil {
		return fmt.Errorf("invalid file type: %w", err)
	}

	if err := w.ensureDir(filepath.Dir(config.Path)); err != nil {
		return err
	}

	tmpFile := config.Path + ".tmp." + randomString(8)

	if err := os.WriteFile(tmpFile, []byte(content), config.Mode); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if config.Owner >= 0 || config.Group >= 0 {
		uid := config.Owner
		gid := config.Group
		if uid < 0 {
			uid = -1
		}
		if gid < 0 {
			gid = -1
		}
		if err := os.Chown(tmpFile, uid, gid); err != nil {
			_ = os.Remove(tmpFile)
			return fmt.Errorf("failed to set ownership: %w", err)
		}
	}

	if err := os.Rename(tmpFile, config.Path); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// validatePath checks for path traversal attempts
func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check path length against OS limit
	if len(path) > MaxPathLen {
		return fmt.Errorf("path length %d exceeds maximum %d for this OS", len(path), MaxPathLen)
	}

	// Reject Windows extended paths and UNC paths
	if strings.HasPrefix(path, `\\?\`) || strings.HasPrefix(path, `\\.\`) {
		return fmt.Errorf("extended paths (\\\\?\\) and device paths (\\\\.\\ ) are not allowed")
	}
	if strings.HasPrefix(path, `\\`) {
		return fmt.Errorf("UNC paths (\\\\server\\share) are not allowed, mount the share locally")
	}

	// Ensure path is absolute for security
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute")
	}

	// Check for path traversal attempts before cleaning
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains '..' which is not allowed")
	}

	return nil
}

// validateFileType ensures the path is not a symlink or special file
func validateFileType(path string) error {
	// Check if file exists
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's OK
			return nil
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Reject symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symbolic links are not allowed")
	}

	// Reject special files (devices, pipes, sockets)
	if !info.Mode().IsRegular() && !info.Mode().IsDir() {
		return fmt.Errorf("only regular files are allowed, got: %s", info.Mode().Type())
	}

	return nil
}

func (w *Writer) ensureDir(dir string) error {
	if dir == "" || dir == "." {
		return nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// ParseMode parses a mode string (e.g., "0600") to os.FileMode
func ParseMode(mode string) (os.FileMode, error) {
	if mode == "" {
		return 0600, nil
	}

	m, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid mode: %w", err)
	}

	fileMode := os.FileMode(m)

	// Validate mode is not too permissive
	if err := validateMode(fileMode); err != nil {
		return 0, err
	}

	return fileMode, nil
}

// validateMode ensures file mode is not overly permissive
func validateMode(mode os.FileMode) error {
	// Extract permission bits (ignore file type bits)
	perm := mode & os.ModePerm

	// Check if world-writable (other write bit set)
	if perm&0002 != 0 {
		return fmt.Errorf("world-writable permissions (0%o) are not allowed", perm)
	}

	// Check if group-writable and world-readable (too permissive for secrets)
	if perm&0020 != 0 && perm&0004 != 0 {
		return fmt.Errorf("group-writable with world-readable (0%o) is too permissive", perm)
	}

	// Warn if more permissive than 0644
	if perm > 0644 {
		return fmt.Errorf("permissions (0%o) are too permissive, maximum is 0644", perm)
	}

	return nil
}

// ParseOwner parses owner/group string to int
func ParseOwner(owner string) (int, error) {
	if owner == "" {
		return -1, nil
	}

	o, err := strconv.Atoi(owner)
	if err != nil {
		return -1, fmt.Errorf("invalid owner: %w", err)
	}

	return o, nil
}

// GetFileInfo returns file information
func GetFileInfo(path string) (os.FileMode, int, int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, -1, -1, err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return info.Mode(), -1, -1, nil
	}

	return info.Mode(), int(stat.Uid), int(stat.Gid), nil
}

// randomString generates a random string of length n for temp file names
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp if random fails
		return fmt.Sprintf("%d", os.Getpid())
	}
	for i := range b {
		b[i] = letters[b[i]%byte(len(letters))]
	}
	return string(b)
}

