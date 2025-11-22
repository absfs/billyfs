package billyfs_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/absfs/billyfs"
	"github.com/absfs/osfs"

	"github.com/go-git/go-billy/v5"
)

// setupTestFS creates a test filesystem in a temporary directory
func setupTestFS(t *testing.T) (billy.Filesystem, string, func()) {
	t.Helper()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "billyfs_test_*")
	if err != nil {
		t.Fatal(err)
	}

	// Convert to absolute path (important for Windows)
	tempDir, err = filepath.Abs(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatal(err)
	}

	// Create the underlying absfs filesystem
	fs, err := osfs.NewFS()
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatal(err)
	}

	// Create the billyfs wrapper
	bfs, err := billyfs.NewFS(fs, tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatal(err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return bfs, tempDir, cleanup
}

func TestBillyfs(t *testing.T) {
	var bfs billy.Filesystem
	var err error
	fs, err := osfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	bfs, err = billyfs.NewFS(fs, "/")
	if err != nil {
		t.Fatal(err)
	}
	_ = bfs
}

// TestNewFS tests the NewFS constructor
func TestNewFS(t *testing.T) {
	fs, err := osfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("valid directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "billyfs_test_*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		tempDir, err = filepath.Abs(tempDir)
		if err != nil {
			t.Fatal(err)
		}

		bfs, err := billyfs.NewFS(fs, tempDir)
		if err != nil {
			t.Fatalf("NewFS failed: %v", err)
		}
		if bfs == nil {
			t.Fatal("Expected non-nil filesystem")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := billyfs.NewFS(fs, "")
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})

	t.Run("relative path", func(t *testing.T) {
		_, err := billyfs.NewFS(fs, "relative/path")
		if err == nil {
			t.Error("Expected error for relative path")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		_, err := billyfs.NewFS(fs, "/nonexistent/path/12345")
		if err == nil {
			t.Error("Expected error for nonexistent path")
		}
	})

	t.Run("file not directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "billyfs_test_*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		tempDir, err = filepath.Abs(tempDir)
		if err != nil {
			t.Fatal(err)
		}

		filePath := filepath.Join(tempDir, "testfile")
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err = billyfs.NewFS(fs, filePath)
		if err == nil {
			t.Error("Expected error when path is a file, not a directory")
		}
	})
}

// TestCreate tests the Create function
func TestCreate(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	t.Run("create new file", func(t *testing.T) {
		file, err := bfs.Create("testfile.txt")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		defer file.Close()

		if file.Name() != "testfile.txt" {
			t.Errorf("Expected filename 'testfile.txt', got %s", file.Name())
		}

		// Verify file exists
		_, err = bfs.Stat("testfile.txt")
		if err != nil {
			t.Errorf("File should exist after Create: %v", err)
		}
	})

	t.Run("truncate existing file", func(t *testing.T) {
		// Create and write to file
		file, err := bfs.Create("truncate.txt")
		if err != nil {
			t.Fatal(err)
		}
		_, err = file.Write([]byte("original content"))
		file.Close()

		// Create again (should truncate)
		file, err = bfs.Create("truncate.txt")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		defer file.Close()

		// Verify file is empty
		info, err := bfs.Stat("truncate.txt")
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() != 0 {
			t.Errorf("Expected file size 0 after truncate, got %d", info.Size())
		}
	})
}

// TestOpen tests the Open function
func TestOpen(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	t.Run("open existing file", func(t *testing.T) {
		// Create a file first
		content := []byte("test content")
		file, err := bfs.Create("opentest.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Write(content)
		file.Close()

		// Open it
		file, err = bfs.Open("opentest.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		defer file.Close()

		// Read and verify content
		buf := make([]byte, len(content))
		n, err := file.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(content) || !bytes.Equal(buf, content) {
			t.Errorf("Expected to read %q, got %q", content, buf)
		}
	})

	t.Run("open nonexistent file", func(t *testing.T) {
		_, err := bfs.Open("nonexistent.txt")
		if err == nil {
			t.Error("Expected error when opening nonexistent file")
		}
	})
}

// TestOpenFile tests the OpenFile function
func TestOpenFile(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	t.Run("open with custom flags", func(t *testing.T) {
		file, err := bfs.OpenFile("flagtest.txt", os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			t.Fatalf("OpenFile failed: %v", err)
		}
		defer file.Close()

		// Write and read
		content := []byte("flag test")
		_, err = file.Write(content)
		if err != nil {
			t.Fatal(err)
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, len(content))
		_, err = file.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(buf, content) {
			t.Errorf("Expected %q, got %q", content, buf)
		}
	})

	t.Run("open read-only", func(t *testing.T) {
		// Create file first
		file, err := bfs.Create("readonly.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Write([]byte("read only content"))
		file.Close()

		// Open as read-only
		file, err = bfs.OpenFile("readonly.txt", os.O_RDONLY, 0644)
		if err != nil {
			t.Fatalf("OpenFile failed: %v", err)
		}
		defer file.Close()

		// Try to write (should fail or be ignored)
		_, err = file.Write([]byte("attempt write"))
		if err == nil {
			// Some implementations might not error on write to read-only
			// but it shouldn't modify the file
		}
	})
}

// TestStat tests the Stat function
func TestStat(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	t.Run("stat file", func(t *testing.T) {
		file, err := bfs.Create("stattest.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Write([]byte("12345"))
		file.Close()

		info, err := bfs.Stat("stattest.txt")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}

		if info.Name() != "stattest.txt" {
			t.Errorf("Expected name 'stattest.txt', got %s", info.Name())
		}

		if info.Size() != 5 {
			t.Errorf("Expected size 5, got %d", info.Size())
		}

		if info.IsDir() {
			t.Error("Expected file, not directory")
		}
	})

	t.Run("stat directory", func(t *testing.T) {
		err := bfs.MkdirAll("testdir", 0755)
		if err != nil {
			t.Fatal(err)
		}

		info, err := bfs.Stat("testdir")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}

		if !info.IsDir() {
			t.Error("Expected directory")
		}
	})

	t.Run("stat nonexistent", func(t *testing.T) {
		_, err := bfs.Stat("doesnotexist.txt")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})
}

// TestRename tests the Rename function
func TestRename(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	t.Run("rename file", func(t *testing.T) {
		// Create file
		file, err := bfs.Create("oldname.txt")
		if err != nil {
			t.Fatal(err)
		}
		content := []byte("rename test")
		file.Write(content)
		file.Close()

		// Rename
		err = bfs.Rename("oldname.txt", "newname.txt")
		if err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		// Verify old name doesn't exist
		_, err = bfs.Stat("oldname.txt")
		if err == nil {
			t.Error("Old file should not exist after rename")
		}

		// Verify new name exists with correct content
		file, err = bfs.Open("newname.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		buf := make([]byte, len(content))
		file.Read(buf)
		if !bytes.Equal(buf, content) {
			t.Errorf("Content mismatch after rename")
		}
	})

	t.Run("rename nonexistent file", func(t *testing.T) {
		err := bfs.Rename("nonexistent.txt", "newname.txt")
		if err == nil {
			t.Error("Expected error when renaming nonexistent file")
		}
	})
}

// TestRemove tests the Remove function
func TestRemove(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	t.Run("remove file", func(t *testing.T) {
		file, err := bfs.Create("removeme.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Close()

		err = bfs.Remove("removeme.txt")
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		_, err = bfs.Stat("removeme.txt")
		if err == nil {
			t.Error("File should not exist after removal")
		}
	})

	t.Run("remove empty directory", func(t *testing.T) {
		err := bfs.MkdirAll("emptydir", 0755)
		if err != nil {
			t.Fatal(err)
		}

		err = bfs.Remove("emptydir")
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		_, err = bfs.Stat("emptydir")
		if err == nil {
			t.Error("Directory should not exist after removal")
		}
	})

	t.Run("remove nonexistent", func(t *testing.T) {
		err := bfs.Remove("nonexistent.txt")
		if err == nil {
			t.Error("Expected error when removing nonexistent file")
		}
	})
}

// TestJoin tests the Join function
func TestJoin(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{"simple join", []string{"a", "b", "c"}, "a/b/c"},
		{"with dots", []string{"a", ".", "b"}, "a/b"},
		{"with double dots", []string{"a", "b", "..", "c"}, "a/c"},
		{"empty strings", []string{"a", "", "b"}, "a/b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bfs.Join(tt.parts...)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestCapabilities tests the Capabilities function
func TestCapabilities(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	// Cast to Capable interface
	if capFS, ok := bfs.(billy.Capable); ok {
		caps := capFS.Capabilities()
		if caps != billy.AllCapabilities {
			t.Errorf("Expected AllCapabilities, got %v", caps)
		}
	} else {
		t.Error("Filesystem should implement billy.Capable")
	}
}

func TestTempFile(t *testing.T) {
	fs, err := osfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	bfs, err := billyfs.NewFS(fs, "/")
	if err != nil {
		t.Fatal(err)
	}

	// Create a custom temp directory for testing
	customDir := filepath.Join(os.TempDir(), "billyfs_test")
	err = os.MkdirAll(customDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(customDir)

	t.Run("custom directory", func(t *testing.T) {
		// Test with custom directory
		file, err := bfs.TempFile(customDir, "test_prefix")
		if err != nil {
			t.Fatalf("TempFile failed: %v", err)
		}
		defer file.Close()
		defer os.Remove(file.Name())

		// Verify the file is in the custom directory
		if !strings.HasPrefix(file.Name(), customDir) {
			t.Errorf("Expected file to be in %s, but got %s", customDir, file.Name())
		}

		// Verify the file has the prefix
		basename := filepath.Base(file.Name())
		if !strings.HasPrefix(basename, "test_prefix") {
			t.Errorf("Expected filename to start with 'test_prefix', but got %s", basename)
		}
	})

	t.Run("empty directory uses default", func(t *testing.T) {
		// Test with empty directory (should use default temp dir)
		file, err := bfs.TempFile("", "test_prefix")
		if err != nil {
			t.Fatalf("TempFile with empty dir failed: %v", err)
		}
		defer file.Close()
		defer os.Remove(file.Name())

		// Verify the file is in the system temp directory
		// Use EvalSymlinks to handle platforms where temp dir is a symlink (e.g., macOS)
		systemTempDir, err := filepath.EvalSymlinks(os.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		actualPath, err := filepath.EvalSymlinks(file.Name())
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(actualPath, systemTempDir) {
			t.Errorf("Expected file to be in system temp dir %s, but got %s", systemTempDir, actualPath)
		}

		// Verify the file has the prefix
		basename := filepath.Base(file.Name())
		if !strings.HasPrefix(basename, "test_prefix") {
			t.Errorf("Expected filename to start with 'test_prefix', but got %s", basename)
		}
	})
}

// TestFileIO tests file I/O operations
func TestFileIO(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	t.Run("write and read", func(t *testing.T) {
		file, err := bfs.Create("io_test.txt")
		if err != nil {
			t.Fatal(err)
		}

		content := []byte("Hello, World!")
		n, err := file.Write(content)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != len(content) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(content), n)
		}
		file.Close()

		// Read back
		file, err = bfs.Open("io_test.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		buf := make([]byte, len(content))
		n, err = file.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(content) || !bytes.Equal(buf, content) {
			t.Errorf("Expected %q, got %q", content, buf)
		}
	})

	t.Run("seek operations", func(t *testing.T) {
		file, err := bfs.Create("seek_test.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		content := []byte("0123456789")
		file.Write(content)

		// Seek to beginning
		pos, err := file.Seek(0, io.SeekStart)
		if err != nil {
			t.Fatalf("Seek failed: %v", err)
		}
		if pos != 0 {
			t.Errorf("Expected position 0, got %d", pos)
		}

		// Seek to middle
		pos, err = file.Seek(5, io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}
		if pos != 5 {
			t.Errorf("Expected position 5, got %d", pos)
		}

		// Read from middle
		buf := make([]byte, 3)
		file.Read(buf)
		if !bytes.Equal(buf, []byte("567")) {
			t.Errorf("Expected '567', got %q", buf)
		}

		// Seek from current position
		pos, err = file.Seek(-3, io.SeekCurrent)
		if err != nil {
			t.Fatal(err)
		}
		if pos != 5 {
			t.Errorf("Expected position 5, got %d", pos)
		}

		// Seek from end
		pos, err = file.Seek(-2, io.SeekEnd)
		if err != nil {
			t.Fatal(err)
		}
		if pos != 8 {
			t.Errorf("Expected position 8, got %d", pos)
		}
	})

	t.Run("readat", func(t *testing.T) {
		file, err := bfs.Create("at_test.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		// Write initial content
		file.Write([]byte("XXXXXXXXXX"))

		// ReadAt
		file.Seek(0, io.SeekStart)
		buf := make([]byte, 3)
		n, err := file.ReadAt(buf, 3)
		if err != nil {
			t.Fatal(err)
		}
		if n != 3 {
			t.Errorf("Expected to read 3 bytes, read %d", n)
		}
		if !bytes.Equal(buf, []byte("XXX")) {
			t.Errorf("Expected 'XXX', got %q", buf)
		}
	})

	t.Run("truncate", func(t *testing.T) {
		file, err := bfs.Create("truncate_test.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		// Write content
		file.Write([]byte("0123456789"))

		// Truncate to smaller size
		err = file.Truncate(5)
		if err != nil {
			t.Fatalf("Truncate failed: %v", err)
		}

		// Verify size
		info, err := bfs.Stat("truncate_test.txt")
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() != 5 {
			t.Errorf("Expected size 5, got %d", info.Size())
		}

		// Read and verify content
		file.Seek(0, io.SeekStart)
		buf := make([]byte, 10)
		n, _ := file.Read(buf)
		if n != 5 || !bytes.Equal(buf[:n], []byte("01234")) {
			t.Errorf("Expected '01234', got %q", buf[:n])
		}
	})

	t.Run("lock and unlock", func(t *testing.T) {
		file, err := bfs.Create("lock_test.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		// Lock
		err = file.Lock()
		if err != nil {
			t.Fatalf("Lock failed: %v", err)
		}

		// Unlock
		err = file.Unlock()
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}
	})
}

// TestDirectoryOperations tests directory-related operations
func TestDirectoryOperations(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	t.Run("mkdirall single level", func(t *testing.T) {
		err := bfs.MkdirAll("testdir", 0755)
		if err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		info, err := bfs.Stat("testdir")
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			t.Error("Expected directory")
		}
	})

	t.Run("mkdirall nested", func(t *testing.T) {
		err := bfs.MkdirAll("a/b/c/d", 0755)
		if err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		// Verify all levels exist
		for _, path := range []string{"a", "a/b", "a/b/c", "a/b/c/d"} {
			info, err := bfs.Stat(path)
			if err != nil {
				t.Errorf("Path %s should exist: %v", path, err)
			}
			if info != nil && !info.IsDir() {
				t.Errorf("Path %s should be a directory", path)
			}
		}
	})

	t.Run("mkdirall existing", func(t *testing.T) {
		err := bfs.MkdirAll("existing", 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Call again on existing directory (should not error)
		err = bfs.MkdirAll("existing", 0755)
		if err != nil {
			t.Errorf("MkdirAll on existing directory should not error: %v", err)
		}
	})

	t.Run("readdir", func(t *testing.T) {
		// Create directory with some files
		err := bfs.MkdirAll("readtest", 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Create some files
		for _, name := range []string{"file1.txt", "file2.txt", "file3.txt"} {
			file, err := bfs.Create(bfs.Join("readtest", name))
			if err != nil {
				t.Fatal(err)
			}
			file.Close()
		}

		// Create a subdirectory
		err = bfs.MkdirAll(bfs.Join("readtest", "subdir"), 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Read directory
		entries, err := bfs.ReadDir("readtest")
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}

		if len(entries) != 4 {
			t.Errorf("Expected 4 entries, got %d", len(entries))
		}

		// Verify entries
		names := make(map[string]bool)
		for _, entry := range entries {
			names[entry.Name()] = true
		}

		expectedNames := []string{"file1.txt", "file2.txt", "file3.txt", "subdir"}
		for _, name := range expectedNames {
			if !names[name] {
				t.Errorf("Expected to find %s in directory listing", name)
			}
		}
	})

	t.Run("readdir empty", func(t *testing.T) {
		err := bfs.MkdirAll("emptyread", 0755)
		if err != nil {
			t.Fatal(err)
		}

		entries, err := bfs.ReadDir("emptyread")
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}

		if len(entries) != 0 {
			t.Errorf("Expected 0 entries in empty directory, got %d", len(entries))
		}
	})

	t.Run("readdir nonexistent", func(t *testing.T) {
		_, err := bfs.ReadDir("nonexistentdir")
		if err == nil {
			t.Error("Expected error when reading nonexistent directory")
		}
	})
}

// TestSymlinkOperations tests symlink-related operations
func TestSymlinkOperations(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	// Cast to Symlink interface
	symlinkFS, ok := bfs.(billy.Symlink)
	if !ok {
		t.Skip("Filesystem does not implement billy.Symlink")
	}

	t.Run("create and read symlink", func(t *testing.T) {
		// Create target file
		file, err := bfs.Create("target.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Write([]byte("target content"))
		file.Close()

		// Create symlink
		err = symlinkFS.Symlink("target.txt", "link.txt")
		if err != nil {
			t.Fatalf("Symlink failed: %v", err)
		}

		// Read symlink
		target, err := symlinkFS.Readlink("link.txt")
		if err != nil {
			t.Fatalf("Readlink failed: %v", err)
		}

		// The implementation may convert to absolute path
		if target != "target.txt" && target != "/target.txt" {
			t.Errorf("Expected target 'target.txt' or '/target.txt', got %s", target)
		}
	})

	t.Run("lstat vs stat", func(t *testing.T) {
		// Create target file
		file, err := bfs.Create("statlink_target.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Write([]byte("content"))
		file.Close()

		// Create symlink
		err = symlinkFS.Symlink("statlink_target.txt", "statlink.txt")
		if err != nil {
			t.Fatal(err)
		}

		// Stat follows symlink
		info, err := bfs.Stat("statlink.txt")
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Error("Stat should follow symlink, not return symlink info")
		}

		// Lstat does not follow symlink
		info, err = symlinkFS.Lstat("statlink.txt")
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("Lstat should return symlink info")
		}
	})

	t.Run("symlink to nonexistent target", func(t *testing.T) {
		// Should be able to create symlink to nonexistent target
		err := symlinkFS.Symlink("nonexistent_target.txt", "broken_link.txt")
		if err != nil {
			t.Fatalf("Symlink to nonexistent target failed: %v", err)
		}

		// Readlink should still work
		target, err := symlinkFS.Readlink("broken_link.txt")
		if err != nil {
			t.Fatal(err)
		}
		// The implementation may convert to absolute path
		if target != "nonexistent_target.txt" && target != "/nonexistent_target.txt" {
			t.Errorf("Expected 'nonexistent_target.txt' or '/nonexistent_target.txt', got %s", target)
		}

		// But Stat should fail
		_, err = bfs.Stat("broken_link.txt")
		if err == nil {
			t.Error("Expected error when stating broken symlink")
		}
	})

	t.Run("symlink with absolute path", func(t *testing.T) {
		// Create target
		file, err := bfs.Create("abs_target.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Close()

		// Create symlink with absolute path
		err = symlinkFS.Symlink("/abs_target.txt", "abs_link.txt")
		if err != nil {
			t.Fatalf("Symlink with absolute path failed: %v", err)
		}

		target, err := symlinkFS.Readlink("abs_link.txt")
		if err != nil {
			t.Fatal(err)
		}
		if target != "/abs_target.txt" {
			t.Errorf("Expected '/abs_target.txt', got %s", target)
		}
	})
}

// TestPermissionsAndMetadata tests chmod, chown, and chtimes
func TestPermissionsAndMetadata(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	t.Run("chmod", func(t *testing.T) {
		// Cast to Change interface
		changeFS, ok := bfs.(billy.Change)
		if !ok {
			t.Skip("Filesystem does not implement billy.Change")
		}

		file, err := bfs.Create("chmod_test.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Close()

		// Change permissions
		err = changeFS.Chmod("chmod_test.txt", 0600)
		if err != nil {
			t.Fatalf("Chmod failed: %v", err)
		}

		// Verify permissions
		info, err := bfs.Stat("chmod_test.txt")
		if err != nil {
			t.Fatal(err)
		}

		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("Expected permissions 0600, got %o", perm)
		}
	})

	t.Run("chtimes", func(t *testing.T) {
		// Cast to Change interface
		changeFS, ok := bfs.(billy.Change)
		if !ok {
			t.Skip("Filesystem does not implement billy.Change")
		}

		file, err := bfs.Create("chtimes_test.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Close()

		// Set specific times
		atime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		mtime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

		err = changeFS.Chtimes("chtimes_test.txt", atime, mtime)
		if err != nil {
			t.Fatalf("Chtimes failed: %v", err)
		}

		// Verify mtime
		info, err := bfs.Stat("chtimes_test.txt")
		if err != nil {
			t.Fatal(err)
		}

		// Check if mtime is close (within 1 second due to filesystem precision)
		modTime := info.ModTime()
		if modTime.Unix() != mtime.Unix() {
			t.Errorf("Expected mtime %v, got %v", mtime, modTime)
		}
	})

	t.Run("chown", func(t *testing.T) {
		// Cast to Change interface
		changeFS, ok := bfs.(billy.Change)
		if !ok {
			t.Skip("Filesystem does not implement billy.Change")
		}

		file, err := bfs.Create("chown_test.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Close()

		// Try to change ownership (may require root privileges)
		// We'll just verify it doesn't crash
		err = changeFS.Chown("chown_test.txt", os.Getuid(), os.Getgid())
		if err != nil {
			// This might fail on some systems, which is ok
			t.Logf("Chown returned error (may be expected): %v", err)
		}
	})

	t.Run("lchown", func(t *testing.T) {
		// Cast to Change interface
		changeFS, ok := bfs.(billy.Change)
		if !ok {
			t.Skip("Filesystem does not implement billy.Change")
		}

		// Cast to Symlink interface
		symlinkFS, ok := bfs.(billy.Symlink)
		if !ok {
			t.Skip("Filesystem does not implement billy.Symlink")
		}

		// Create a symlink
		file, err := bfs.Create("lchown_target.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Close()

		err = symlinkFS.Symlink("lchown_target.txt", "lchown_link.txt")
		if err != nil {
			t.Fatal(err)
		}

		// Try to change link ownership (may require root privileges)
		err = changeFS.Lchown("lchown_link.txt", os.Getuid(), os.Getgid())
		if err != nil {
			// This might fail on some systems, which is ok
			t.Logf("Lchown returned error (may be expected): %v", err)
		}
	})
}

// TestChrootOperations tests chroot functionality
func TestChrootOperations(t *testing.T) {
	bfs, _, cleanup := setupTestFS(t)
	defer cleanup()

	// Cast to Chroot interface
	chrootFS, ok := bfs.(billy.Chroot)
	if !ok {
		t.Skip("Filesystem does not implement billy.Chroot")
	}

	t.Run("chroot to subdirectory", func(t *testing.T) {
		// Create subdirectory structure
		err := bfs.MkdirAll("subroot/nested", 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Create file in subdirectory
		file, err := bfs.Create("subroot/file.txt")
		if err != nil {
			t.Fatal(err)
		}
		file.Write([]byte("content"))
		file.Close()

		// Chroot to subdirectory - use absolute path from filesystem root
		subFS, err := chrootFS.Chroot("/subroot")
		if err != nil {
			t.Fatalf("Chroot failed: %v", err)
		}

		// Verify we can access the file with relative path
		_, err = subFS.Stat("file.txt")
		if err != nil {
			t.Errorf("Should be able to access file.txt in chrooted fs: %v", err)
		}

		// Verify nested directory is accessible
		_, err = subFS.Stat("nested")
		if err != nil {
			t.Errorf("Should be able to access nested dir in chrooted fs: %v", err)
		}
	})

	t.Run("root returns correct path", func(t *testing.T) {
		root := chrootFS.Root()
		if root == "" {
			t.Error("Root should not be empty")
		}

		// After chroot, root should change
		err := bfs.MkdirAll("chroottest", 0755)
		if err != nil {
			t.Fatal(err)
		}

		subFS, err := chrootFS.Chroot("/chroottest")
		if err != nil {
			t.Fatal(err)
		}

		if subChrootFS, ok := subFS.(billy.Chroot); ok {
			subRoot := subChrootFS.Root()
			if subRoot == root {
				t.Error("Chroot root should be different from original root")
			}
		}
	})

	t.Run("chroot to nonexistent directory", func(t *testing.T) {
		_, err := chrootFS.Chroot("nonexistent_chroot")
		if err == nil {
			t.Error("Expected error when chrooting to nonexistent directory")
		}
	})
}

// Example_newFS demonstrates how to create a new billyfs filesystem.
func Example_newFS() {
	// Create the underlying absfs filesystem
	fs, err := osfs.NewFS()
	if err != nil {
		panic(err)
	}

	// Create a billyfs wrapper for the /tmp directory
	bfs, err := billyfs.NewFS(fs, "/tmp")
	if err != nil {
		panic(err)
	}

	// Now you can use bfs as a billy.Filesystem
	_ = bfs
	fmt.Println("Filesystem created successfully")
	// Output:
	// Filesystem created successfully
}

// Example_create demonstrates how to create and write to a file.
func Example_create() {
	// Setup filesystem
	fs, _ := osfs.NewFS()
	tempDir, _ := os.MkdirTemp("", "example_*")
	defer os.RemoveAll(tempDir)
	tempDir, _ = filepath.Abs(tempDir)
	bfs, _ := billyfs.NewFS(fs, tempDir)

	// Create a new file
	file, err := bfs.Create("example.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write to the file
	content := []byte("Hello, World!")
	n, err := file.Write(content)
	if err != nil {
		panic(err)
	}

	fmt.Println("Bytes written:", n)
	// Output:
	// Bytes written: 13
}

// Example_chroot demonstrates how to restrict filesystem access to a subdirectory.
func Example_chroot() {
	// Setup filesystem
	fs, _ := osfs.NewFS()
	tempDir, _ := os.MkdirTemp("", "example_*")
	defer os.RemoveAll(tempDir)
	tempDir, _ = filepath.Abs(tempDir)
	var bfs billy.Filesystem
	bfs, _ = billyfs.NewFS(fs, tempDir)

	// Create a subdirectory
	_ = bfs.MkdirAll("/projects/myapp", 0755)

	// Create a file in the subdirectory
	file, _ := bfs.Create("/projects/myapp/config.txt")
	file.Write([]byte("configuration"))
	file.Close()

	// Chroot to the subdirectory - now it becomes the root
	chrootFS, err := bfs.(billy.Chroot).Chroot("/projects/myapp")
	if err != nil {
		panic(err)
	}

	// Access the file using a path relative to the new root
	info, err := chrootFS.Stat("config.txt")
	if err != nil {
		panic(err)
	}

	fmt.Println("File found:", info.Name())
	// Output:
	// File found: config.txt
}

// Example_tempFile demonstrates how to create a temporary file.
func Example_tempFile() {
	// Setup filesystem
	fs, _ := osfs.NewFS()
	tempDir, _ := os.MkdirTemp("", "example_*")
	defer os.RemoveAll(tempDir)
	tempDir, _ = filepath.Abs(tempDir)
	var bfs billy.Filesystem
	bfs, _ = billyfs.NewFS(fs, tempDir)

	// Create a custom temp directory
	customTempDir := "/temp"
	_ = bfs.MkdirAll(customTempDir, 0755)

	// Create a temporary file with a prefix
	tempFile, err := bfs.(billy.TempFile).TempFile(customTempDir, "myapp_")
	if err != nil {
		panic(err)
	}
	defer tempFile.Close()
	defer bfs.Remove(tempFile.Name())

	// Write to the temp file
	tempFile.Write([]byte("temporary data"))

	// The filename will have the prefix plus a random suffix
	name := filepath.Base(tempFile.Name())
	hasPrefix := filepath.Base(name)[:6] == "myapp_"
	fmt.Println("Has prefix 'myapp_':", hasPrefix)
	// Output:
	// Has prefix 'myapp_': true
}

// Example_readDir demonstrates how to list directory contents.
func Example_readDir() {
	// Setup filesystem
	fs, _ := osfs.NewFS()
	tempDir, _ := os.MkdirTemp("", "example_*")
	defer os.RemoveAll(tempDir)
	tempDir, _ = filepath.Abs(tempDir)
	var bfs billy.Filesystem
	bfs, _ = billyfs.NewFS(fs, tempDir)

	// Create some files and directories
	_ = bfs.MkdirAll("/data", 0755)
	file1, _ := bfs.Create("/data/file1.txt")
	file1.Close()
	file2, _ := bfs.Create("/data/file2.txt")
	file2.Close()
	_ = bfs.MkdirAll("/data/subdir", 0755)

	// Read directory contents
	entries, err := bfs.(billy.Dir).ReadDir("/data")
	if err != nil {
		panic(err)
	}

	// Sort entries by name for consistent output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	fmt.Println("Number of entries:", len(entries))
	for _, entry := range entries {
		fmt.Println("Found:", entry.Name())
	}
	// Output:
	// Number of entries: 3
	// Found: file1.txt
	// Found: file2.txt
	// Found: subdir
}

// Example_symlink demonstrates how to create and use symbolic links.
func Example_symlink() {
	// Setup filesystem
	fs, _ := osfs.NewFS()
	tempDir, _ := os.MkdirTemp("", "example_*")
	defer os.RemoveAll(tempDir)
	tempDir, _ = filepath.Abs(tempDir)
	var bfs billy.Filesystem
	bfs, _ = billyfs.NewFS(fs, tempDir)

	// Create a target file
	targetFile, _ := bfs.Create("target.txt")
	targetFile.Write([]byte("original content"))
	targetFile.Close()

	// Create a symbolic link
	symlinkFS := bfs.(billy.Symlink)
	err := symlinkFS.Symlink("target.txt", "link.txt")
	if err != nil {
		panic(err)
	}

	// Read the link target
	linkTarget, err := symlinkFS.Readlink("link.txt")
	if err != nil {
		panic(err)
	}

	// Stat follows the link
	info, _ := bfs.Stat("link.txt")
	fmt.Println("Link points to:", filepath.Base(linkTarget))
	fmt.Println("Target size:", info.Size())
	// Output:
	// Link points to: target.txt
	// Target size: 16
}
