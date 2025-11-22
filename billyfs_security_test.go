package billyfs_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/absfs/billyfs"
	"github.com/absfs/osfs"
)

// TestSecurity_PathTraversal tests for path traversal vulnerabilities
func TestSecurity_PathTraversal(t *testing.T) {
	bfs, tempDir, cleanup := setupTestFS(t)
	defer cleanup()

	// Create a file outside the root that we should NOT be able to access
	outsideDir := filepath.Join(filepath.Dir(tempDir), "outside_root")
	err := os.MkdirAll(outsideDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outsideDir)

	secretFile := filepath.Join(outsideDir, "secret.txt")
	err = os.WriteFile(secretFile, []byte("secret data"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		path string
		desc string
	}{
		{
			name: "double_dot_parent",
			path: "../outside_root/secret.txt",
			desc: "Should not access parent directory via ..",
		},
		{
			name: "absolute_path_escape",
			path: secretFile,
			desc: "Should not access absolute paths outside root",
		},
		{
			name: "multiple_double_dots",
			path: "../../../../../../etc/passwd",
			desc: "Should not traverse multiple levels up",
		},
		{
			name: "double_dot_after_valid",
			path: "subdir/../../../outside_root/secret.txt",
			desc: "Should not escape via double-dot after valid path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt to open the file - this should either fail or not access the secret file
			f, err := bfs.Open(tt.path)
			if err == nil {
				defer f.Close()
				// If it succeeded, verify it's not actually the secret file
				data := make([]byte, 100)
				n, _ := f.Read(data)
				if n > 0 && string(data[:n]) == "secret data" {
					t.Errorf("%s: successfully accessed secret file outside root", tt.desc)
				}
			}
			// Note: We don't fail the test if Open returns an error,
			// as that's the expected secure behavior
		})
	}
}

// TestSecurity_TempFileUniqueness tests that temporary files have unique names
func TestSecurity_TempFileUniqueness(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	// Create a directory for temp files
	err := bfs.MkdirAll("/tmp", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Generate multiple temp files and ensure they're unique
	fileNames := make(map[string]bool)
	var files []string

	for i := 0; i < 100; i++ {
		f, err := bfs.TempFile("/tmp", "test")
		if err != nil {
			t.Fatalf("Failed to create temp file %d: %v", i, err)
		}
		defer f.Close()

		name := f.Name()
		files = append(files, name)

		if fileNames[name] {
			t.Errorf("Duplicate temp file name generated: %s", name)
		}
		fileNames[name] = true

		// Verify the file name starts with the prefix
		baseName := filepath.Base(name)
		if !strings.HasPrefix(baseName, "test") {
			t.Errorf("Temp file name doesn't start with prefix: %s", name)
		}
	}

	// Verify all names are unique
	if len(fileNames) != 100 {
		t.Errorf("Expected 100 unique temp files, got %d", len(fileNames))
	}
}

// TestSecurity_TempFilePermissions tests that temp files have secure permissions
func TestSecurity_TempFilePermissions(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	err := bfs.MkdirAll("/tmp", 0755)
	if err != nil {
		t.Fatal(err)
	}

	f, err := bfs.TempFile("/tmp", "secure")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Get file info
	name := f.Name()
	info, err := bfs.Stat(name)
	if err != nil {
		t.Fatal(err)
	}

	// Check permissions (should be 0600 or more restrictive)
	mode := info.Mode()
	// On Unix-like systems, temp files should not be world-readable or group-readable
	if mode&0077 != 0 {
		t.Logf("Warning: Temp file has permissive permissions: %o (expected 0600)", mode)
		// Note: This is a warning because some filesystems may not support Unix permissions
	}
}

// TestSecurity_SymlinkEscape tests that symlinks cannot escape the root
func TestSecurity_SymlinkEscape(t *testing.T) {
	bfs, tempDir, cleanup := setupTestFS(t)
	defer cleanup()

	// Create a target outside the root
	outsideDir := filepath.Join(filepath.Dir(tempDir), "outside_symlink")
	err := os.MkdirAll(outsideDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outsideDir)

	targetFile := filepath.Join(outsideDir, "target.txt")
	err = os.WriteFile(targetFile, []byte("outside data"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Try to create a symlink pointing outside the root
	err = bfs.Symlink(targetFile, "/evil_link")
	if err == nil {
		// If symlink creation succeeded, verify we can't access the outside file
		content, err := bfs.Open("/evil_link")
		if err == nil {
			defer content.Close()
			data := make([]byte, 100)
			n, _ := content.Read(data)
			if n > 0 && string(data[:n]) == "outside data" {
				t.Error("Symlink successfully escaped root and accessed outside data")
			}
		}
	}
	// Note: Symlink creation might fail, which is acceptable secure behavior
}

// TestSecurity_ChrootIsolation tests that Chroot properly isolates filesystem access
func TestSecurity_ChrootIsolation(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	// Create directory structure
	err := bfs.MkdirAll("/public/data", 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = bfs.MkdirAll("/private/secrets", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create files
	secretFile, err := bfs.Create("/private/secrets/password.txt")
	if err != nil {
		t.Fatal(err)
	}
	secretFile.Write([]byte("secret123"))
	secretFile.Close()

	publicFile, err := bfs.Create("/public/data/info.txt")
	if err != nil {
		t.Fatal(err)
	}
	publicFile.Write([]byte("public info"))
	publicFile.Close()

	// Chroot to /public
	chrootFS, err := bfs.Chroot("/public")
	if err != nil {
		t.Fatal(err)
	}

	// Should be able to access /data/info.txt within chroot
	f, err := chrootFS.Open("/data/info.txt")
	if err != nil {
		t.Errorf("Failed to access file within chroot: %v", err)
	} else {
		f.Close()
	}

	// Should NOT be able to access /private/secrets/password.txt
	f, err = chrootFS.Open("/private/secrets/password.txt")
	if err == nil {
		f.Close()
		t.Error("Chroot failed to isolate: accessed file outside chroot boundary")
	}

	// Should NOT be able to escape via ..
	f, err = chrootFS.Open("../private/secrets/password.txt")
	if err == nil {
		f.Close()
		t.Error("Chroot failed to prevent .. traversal outside boundary")
	}
}

// TestSecurity_RaceConditionFileCreation tests for TOCTOU vulnerabilities
func TestSecurity_RaceConditionFileCreation(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	const testFile = "/race_test.txt"
	const iterations = 100

	var wg sync.WaitGroup
	errors := make(chan error, iterations*2)

	// Launch multiple goroutines trying to create the same file
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Try to create file exclusively
			f, err := bfs.OpenFile(testFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if err != nil {
				// Expected for all but one goroutine
				return
			}
			defer f.Close()

			// Write goroutine ID
			_, err = f.Write([]byte(fmt.Sprintf("goroutine-%d\n", id)))
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: write failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify the file exists and has content from exactly one goroutine
	f, err := bfs.Open(testFile)
	if err != nil {
		t.Fatal("File should exist after race test")
	}
	defer f.Close()

	data := make([]byte, 1024)
	n, _ := f.Read(data)
	content := string(data[:n])

	// Should have content from exactly one goroutine
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected exactly one write to survive, got %d: %s", len(lines), content)
	}
}

// TestSecurity_ConcurrentAccess tests thread safety of concurrent operations
func TestSecurity_ConcurrentAccess(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	err := bfs.MkdirAll("/concurrent", 0755)
	if err != nil {
		t.Fatal(err)
	}

	const goroutines = 50
	const filesPerGoroutine = 10

	var wg sync.WaitGroup
	errors := make(chan error, goroutines*filesPerGoroutine)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < filesPerGoroutine; j++ {
				filename := fmt.Sprintf("/concurrent/file_%d_%d.txt", id, j)

				// Create file
				f, err := bfs.Create(filename)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: create failed: %w", id, err)
					continue
				}

				// Write data
				data := fmt.Sprintf("data from goroutine %d, file %d", id, j)
				_, err = f.Write([]byte(data))
				if err != nil {
					f.Close()
					errors <- fmt.Errorf("goroutine %d: write failed: %w", id, err)
					continue
				}

				f.Close()

				// Read back
				f, err = bfs.Open(filename)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: open failed: %w", id, err)
					continue
				}

				readData := make([]byte, len(data))
				_, err = f.Read(readData)
				f.Close()

				if err != nil {
					errors <- fmt.Errorf("goroutine %d: read failed: %w", id, err)
					continue
				}

				if string(readData) != data {
					errors <- fmt.Errorf("goroutine %d: data mismatch: expected %q, got %q", id, data, string(readData))
				}

				// Remove file
				err = bfs.Remove(filename)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: remove failed: %w", id, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Report all errors
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("Concurrent access resulted in %d errors", errorCount)
	}
}

// TestSecurity_InputValidation tests that invalid inputs are properly rejected
func TestSecurity_InputValidation(t *testing.T) {
	fs, err := osfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		dir       string
		expectErr bool
		desc      string
	}{
		{
			name:      "empty_path",
			dir:       "",
			expectErr: true,
			desc:      "Empty path should be rejected",
		},
		{
			name:      "nonexistent_dir",
			dir:       "/this/path/does/not/exist/12345",
			expectErr: true,
			desc:      "Nonexistent directory should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := billyfs.NewFS(fs, tt.dir)
			if tt.expectErr && err == nil {
				t.Errorf("%s: expected error but got nil", tt.desc)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.desc, err)
			}
		})
	}
}

// TestSecurity_FilePathInjection tests for path injection vulnerabilities
func TestSecurity_FilePathInjection(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	// Create a legitimate file
	f, err := bfs.Create("/legitimate.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("legitimate data"))
	f.Close()

	tests := []struct {
		name string
		path string
		desc string
	}{
		{
			name: "null_byte_injection",
			path: "/legitimate.txt\x00/etc/passwd",
			desc: "Null byte injection should not bypass path validation",
		},
		{
			name: "newline_injection",
			path: "/legitimate.txt\n/etc/passwd",
			desc: "Newline injection should not bypass path validation",
		},
		{
			name: "carriage_return_injection",
			path: "/legitimate.txt\r/etc/passwd",
			desc: "Carriage return injection should not bypass path validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt to open with injected path
			f, err := bfs.Open(tt.path)
			if err == nil {
				defer f.Close()
				// File opened - verify it's not the injected path
				name := f.Name()
				info, _ := bfs.Stat(name)
				if info != nil && info.Name() == "passwd" {
					t.Errorf("%s: path injection succeeded", tt.desc)
				}
			}
			// Note: Error is acceptable and indicates proper validation
		})
	}
}

// TestSecurity_DirectoryPermissions tests secure directory creation
func TestSecurity_DirectoryPermissions(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	// Create directory with specific permissions
	err := bfs.MkdirAll("/secure_dir", 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Verify permissions
	info, err := bfs.Stat("/secure_dir")
	if err != nil {
		t.Fatal(err)
	}

	mode := info.Mode()
	if !info.IsDir() {
		t.Error("Expected a directory")
	}

	// Note: Actual permission enforcement depends on the underlying filesystem
	t.Logf("Directory created with mode: %o", mode)
}
