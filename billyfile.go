package billyfs

import (
	"fmt"
	"sync"

	"github.com/absfs/absfs"
)

// File implements the billy.File interface by using the absfs.File interface.
type File struct {
	f  absfs.File
	mu sync.Mutex
}

func (f *File) Name() string {
	return f.f.Name()
}

// io.Writer interface
func (f *File) Write(p []byte) (n int, err error) {
	n, err = f.f.Write(p)
	if err != nil {
		return n, fmt.Errorf("write %s: %w", f.f.Name(), err)
	}
	return n, nil
}

// io.Reader interface
func (f *File) Read(p []byte) (n int, err error) {
	n, err = f.f.Read(p)
	if err != nil {
		return n, fmt.Errorf("read %s: %w", f.f.Name(), err)
	}
	return n, nil
}

// io.ReaderAt interface
func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = f.f.ReadAt(p, off)
	if err != nil {
		return n, fmt.Errorf("readat %s (offset=%d): %w", f.f.Name(), off, err)
	}
	return n, nil
}

// io.WriterAt interface
func (f *File) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = f.f.WriteAt(p, off)
	if err != nil {
		return n, fmt.Errorf("writeat %s (offset=%d): %w", f.f.Name(), off, err)
	}
	return n, nil
}

// io.Seeker interface
func (f *File) Seek(offset int64, whence int) (int64, error) {
	pos, err := f.f.Seek(offset, whence)
	if err != nil {
		return pos, fmt.Errorf("seek %s (offset=%d, whence=%d): %w", f.f.Name(), offset, whence, err)
	}
	return pos, nil
}

// io.Closer interface
func (f *File) Close() error {
	if err := f.f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", f.f.Name(), err)
	}
	return nil
}

// Truncate the file.
func (f *File) Truncate(size int64) error {
	if err := f.f.Truncate(size); err != nil {
		return fmt.Errorf("truncate %s to %d bytes: %w", f.f.Name(), size, err)
	}
	return nil
}

func (f *File) Lock() error {
	f.mu.Lock()
	return nil
}

func (f *File) Unlock() error {
	f.mu.Unlock()
	return nil
}
