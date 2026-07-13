package devops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "opsforall-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	entries, err := ListDir(tempDir)
	if err != nil {
		t.Fatalf("ListDir failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Name != "test.txt" {
		t.Errorf("Expected name test.txt, got %s", entries[0].Name)
	}

	if entries[0].IsBinary {
		t.Error("Text file marked as binary")
	}
}

func TestReadFile_Safety(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "opsforall-read-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test binary detection
	binFile := filepath.Join(tempDir, "test.bin")
	if err := os.WriteFile(binFile, []byte{0, 1, 2, 3, 0, 4}, 0644); err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	content, err := ReadFile(binFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if content != "// [OpsForAll] Binary file content hidden for safety." {
		t.Errorf("Binary file read incorrectly: %q", content)
	}
}
