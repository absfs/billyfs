package billyfs_test

import (
	"bytes"
	"io"
	"path"
	"testing"

	"github.com/absfs/billyfs"
	"github.com/absfs/osfs"
)

// newFuzzFS creates a filesystem for fuzz testing
func newFuzzFS(t testing.TB) *billyfs.Filesystem {
	t.Helper()
	tmpDir := t.(*testing.T).TempDir()

	fs, err := osfs.NewFS()
	if err != nil {
		t.Fatalf("failed to create osfs: %v", err)
	}

	bfs, err := billyfs.NewFS(fs, tmpDir)
	if err != nil {
		t.Fatalf("failed to create billyfs: %v", err)
	}

	return bfs
}

// FuzzCreateFile tests file creation with random paths
func FuzzCreateFile(f *testing.F) {
	// Seed corpus with typical and edge-case paths
	f.Add("file.txt")
	f.Add("a/b/c.txt")
	f.Add("test")
	f.Add(".hidden")
	f.Add("path/with spaces/file.txt")
	f.Add("deep/nested/path/to/file.txt")
	f.Add("UPPER/lower/MiXeD.txt")

	f.Fuzz(func(t *testing.T, name string) {
		if len(name) == 0 {
			return
		}
		// Skip paths with null bytes or that are too long
		for _, c := range name {
			if c == 0 {
				return
			}
		}
		if len(name) > 255 {
			return
		}

		bfs := newFuzzFS(t)

		// Create parent directories if path has directory components
		dir := path.Dir(name)
		if dir != "." && dir != "" {
			bfs.MkdirAll(dir, 0755)
		}

		// Create and write to file — errors are expected for many inputs
		file, err := bfs.Create(name)
		if err != nil {
			return
		}
		defer file.Close()

		_, err = file.Write([]byte("test data"))
		if err != nil {
			t.Errorf("failed to write to created file %q: %v", name, err)
		}
	})
}

// FuzzReadWrite tests read/write operations with random data
func FuzzReadWrite(f *testing.F) {
	// Seed corpus with various data patterns
	f.Add([]byte("hello world"))
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte(""))
	f.Add([]byte("a very long string that might cause buffer issues if not handled properly"))
	f.Add(bytes.Repeat([]byte("x"), 4096))

	f.Fuzz(func(t *testing.T, data []byte) {
		bfs := newFuzzFS(t)

		// Create file and write data
		file, err := bfs.Create("test.bin")
		if err != nil {
			t.Fatal(err)
		}

		n, err := file.Write(data)
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}
		if n != len(data) {
			t.Fatalf("expected to write %d bytes, wrote %d", len(data), n)
		}
		file.Close()

		// Read it back
		file, err = bfs.Open("test.bin")
		if err != nil {
			t.Fatal(err)
		}

		readBuf, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}

		if !bytes.Equal(readBuf, data) {
			t.Fatalf("data mismatch: wrote %d bytes, read %d bytes", len(data), len(readBuf))
		}
	})
}

// FuzzSymlink tests symlink creation with random targets and paths
func FuzzSymlink(f *testing.F) {
	f.Add("target.txt", "link")
	f.Add("a/b/c", "x/y/z")
	f.Add("../relative", "link2")
	f.Add("file.txt", "dir/link")

	f.Fuzz(func(t *testing.T, target, linkPath string) {
		if len(linkPath) == 0 || len(target) == 0 {
			return
		}
		// Skip null bytes
		for _, c := range linkPath {
			if c == 0 {
				return
			}
		}
		for _, c := range target {
			if c == 0 {
				return
			}
		}
		if len(linkPath) > 255 || len(target) > 255 {
			return
		}

		bfs := newFuzzFS(t)

		// Create parent directory for link
		dir := path.Dir(linkPath)
		if dir != "." && dir != "" {
			bfs.MkdirAll(dir, 0755)
		}

		// Create symlink — errors are expected for many inputs
		err := bfs.Symlink(target, linkPath)
		if err != nil {
			return
		}

		// Verify we can read it back
		readTarget, err := bfs.Readlink(linkPath)
		if err != nil {
			t.Errorf("failed to readlink %q: %v", linkPath, err)
			return
		}

		if readTarget != target {
			t.Errorf("symlink target mismatch: got %q, want %q", readTarget, target)
		}
	})
}
