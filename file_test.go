package billyfs_test

import (
	"io"
	"os"
	"testing"

	"github.com/absfs/billyfs"
	"github.com/absfs/osfs"
)

// newFileTestFS creates a new test filesystem for file tests
func newFileTestFS(t *testing.T) *billyfs.Filesystem {
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

	return bfs
}

// TestFileName tests the File.Name method
func TestFileName(t *testing.T) {
	bfs := newFileTestFS(t)

	f, err := bfs.Create("testname.txt")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer f.Close()

	if f.Name() != "testname.txt" {
		t.Errorf("expected name 'testname.txt', got '%s'", f.Name())
	}
}

// TestFileWrite tests the File.Write method
func TestFileWrite(t *testing.T) {
	bfs := newFileTestFS(t)

	t.Run("write data", func(t *testing.T) {
		f, err := bfs.Create("writetest.txt")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		data := []byte("hello world")
		n, err := f.Write(data)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("expected to write %d bytes, wrote %d", len(data), n)
		}
		f.Close()

		// Verify written data
		f2, _ := bfs.Open("writetest.txt")
		defer f2.Close()
		readData, _ := io.ReadAll(f2)
		if string(readData) != "hello world" {
			t.Errorf("expected 'hello world', got '%s'", string(readData))
		}
	})

	t.Run("write empty", func(t *testing.T) {
		f, err := bfs.Create("writeempty.txt")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		defer f.Close()

		n, err := f.Write([]byte{})
		if err != nil {
			t.Fatalf("Write empty failed: %v", err)
		}
		if n != 0 {
			t.Errorf("expected to write 0 bytes, wrote %d", n)
		}
	})

	t.Run("multiple writes", func(t *testing.T) {
		f, _ := bfs.Create("multiwrite.txt")

		f.Write([]byte("first"))
		f.Write([]byte(" second"))
		f.Write([]byte(" third"))
		f.Close()

		f2, _ := bfs.Open("multiwrite.txt")
		defer f2.Close()
		data, _ := io.ReadAll(f2)
		if string(data) != "first second third" {
			t.Errorf("expected 'first second third', got '%s'", string(data))
		}
	})
}

// TestFileRead tests the File.Read method
func TestFileRead(t *testing.T) {
	bfs := newFileTestFS(t)

	t.Run("read data", func(t *testing.T) {
		// Setup: create file with content
		f, _ := bfs.Create("readtest.txt")
		f.Write([]byte("test data"))
		f.Close()

		// Test read
		f2, err := bfs.Open("readtest.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		defer f2.Close()

		buf := make([]byte, 4)
		n, err := f2.Read(buf)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if n != 4 {
			t.Errorf("expected to read 4 bytes, read %d", n)
		}
		if string(buf) != "test" {
			t.Errorf("expected 'test', got '%s'", string(buf))
		}
	})

	t.Run("read EOF", func(t *testing.T) {
		f, _ := bfs.Create("eoftest.txt")
		f.Write([]byte("x"))
		f.Close()

		f2, _ := bfs.Open("eoftest.txt")
		defer f2.Close()

		buf := make([]byte, 10)
		n, err := f2.Read(buf)
		if n != 1 {
			t.Errorf("expected to read 1 byte, read %d", n)
		}

		// Second read should return EOF
		_, err = f2.Read(buf)
		if err != io.EOF {
			t.Errorf("expected EOF, got %v", err)
		}
	})

	t.Run("read empty file", func(t *testing.T) {
		f, _ := bfs.Create("emptyread.txt")
		f.Close()

		f2, _ := bfs.Open("emptyread.txt")
		defer f2.Close()

		buf := make([]byte, 10)
		n, err := f2.Read(buf)
		if err != io.EOF {
			t.Errorf("expected EOF for empty file, got %v", err)
		}
		if n != 0 {
			t.Errorf("expected 0 bytes, got %d", n)
		}
	})
}

// TestFileReadAt tests the File.ReadAt method
func TestFileReadAt(t *testing.T) {
	bfs := newFileTestFS(t)

	// Setup
	f, _ := bfs.Create("readat.txt")
	f.Write([]byte("0123456789"))
	f.Close()

	t.Run("read at offset 0", func(t *testing.T) {
		f2, _ := bfs.Open("readat.txt")
		defer f2.Close()

		buf := make([]byte, 3)
		n, err := f2.ReadAt(buf, 0)
		if err != nil {
			t.Fatalf("ReadAt failed: %v", err)
		}
		if n != 3 || string(buf) != "012" {
			t.Errorf("expected '012', got '%s'", string(buf))
		}
	})

	t.Run("read at middle offset", func(t *testing.T) {
		f2, _ := bfs.Open("readat.txt")
		defer f2.Close()

		buf := make([]byte, 3)
		n, err := f2.ReadAt(buf, 5)
		if err != nil {
			t.Fatalf("ReadAt failed: %v", err)
		}
		if n != 3 || string(buf) != "567" {
			t.Errorf("expected '567', got '%s'", string(buf))
		}
	})

	t.Run("read at end", func(t *testing.T) {
		f2, _ := bfs.Open("readat.txt")
		defer f2.Close()

		buf := make([]byte, 5)
		n, _ := f2.ReadAt(buf, 8)
		// Should read 2 bytes and get EOF
		if n != 2 {
			t.Errorf("expected 2 bytes, got %d", n)
		}
		if string(buf[:n]) != "89" {
			t.Errorf("expected '89', got '%s'", string(buf[:n]))
		}
	})
}

// TestFileWriteAt tests the File.WriteAt method
func TestFileWriteAt(t *testing.T) {
	bfs := newFileTestFS(t)

	t.Run("write at offset", func(t *testing.T) {
		// Create initial file
		f, _ := bfs.Create("writeat.txt")
		f.Write([]byte("0123456789"))
		f.Close()

		// Open with read-write to use WriteAt
		f2, _ := bfs.OpenFile("writeat.txt", os.O_RDWR, 0644)
		defer f2.Close()

		// Type assert to access WriteAt method
		writerAt, ok := f2.(io.WriterAt)
		if !ok {
			t.Skip("File does not implement io.WriterAt")
		}

		n, err := writerAt.WriteAt([]byte("XXX"), 3)
		if err != nil {
			t.Fatalf("WriteAt failed: %v", err)
		}
		if n != 3 {
			t.Errorf("expected to write 3 bytes, wrote %d", n)
		}
		f2.Close()

		// Verify
		f3, _ := bfs.Open("writeat.txt")
		defer f3.Close()
		data, _ := io.ReadAll(f3)
		if string(data) != "012XXX6789" {
			t.Errorf("expected '012XXX6789', got '%s'", string(data))
		}
	})

	t.Run("write at beginning", func(t *testing.T) {
		f, _ := bfs.Create("writeatbegin.txt")
		f.Write([]byte("AAAAA"))
		f.Close()

		f2, _ := bfs.OpenFile("writeatbegin.txt", os.O_RDWR, 0644)
		defer f2.Close()

		writerAt, ok := f2.(io.WriterAt)
		if !ok {
			t.Skip("File does not implement io.WriterAt")
		}

		n, err := writerAt.WriteAt([]byte("BB"), 0)
		if err != nil {
			t.Fatalf("WriteAt failed: %v", err)
		}
		if n != 2 {
			t.Errorf("expected to write 2 bytes, wrote %d", n)
		}
		f2.Close()

		f3, _ := bfs.Open("writeatbegin.txt")
		defer f3.Close()
		data, _ := io.ReadAll(f3)
		if string(data) != "BBAAA" {
			t.Errorf("expected 'BBAAA', got '%s'", string(data))
		}
	})
}

// TestFileSeek tests the File.Seek method
func TestFileSeek(t *testing.T) {
	bfs := newFileTestFS(t)

	// Setup
	f, _ := bfs.Create("seektest.txt")
	f.Write([]byte("0123456789"))
	f.Close()

	t.Run("SEEK_SET", func(t *testing.T) {
		f2, _ := bfs.Open("seektest.txt")
		defer f2.Close()

		pos, err := f2.Seek(5, io.SeekStart)
		if err != nil {
			t.Fatalf("Seek failed: %v", err)
		}
		if pos != 5 {
			t.Errorf("expected position 5, got %d", pos)
		}

		buf := make([]byte, 3)
		f2.Read(buf)
		if string(buf) != "567" {
			t.Errorf("expected '567', got '%s'", string(buf))
		}
	})

	t.Run("SEEK_CUR", func(t *testing.T) {
		f2, _ := bfs.Open("seektest.txt")
		defer f2.Close()

		f2.Seek(2, io.SeekStart)
		pos, err := f2.Seek(3, io.SeekCurrent)
		if err != nil {
			t.Fatalf("Seek failed: %v", err)
		}
		if pos != 5 {
			t.Errorf("expected position 5, got %d", pos)
		}
	})

	t.Run("SEEK_END", func(t *testing.T) {
		f2, _ := bfs.Open("seektest.txt")
		defer f2.Close()

		pos, err := f2.Seek(-3, io.SeekEnd)
		if err != nil {
			t.Fatalf("Seek failed: %v", err)
		}
		if pos != 7 {
			t.Errorf("expected position 7, got %d", pos)
		}

		buf := make([]byte, 3)
		f2.Read(buf)
		if string(buf) != "789" {
			t.Errorf("expected '789', got '%s'", string(buf))
		}
	})
}

// TestFileTruncate tests the File.Truncate method
func TestFileTruncate(t *testing.T) {
	bfs := newFileTestFS(t)

	t.Run("truncate to smaller", func(t *testing.T) {
		f, _ := bfs.Create("truncsmaller.txt")
		f.Write([]byte("1234567890"))

		err := f.Truncate(5)
		if err != nil {
			t.Fatalf("Truncate failed: %v", err)
		}
		f.Close()

		f2, _ := bfs.Open("truncsmaller.txt")
		defer f2.Close()
		data, _ := io.ReadAll(f2)
		if len(data) != 5 || string(data) != "12345" {
			t.Errorf("expected '12345', got '%s'", string(data))
		}
	})

	t.Run("truncate to zero", func(t *testing.T) {
		f, _ := bfs.Create("trunczero.txt")
		f.Write([]byte("content"))

		err := f.Truncate(0)
		if err != nil {
			t.Fatalf("Truncate failed: %v", err)
		}
		f.Close()

		info, _ := bfs.Stat("trunczero.txt")
		if info.Size() != 0 {
			t.Errorf("expected size 0, got %d", info.Size())
		}
	})

	t.Run("truncate to larger", func(t *testing.T) {
		f, _ := bfs.Create("trunclarger.txt")
		f.Write([]byte("abc"))

		err := f.Truncate(10)
		if err != nil {
			t.Fatalf("Truncate failed: %v", err)
		}
		f.Close()

		info, _ := bfs.Stat("trunclarger.txt")
		if info.Size() != 10 {
			t.Errorf("expected size 10, got %d", info.Size())
		}
	})
}

// TestFileClose tests the File.Close method
func TestFileClose(t *testing.T) {
	bfs := newFileTestFS(t)

	t.Run("close file", func(t *testing.T) {
		f, _ := bfs.Create("closetest.txt")

		err := f.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	})

	t.Run("write after close fails", func(t *testing.T) {
		f, _ := bfs.Create("writeafterclose.txt")
		f.Close()

		_, err := f.Write([]byte("test"))
		if err == nil {
			t.Error("expected error writing to closed file")
		}
	})
}

// TestFileLock tests the File.Lock and Unlock methods
func TestFileLock(t *testing.T) {
	bfs := newFileTestFS(t)

	f, _ := bfs.Create("locktest.txt")
	defer f.Close()

	t.Run("lock file", func(t *testing.T) {
		err := f.Lock()
		if err != nil {
			t.Fatalf("Lock failed: %v", err)
		}
	})

	t.Run("unlock file", func(t *testing.T) {
		err := f.Unlock()
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}
	})

	t.Run("lock and unlock", func(t *testing.T) {
		err := f.Lock()
		if err != nil {
			t.Fatalf("Lock failed: %v", err)
		}

		err = f.Unlock()
		if err != nil {
			t.Fatalf("Unlock failed: %v", err)
		}
	})
}
