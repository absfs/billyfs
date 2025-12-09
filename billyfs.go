package billyfs

import (
	"math/rand"
	"os"
	"path"
	"sync"
	"time"

	"github.com/absfs/absfs"
	"github.com/absfs/basefs"
	"github.com/go-git/go-billy/v5"
)

var (
	rng     *rand.Rand
	rngMu   sync.Mutex
	rngOnce sync.Once
)

func initRNG() {
	rngOnce.Do(func() {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
}

// Filesystem implements all functions of the go-billy Filesystem interface
// by using the absfs.FileSystem interface.
type Filesystem struct {
	fs absfs.SymlinkFileSystem
}

// NewFS wraps a absfs.FileSystem go-billy  from a `absfs.FileSystem` compatible object
// and a path. The path must be an absolute path and must already exist in the
// fs provided otherwise an error is returned.
func NewFS(fs absfs.SymlinkFileSystem, dir string) (*Filesystem, error) {
	//if dir == "" {
	//	return nil, os.ErrInvalid
	//}

	// if !path.IsAbs(dir) {
	// 	return nil, errors.New("not an absolute path")
	// }
	// info, err := fs.Stat(dir)
	// if err != nil {
	// 	return nil, err
	// }
	// if !info.IsDir() {
	// 	return nil, errors.New("not a directory")
	// }

	fs, err := basefs.NewFS(fs, dir)
	if err != nil {
		return nil, err
	}

	return &Filesystem{fs: fs}, nil
}

// go-billy Basic interface functions

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
func (f *Filesystem) Create(filename string) (billy.File, error) {
	file, err := f.fs.Create(filename)
	if err != nil {
		return nil, err
	}
	return &File{f: file}, nil
}

// Open opens the named file for reading. If successful, methods on the
// returned file can be used for reading; the associated file descriptor has
// mode O_RDONLY.
func (f *Filesystem) Open(filename string) (billy.File, error) {
	file, err := f.fs.Open(filename)
	if err != nil {
		return nil, err
	}
	return &File{f: file}, nil
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (f *Filesystem) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	file, err := f.fs.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	return &File{f: file}, nil
}

// Stat returns a FileInfo describing the named file.
func (f *Filesystem) Stat(filename string) (os.FileInfo, error) {
	return f.fs.Stat(filename)
}

// Rename renames (moves) oldpath to newpath. If newpath already exists and
// is not a directory, Rename replaces it. OS-specific restrictions may
// apply when oldpath and newpath are in different directories.
func (f *Filesystem) Rename(oldpath, newpath string) error {
	return f.fs.Rename(oldpath, newpath)
}

// Remove removes the named file or directory.
func (f *Filesystem) Remove(filename string) error {
	return f.fs.Remove(filename)
}

// Join joins any number of path elements into a single path, adding a
// Separator if necessary. Join calls filepath.Clean on the result; in
// particular, all empty strings are ignored. On Windows, the result is a
// UNC path if and only if the first path element is a UNC path.
func (f *Filesystem) Join(elem ...string) string {
	return path.Join(elem...)
}

// go-billy Capabilities interface

// Capabilities returns the features supported by a filesystem. Absfs supports
// all capabilities.
func (f *Filesystem) Capabilities() billy.Capability {
	return billy.AllCapabilities
}

// go-billy Change interface functions

// Chmod changes the mode of the named file to mode. If the file is a
// symbolic link, it changes the mode of the link's target.
func (f *Filesystem) Chmod(name string, mode os.FileMode) error {
	return f.fs.Chmod(name, mode)
}

// Lchown changes the numeric uid and gid of the named file. If the file is
// a symbolic link, it changes the uid and gid of the link itself.
func (f *Filesystem) Lchown(name string, uid, gid int) error {
	return f.fs.Lchown(name, uid, gid)
}

// Chown changes the numeric uid and gid of the named file. If the file is a
// symbolic link, it changes the uid and gid of the link's target.
func (f *Filesystem) Chown(name string, uid, gid int) error {
	return f.fs.Chown(name, uid, gid)
}

// Chtimes changes the access and modification times of the named file,
// similar to the Unix utime() or utimes() functions.
//
// The underlying filesystem may truncate or round the values to a less
// precise time unit.
func (f *Filesystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return f.fs.Chtimes(name, atime, mtime)
}

// go-billy Chroot interface functions

// Chroot returns a new filesystem from the same type where the new root is
// the given path. Files outside of the designated directory tree cannot be
// accessed.
func (f *Filesystem) Chroot(name string) (billy.Filesystem, error) {
	// Convert the path to an absolute path using the filesystem's root prefix
	// since basefs.NewFS requires an absolute path
	var absPath string
	if path.IsAbs(name) {
		absPath = name
	} else {
		// Get the current root and join with the relative path
		// basefs cwd is always "/" so we use path.Join instead of filepath.Join
		cwd, err := f.fs.Getwd()
		if err != nil {
			return &Filesystem{}, err
		}
		prefix := basefs.Prefix(f.fs)
		// Join cwd and name using path (not filepath) since basefs uses "/" internally
		relPath := path.Join(cwd, name)
		// Now convert to absolute using the prefix
		absPath = path.Join(prefix, relPath)
	}

	// Unwrap to get the underlying filesystem to avoid double-wrapping
	underlying := basefs.Unwrap(f.fs)
	// Type assert to SymlinkFileSystem
	symlinkFS, ok := underlying.(absfs.SymlinkFileSystem)
	if !ok {
		// If underlying doesn't support symlinks, use the wrapped filesystem
		symlinkFS = f.fs
	}

	fs, err := basefs.NewFS(symlinkFS, absPath)
	if err != nil {
		return &Filesystem{}, err
	}

	return &Filesystem{fs}, nil
}

// Root returns the root path of the filesystem.
func (f *Filesystem) Root() string {
	path := basefs.Prefix(f.fs)
	if path == "" {
		return "/"
	}
	return path
}

// go-billy Dir interface functions

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (f *Filesystem) ReadDir(path string) ([]os.FileInfo, error) {
	// open directory at path and read all files
	dir, err := f.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	return dir.Readdir(0)
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error. The permission bits
// perm are used for all directories that MkdirAll creates. If path is/
// already a directory, MkdirAll does nothing and returns nil.
func (f *Filesystem) MkdirAll(filename string, perm os.FileMode) error {
	return f.fs.MkdirAll(filename, perm)
}

// go-billy Symlink interface functions

// Lstat returns a FileInfo describing the named file. If the file is a
// symbolic link, the returned FileInfo describes the symbolic link. Lstat
// makes no attempt to follow the link.
func (f *Filesystem) Lstat(filename string) (os.FileInfo, error) {
	return f.fs.Lstat(filename)
}

// Symlink creates a symbolic-link from link to target. target may be an
// absolute or relative path, and need not refer to an existing node.
// Parent directories of link are created as necessary.
func (f *Filesystem) Symlink(target, link string) error {
	return f.fs.Symlink(target, link)
}

// Readlink returns the target path of link.
func (f *Filesystem) Readlink(link string) (string, error) {
	return f.fs.Readlink(link)
}

// go-billy TempFile interface functions

// TempFile creates a new temporary file in the directory dir with a name
// beginning with prefix, opens the file for reading and writing, and
// returns the resulting *os.File. If dir is the empty string, TempFile
// uses the default directory for temporary files (see os.TempDir).
// Multiple programs calling TempFile simultaneously will not choose the
// same file. The caller can use f.Name() to find the pathname of the file.
// It is the caller's responsibility to remove the file when no longer
// needed.
func (f *Filesystem) TempFile(dir string, prefix string) (billy.File, error) {
	// get the temp directory, then create a temp file
	initRNG()
	p := path.Join(f.fs.TempDir(), prefix+"_"+randSeq(5))
	file, err := f.fs.Create(p)
	if err != nil {
		return nil, err
	}
	return &File{f: file}, nil
}

// randSeq generates a random string of length n
func randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	rngMu.Lock()
	defer rngMu.Unlock()
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return string(b)
}
