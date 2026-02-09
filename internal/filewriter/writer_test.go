package filewriter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	writer := NewWriter()
	config := FileConfig{
		Path:  filePath,
		Mode:  0644,
		Owner: -1,
		Group: -1,
	}

	content := "test content"
	if err := writer.WriteFile(config, content); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected '%s', got '%s'", content, string(data))
	}
}

func TestWriteFile_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "subdir", "test.txt")

	writer := NewWriter()
	config := FileConfig{
		Path:  filePath,
		Mode:  0600,
		Owner: -1,
		Group: -1,
	}

	if err := writer.WriteFile(config, "content"); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("file was not created")
	}
}

func TestWriteFile_Permissions(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	writer := NewWriter()
	config := FileConfig{
		Path:  filePath,
		Mode:  0600,
		Owner: -1,
		Group: -1,
	}

	if err := writer.WriteFile(config, "content"); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode 0600, got %o", info.Mode().Perm())
	}
}

func TestWriteFile_Atomic(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	_ = os.WriteFile(filePath, []byte("original"), 0644)

	writer := NewWriter()
	config := FileConfig{
		Path:  filePath,
		Mode:  0644,
		Owner: -1,
		Group: -1,
	}

	if err := writer.WriteFile(config, "updated"); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != "updated" {
		t.Errorf("expected 'updated', got '%s'", string(data))
	}

	// Check no .tmp.* files left behind
	tmpPattern := filePath + ".tmp.*"
	matches, _ := filepath.Glob(tmpPattern)
	if len(matches) > 0 {
		t.Error("temp file was not cleaned up")
	}
}

func TestWriteFile_RejectsSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	symlinkFile := filepath.Join(tmpDir, "symlink.txt")

	// Create target file
	_ = os.WriteFile(targetFile, []byte("target"), 0644)

	// Create symlink
	if err := os.Symlink(targetFile, symlinkFile); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	writer := NewWriter()
	config := FileConfig{
		Path:  symlinkFile,
		Mode:  0644,
		Owner: -1,
		Group: -1,
	}

	err := writer.WriteFile(config, "content")
	if err == nil {
		t.Fatal("expected error for symlink, got nil")
	}
	if !contains(err.Error(), "symbolic link") {
		t.Errorf("expected symlink error, got: %v", err)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestWriteFile_RejectsLargeContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large.txt")

	writer := NewWriter()
	config := FileConfig{
		Path:  filePath,
		Mode:  0644,
		Owner: -1,
		Group: -1,
	}

	// Create content larger than MaxSecretSize
	largeContent := string(make([]byte, MaxSecretSize+1))

	err := writer.WriteFile(config, largeContent)
	if err == nil {
		t.Fatal("expected error for large content, got nil")
	}
	if !contains(err.Error(), "exceeds maximum") {
		t.Errorf("expected size error, got: %v", err)
	}
}

func TestWriteFile_RandomTempNames(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	writer := NewWriter()
	config := FileConfig{
		Path:  filePath,
		Mode:  0644,
		Owner: -1,
		Group: -1,
	}

	// Write multiple times and verify no temp files remain
	for i := 0; i < 5; i++ {
		if err := writer.WriteFile(config, fmt.Sprintf("content-%d", i)); err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}
	}

	// Check no .tmp.* files left behind
	pattern := filepath.Join(tmpDir, "*.tmp.*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(matches) > 0 {
		t.Errorf("found %d orphaned temp files: %v", len(matches), matches)
	}
}

func TestParseMode_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected os.FileMode
	}{
		{"0600", 0600},
		{"0644", 0644},
		{"", 0600},
	}

	for _, tt := range tests {
		mode, err := ParseMode(tt.input)
		if err != nil {
			t.Errorf("ParseMode(%s) failed: %v", tt.input, err)
		}
		if mode != tt.expected {
			t.Errorf("ParseMode(%s) = %o, expected %o", tt.input, mode, tt.expected)
		}
	}
}

func TestParseMode_Invalid(t *testing.T) {
	tests := []string{
		"invalid",
		"0777", // world-writable
		"0666", // world-writable
		"0755", // too permissive
	}

	for _, input := range tests {
		_, err := ParseMode(input)
		if err == nil {
			t.Errorf("expected error for mode %s, got nil", input)
		}
	}
}

func TestParseOwner_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1000", 1000},
		{"0", 0},
		{"", -1},
	}

	for _, tt := range tests {
		owner, err := ParseOwner(tt.input)
		if err != nil {
			t.Errorf("ParseOwner(%s) failed: %v", tt.input, err)
		}
		if owner != tt.expected {
			t.Errorf("ParseOwner(%s) = %d, expected %d", tt.input, owner, tt.expected)
		}
	}
}

func TestParseOwner_Invalid(t *testing.T) {
	_, err := ParseOwner("invalid")
	if err == nil {
		t.Error("expected error for invalid owner, got nil")
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid absolute path", "/tmp/secret.txt", false},
		{"valid nested path", "/var/lib/secrets/db.txt", false},
		{"empty path", "", true},
		{"relative path", "secret.txt", true},
		{"relative with dot", "./secret.txt", true},
		{"path traversal", "/tmp/../etc/passwd", true},
		{"path traversal nested", "/var/lib/../../etc/passwd", true},
		{"too long path", "/" + string(make([]byte, MaxPathLen)), true},
		{"windows extended path", `\\?\C:\secrets\test`, true},
		{"windows device path", `\\.\pipe\test`, true},
		{"windows UNC path", `\\server\share\secret`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath(%s) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestGetFileInfo(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	_ = os.WriteFile(filePath, []byte("content"), 0644)

	mode, uid, gid, err := GetFileInfo(filePath)
	if err != nil {
		t.Fatalf("failed to get file info: %v", err)
	}

	if mode.Perm() != 0644 {
		t.Errorf("expected mode 0644, got %o", mode.Perm())
	}

	if uid < 0 {
		t.Error("expected valid uid")
	}

	if gid < 0 {
		t.Error("expected valid gid")
	}
}

func TestGetFileInfo_NonExistent(t *testing.T) {
	_, _, _, err := GetFileInfo("/nonexistent/file")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}
