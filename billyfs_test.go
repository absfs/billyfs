package billyfs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/absfs/billyfs"
	"github.com/absfs/osfs"

	"github.com/go-git/go-billy/v5"
)

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
		systemTempDir := os.TempDir()
		if !strings.HasPrefix(file.Name(), systemTempDir) {
			t.Errorf("Expected file to be in system temp dir %s, but got %s", systemTempDir, file.Name())
		}

		// Verify the file has the prefix
		basename := filepath.Base(file.Name())
		if !strings.HasPrefix(basename, "test_prefix") {
			t.Errorf("Expected filename to start with 'test_prefix', but got %s", basename)
		}
	})
}
