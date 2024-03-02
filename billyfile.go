package billyfs

import (
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
	return f.f.Write(p)
}

// io.Reader interface
func (f *File) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

// io.ReaderAt interface
func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	return f.f.ReadAt(p, off)
}

// io.WriterAt interface
func (f *File) WriteAt(p []byte, off int64) (n int, err error) {
	return f.f.WriteAt(p, off)
}

// io.Seeker interface
func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.f.Seek(offset, whence)
}

// io.Closer interface
func (f *File) Close() error {
	return f.f.Close()
}

// Truncate the file.
func (f *File) Truncate(size int64) error {
	return f.f.Truncate(size)
}

func (f *File) Lock() error {
	f.mu.Lock()
	return nil
}

func (f *File) Unlock() error {
	f.mu.Unlock()
	return nil
}
