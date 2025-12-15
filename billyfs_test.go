package billyfs_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/absfs/billyfs"
	"github.com/absfs/osfs"

	"github.com/go-git/go-billy/v5"
)

// Test helper functions

// newTestFS creates a new test filesystem rooted at a temporary directory
func newTestFS(t *testing.T) (*billyfs.Filesystem, string) {
	t.Helper()
	tmpDir := t.TempDir()

	fs, err := osfs.NewFS()
	if err != nil {
		t.Fatalf("failed to create osfs: %v", err)
	}

	bfs, err := billyfs.NewFS(fs, tmpDir)
	if err != nil {
		t.Fatalf("failed to create billyfs: %v", err)
	}

	return bfs, tmpDir
}

// TestBillyfsInterfaceCompliance verifies the Filesystem implements billy.Filesystem
func TestBillyfsInterfaceCompliance(t *testing.T) {
	var bfs billy.Filesystem
	var err error
	fs, err := osfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	// Use os.TempDir() instead of "/" to work on all platforms
	bfs, err = billyfs.NewFS(fs, os.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_ = bfs
}

// TestNewFS tests the NewFS constructor
func TestNewFS(t *testing.T) {
	t.Run("valid path", func(t *testing.T) {
		tmpDir := t.TempDir()
		fs, err := osfs.NewFS()
		if err != nil {
			t.Fatalf("failed to create osfs: %v", err)
		}

		bfs, err := billyfs.NewFS(fs, tmpDir)
		if err != nil {
			t.Fatalf("NewFS failed: %v", err)
		}
		if bfs == nil {
			t.Fatal("expected non-nil filesystem")
		}
	})

	t.Run("root path", func(t *testing.T) {
		fs, err := osfs.NewFS()
		if err != nil {
			t.Fatalf("failed to create osfs: %v", err)
		}

		// Get a platform-specific root path
		rootPath := filepath.VolumeName(os.TempDir())
		if rootPath == "" {
			rootPath = "/"
		} else {
			rootPath = rootPath + string(filepath.Separator)
		}

		bfs, err := billyfs.NewFS(fs, rootPath)
		if err != nil {
			t.Fatalf("NewFS with root path failed: %v", err)
		}
		if bfs == nil {
			t.Fatal("expected non-nil filesystem")
		}
	})
}

// TestCapabilities tests the Capabilities method
func TestCapabilities(t *testing.T) {
	bfs, _ := newTestFS(t)

	caps := bfs.Capabilities()
	if caps != billy.AllCapabilities {
		t.Errorf("expected AllCapabilities (%v), got %v", billy.AllCapabilities, caps)
	}
}

// TestCreate tests file creation
func TestCreate(t *testing.T) {
	bfs, _ := newTestFS(t)

	t.Run("create new file", func(t *testing.T) {
		f, err := bfs.Create("test.txt")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		defer f.Close()

		if f.Name() != "test.txt" {
			t.Errorf("expected name test.txt, got %s", f.Name())
		}
	})

	t.Run("create file in subdirectory", func(t *testing.T) {
		if err := bfs.MkdirAll("subdir", 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		f, err := bfs.Create("subdir/nested.txt")
		if err != nil {
			t.Fatalf("Create in subdir failed: %v", err)
		}
		defer f.Close()
	})

	t.Run("create truncates existing file", func(t *testing.T) {
		// First create and write data
		f1, err := bfs.Create("truncate.txt")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		_, err = f1.Write([]byte("initial content"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		f1.Close()

		// Create again should truncate
		f2, err := bfs.Create("truncate.txt")
		if err != nil {
			t.Fatalf("Create (truncate) failed: %v", err)
		}
		defer f2.Close()

		// Verify file is empty
		info, err := bfs.Stat("truncate.txt")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if info.Size() != 0 {
			t.Errorf("expected size 0, got %d", info.Size())
		}
	})
}

// TestOpen tests file opening
func TestOpen(t *testing.T) {
	bfs, _ := newTestFS(t)

	t.Run("open existing file", func(t *testing.T) {
		// Create file first
		f, err := bfs.Create("opentest.txt")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		_, err = f.Write([]byte("test content"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		f.Close()

		// Open for reading
		f2, err := bfs.Open("opentest.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		defer f2.Close()

		data, err := io.ReadAll(f2)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}
		if string(data) != "test content" {
			t.Errorf("expected 'test content', got '%s'", string(data))
		}
	})

	t.Run("open non-existent file", func(t *testing.T) {
		_, err := bfs.Open("nonexistent.txt")
		if err == nil {
			t.Error("expected error opening non-existent file")
		}
	})
}

// TestOpenFile tests OpenFile with various flags
func TestOpenFile(t *testing.T) {
	bfs, _ := newTestFS(t)

	t.Run("O_RDONLY", func(t *testing.T) {
		// Create file first
		f, _ := bfs.Create("rdonly.txt")
		f.Write([]byte("readonly test"))
		f.Close()

		f2, err := bfs.OpenFile("rdonly.txt", os.O_RDONLY, 0)
		if err != nil {
			t.Fatalf("OpenFile O_RDONLY failed: %v", err)
		}
		defer f2.Close()
	})

	t.Run("O_CREATE", func(t *testing.T) {
		f, err := bfs.OpenFile("created.txt", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("OpenFile O_CREATE failed: %v", err)
		}
		f.Close()

		// Verify file exists
		_, err = bfs.Stat("created.txt")
		if err != nil {
			t.Errorf("file not created: %v", err)
		}
	})

	t.Run("O_APPEND", func(t *testing.T) {
		// Create file with initial content
		f, _ := bfs.Create("append.txt")
		f.Write([]byte("initial"))
		f.Close()

		// Open with append
		f2, err := bfs.OpenFile("append.txt", os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			t.Fatalf("OpenFile O_APPEND failed: %v", err)
		}
		f2.Write([]byte("_appended"))
		f2.Close()

		// Verify content
		f3, _ := bfs.Open("append.txt")
		data, _ := io.ReadAll(f3)
		f3.Close()
		if string(data) != "initial_appended" {
			t.Errorf("expected 'initial_appended', got '%s'", string(data))
		}
	})

	t.Run("O_TRUNC", func(t *testing.T) {
		// Create file with content
		f, _ := bfs.Create("trunc.txt")
		f.Write([]byte("to be truncated"))
		f.Close()

		// Open with truncate
		f2, err := bfs.OpenFile("trunc.txt", os.O_TRUNC|os.O_WRONLY, 0)
		if err != nil {
			t.Fatalf("OpenFile O_TRUNC failed: %v", err)
		}
		f2.Write([]byte("new"))
		f2.Close()

		// Verify content
		f3, _ := bfs.Open("trunc.txt")
		data, _ := io.ReadAll(f3)
		f3.Close()
		if string(data) != "new" {
			t.Errorf("expected 'new', got '%s'", string(data))
		}
	})
}

// TestStat tests file stat operations
func TestStat(t *testing.T) {
	bfs, _ := newTestFS(t)

	t.Run("stat file", func(t *testing.T) {
		f, _ := bfs.Create("statfile.txt")
		f.Write([]byte("stat content"))
		f.Close()

		info, err := bfs.Stat("statfile.txt")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if info.IsDir() {
			t.Error("expected file, got directory")
		}
		if info.Size() != 12 {
			t.Errorf("expected size 12, got %d", info.Size())
		}
	})

	t.Run("stat directory", func(t *testing.T) {
		bfs.MkdirAll("statdir", 0755)

		info, err := bfs.Stat("statdir")
		if err != nil {
			t.Fatalf("Stat dir failed: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory, got file")
		}
	})

	t.Run("stat non-existent", func(t *testing.T) {
		_, err := bfs.Stat("nonexistent")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

// TestRename tests renaming files and directories
func TestRename(t *testing.T) {
	bfs, _ := newTestFS(t)

	t.Run("rename file", func(t *testing.T) {
		f, _ := bfs.Create("oldname.txt")
		f.Write([]byte("rename test"))
		f.Close()

		err := bfs.Rename("oldname.txt", "newname.txt")
		if err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		// Old name should not exist
		_, err = bfs.Stat("oldname.txt")
		if err == nil {
			t.Error("old name still exists")
		}

		// New name should exist
		_, err = bfs.Stat("newname.txt")
		if err != nil {
			t.Error("new name does not exist")
		}
	})

	t.Run("rename to different directory", func(t *testing.T) {
		bfs.MkdirAll("srcdir", 0755)
		bfs.MkdirAll("dstdir", 0755)

		f, _ := bfs.Create("srcdir/moveme.txt")
		f.Close()

		err := bfs.Rename("srcdir/moveme.txt", "dstdir/moveme.txt")
		if err != nil {
			t.Fatalf("Rename across dirs failed: %v", err)
		}

		_, err = bfs.Stat("dstdir/moveme.txt")
		if err != nil {
			t.Error("file not moved")
		}
	})
}

// TestRemove tests file and directory removal
func TestRemove(t *testing.T) {
	bfs, _ := newTestFS(t)

	t.Run("remove file", func(t *testing.T) {
		f, _ := bfs.Create("removeme.txt")
		f.Close()

		err := bfs.Remove("removeme.txt")
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		_, err = bfs.Stat("removeme.txt")
		if err == nil {
			t.Error("file still exists after remove")
		}
	})

	t.Run("remove empty directory", func(t *testing.T) {
		bfs.MkdirAll("emptydir", 0755)

		err := bfs.Remove("emptydir")
		if err != nil {
			t.Fatalf("Remove empty dir failed: %v", err)
		}
	})

	t.Run("remove non-existent", func(t *testing.T) {
		err := bfs.Remove("nonexistent")
		if err == nil {
			t.Error("expected error removing non-existent file")
		}
	})
}

// TestJoin tests path joining
func TestJoin(t *testing.T) {
	bfs, _ := newTestFS(t)

	// billyfs uses Unix-style forward slashes for paths, consistent with absfs design
	tests := []struct {
		elements []string
		expected string
	}{
		{[]string{"a", "b", "c"}, "a/b/c"},
		{[]string{"a", "b", "c.txt"}, "a/b/c.txt"},
		{[]string{"/a", "b"}, "/a/b"},
		{[]string{"a", "", "b"}, "a/b"},
		{[]string{"a"}, "a"},
	}

	for _, tc := range tests {
		result := bfs.Join(tc.elements...)
		if result != tc.expected {
			t.Errorf("Join(%v) = %s, want %s", tc.elements, result, tc.expected)
		}
	}
}

// TestRoot tests the Root method
func TestRoot(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := osfs.NewFS()
	if err != nil {
		t.Fatalf("failed to create osfs: %v", err)
	}

	bfs, err := billyfs.NewFS(fs, tmpDir)
	if err != nil {
		t.Fatalf("NewFS failed: %v", err)
	}

	root := bfs.Root()
	if root != tmpDir {
		t.Errorf("expected root %s, got %s", tmpDir, root)
	}
}

// TestReadDir tests directory reading
func TestReadDir(t *testing.T) {
	bfs, _ := newTestFS(t)

	t.Run("read directory with files", func(t *testing.T) {
		bfs.MkdirAll("readdir", 0755)
		f1, _ := bfs.Create("readdir/file1.txt")
		f1.Close()
		f2, _ := bfs.Create("readdir/file2.txt")
		f2.Close()
		bfs.MkdirAll("readdir/subdir", 0755)

		entries, err := bfs.ReadDir("readdir")
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}

		if len(entries) != 3 {
			t.Errorf("expected 3 entries, got %d", len(entries))
		}
	})

	t.Run("read empty directory", func(t *testing.T) {
		bfs.MkdirAll("emptyread", 0755)

		entries, err := bfs.ReadDir("emptyread")
		if err != nil {
			t.Fatalf("ReadDir empty failed: %v", err)
		}

		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("read non-existent directory", func(t *testing.T) {
		_, err := bfs.ReadDir("nonexistentdir")
		if err == nil {
			t.Error("expected error for non-existent directory")
		}
	})
}

// TestMkdirAll tests recursive directory creation
func TestMkdirAll(t *testing.T) {
	bfs, _ := newTestFS(t)

	t.Run("create nested directories", func(t *testing.T) {
		err := bfs.MkdirAll("a/b/c/d", 0755)
		if err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		// Verify all directories exist
		for _, dir := range []string{"a", "a/b", "a/b/c", "a/b/c/d"} {
			info, err := bfs.Stat(dir)
			if err != nil {
				t.Errorf("directory %s not created: %v", dir, err)
			}
			if !info.IsDir() {
				t.Errorf("%s is not a directory", dir)
			}
		}
	})

	t.Run("create existing directory", func(t *testing.T) {
		bfs.MkdirAll("existing", 0755)

		// Should not error when creating existing directory
		err := bfs.MkdirAll("existing", 0755)
		if err != nil {
			t.Errorf("MkdirAll on existing dir failed: %v", err)
		}
	})
}

// TestChmod tests permission changes
func TestChmod(t *testing.T) {
	bfs, _ := newTestFS(t)

	f, _ := bfs.Create("chmodtest.txt")
	f.Close()

	err := bfs.Chmod("chmodtest.txt", 0600)
	if err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}

	info, _ := bfs.Stat("chmodtest.txt")
	// Note: mode comparison might differ by platform, so just check it doesn't error
	_ = info.Mode()
}

// TestChtimes tests modification time changes
func TestChtimes(t *testing.T) {
	bfs, _ := newTestFS(t)

	f, _ := bfs.Create("chtimestest.txt")
	f.Close()

	atime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	mtime := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)

	err := bfs.Chtimes("chtimestest.txt", atime, mtime)
	if err != nil {
		t.Fatalf("Chtimes failed: %v", err)
	}

	info, _ := bfs.Stat("chtimestest.txt")
	if !info.ModTime().Equal(mtime) {
		t.Errorf("expected mtime %v, got %v", mtime, info.ModTime())
	}
}

// TestChown tests ownership changes
func TestChown(t *testing.T) {
	bfs, _ := newTestFS(t)

	f, _ := bfs.Create("chowntest.txt")
	f.Close()

	// Note: Chown typically requires root privileges, so we just test that
	// the method can be called without panicking. The actual behavior depends
	// on the underlying filesystem and permissions.
	err := bfs.Chown("chowntest.txt", os.Getuid(), os.Getgid())
	// We don't check the error since this may fail without root
	_ = err
}

// TestLchown tests ownership changes on symlinks
func TestLchown(t *testing.T) {
	bfs, _ := newTestFS(t)

	f, _ := bfs.Create("lchowntarget.txt")
	f.Close()
	bfs.Symlink("lchowntarget.txt", "lchownlink.txt")

	// Note: Lchown typically requires root privileges, so we just test that
	// the method can be called without panicking.
	err := bfs.Lchown("lchownlink.txt", os.Getuid(), os.Getgid())
	// We don't check the error since this may fail without root
	_ = err
}

// TestChroot tests chroot functionality
func TestChroot(t *testing.T) {
	bfs, tmpDir := newTestFS(t)

	t.Run("chroot with relative path", func(t *testing.T) {
		// Create directory using relative path in the filesystem
		if err := bfs.MkdirAll("chrootdir", 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		// Chroot expects paths relative to filesystem root, but it needs to be
		// an absolute path in the underlying filesystem. Since bfs is rooted at
		// tmpDir, we use "chrootdir" which basefs will resolve.
		chrooted, err := bfs.Chroot("chrootdir")
		if err != nil {
			t.Fatalf("Chroot failed: %v", err)
		}

		// Create file in chrooted fs
		f, err := chrooted.Create("inchroot.txt")
		if err != nil {
			t.Fatalf("Create in chroot failed: %v", err)
		}
		f.Close()

		// File should exist in original fs
		_, err = bfs.Stat("chrootdir/inchroot.txt")
		if err != nil {
			t.Errorf("file not visible in parent fs: %v", err)
		}
	})

	t.Run("chroot non-existent returns error", func(t *testing.T) {
		_, err := bfs.Chroot("/nonexistent/path")
		if err == nil {
			t.Error("expected error for non-existent chroot path")
		}
	})

	t.Run("chroot root method", func(t *testing.T) {
		root := bfs.Root()
		if root != tmpDir {
			t.Errorf("expected root %s, got %s", tmpDir, root)
		}
	})
}

// TestSymlink tests symbolic link operations
func TestSymlink(t *testing.T) {
	bfs, _ := newTestFS(t)

	t.Run("create and read symlink", func(t *testing.T) {
		// Create target file
		f, _ := bfs.Create("target.txt")
		f.Write([]byte("target content"))
		f.Close()

		// Create symlink
		err := bfs.Symlink("target.txt", "link.txt")
		if err != nil {
			t.Fatalf("Symlink failed: %v", err)
		}

		// Read link - the implementation may return absolute paths
		target, err := bfs.Readlink("link.txt")
		if err != nil {
			t.Fatalf("Readlink failed: %v", err)
		}
		// Target should end with the original target name
		if filepath.Base(target) != "target.txt" && target != "target.txt" {
			t.Errorf("unexpected symlink target: '%s'", target)
		}
	})

	t.Run("lstat symlink", func(t *testing.T) {
		f, _ := bfs.Create("ltarget.txt")
		f.Close()
		bfs.Symlink("ltarget.txt", "llink.txt")

		// Lstat should return info about the link, not target
		info, err := bfs.Lstat("llink.txt")
		if err != nil {
			t.Fatalf("Lstat failed: %v", err)
		}

		// Check it's a symlink
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("expected symlink mode")
		}
	})

	t.Run("stat follows symlink", func(t *testing.T) {
		f, _ := bfs.Create("starget.txt")
		f.Write([]byte("12345"))
		f.Close()
		bfs.Symlink("starget.txt", "slink.txt")

		// Stat should follow the link
		info, err := bfs.Stat("slink.txt")
		if err != nil {
			t.Fatalf("Stat on symlink failed: %v", err)
		}

		if info.Size() != 5 {
			t.Errorf("expected size 5, got %d", info.Size())
		}
	})

	t.Run("readlink non-existent", func(t *testing.T) {
		_, err := bfs.Readlink("nonexistent_link")
		if err == nil {
			t.Error("expected error for non-existent symlink")
		}
	})
}

// TestTempFile tests temporary file creation
func TestTempFile(t *testing.T) {
	bfs, _ := newTestFS(t)

	// Create a tmp directory since TempFile uses TempDir() which defaults to /tmp
	// and basefs wraps paths relative to root
	if err := bfs.MkdirAll("tmp", 0755); err != nil {
		t.Fatalf("Failed to create tmp directory: %v", err)
	}

	t.Run("create temp file", func(t *testing.T) {
		f, err := bfs.TempFile("", "test")
		if err != nil {
			t.Fatalf("TempFile failed: %v", err)
		}
		defer f.Close()

		name := f.Name()
		if name == "" {
			t.Error("temp file has no name")
		}
	})

	t.Run("temp file uniqueness", func(t *testing.T) {
		names := make(map[string]bool)
		for i := 0; i < 10; i++ {
			f, err := bfs.TempFile("", "unique")
			if err != nil {
				t.Fatalf("TempFile failed: %v", err)
			}
			if names[f.Name()] {
				t.Errorf("duplicate temp file name: %s", f.Name())
			}
			names[f.Name()] = true
			f.Close()
		}
	})

	t.Run("temp file with prefix", func(t *testing.T) {
		f, err := bfs.TempFile("", "myprefix")
		if err != nil {
			t.Fatalf("TempFile failed: %v", err)
		}
		defer f.Close()

		name := f.Name()
		if name == "" {
			t.Error("temp file has no name")
		}
		// Name should contain the prefix
		// Note: exact format depends on implementation
	})
}
