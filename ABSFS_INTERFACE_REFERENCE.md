# ABSFS Filesystem Interface Standards

## Overview
ABSFS (Abstract File System) is a Go package that provides abstract interfaces for implementing custom filesystems. It defines a comprehensive set of interfaces that enable implementations to support common filesystem operations while maintaining compatibility across different filesystem backends.

**Key Reference**: billyfs is an adapter that converts `absfs.SymlinkFileSystem` implementations to `go-billy/v5` compatible filesystems.

---

## Interface Hierarchy

```
ABSFS Interface Architecture:

UnSeekable (Base file operations)
    ↓
Seekable (UnSeekable + Seek)
    ↓
File (Seekable + ReadAt/WriteAt/WriteString/Truncate/Readdirnames)

ReadOnlyFiler
WriteOnlyFiler
Filer (Core filesystem operations)
    ↓
FileSystem (Filer + convenience methods + path operations)
    ↓
SymLinker
    ↓
SymlinkFileSystem (FileSystem + SymLinker)
```

---

## Core Interfaces

### 1. UnSeekable Interface (Base File Operations)
The foundation for all file handle implementations. Supports basic read/write operations.

**Required Methods:**
```go
type UnSeekable interface {
    Name() string                          // Returns the file path as presented to Open
    Read(b []byte) (int, error)            // Read up to len(b) bytes
    Write(b []byte) (int, error)           // Write len(b) bytes
    Close() error                          // Close the file
    Sync() error                           // Commit contents to stable storage
    Stat() (os.FileInfo, error)            // Get file info
    Readdir(int) ([]os.FileInfo, error)    // Read directory entries
}
```

### 2. Seekable Interface (File Positioning)
Extends UnSeekable with file positioning capabilities.

**Additional Methods:**
```go
type Seekable interface {
    UnSeekable
    io.Seeker  // Seek(offset int64, whence int) (int64, error)
}
```

### 3. File Interface (Complete File Operations)
The complete file interface with all advanced operations.

**Additional Methods:**
```go
type File interface {
    Seekable   // Inherits all UnSeekable + Seek methods
    
    ReadAt(b []byte, off int64) (n int, err error)      // Read at specific offset
    WriteAt(b []byte, off int64) (n int, err error)     // Write at specific offset
    WriteString(s string) (n int, err error)            // Write string efficiently
    Truncate(size int64) error                          // Resize file
    Readdirnames(n int) (names []string, err error)     // Read directory names only
}
```

### 4. Filer Interface (Core Filesystem Operations)
The minimal filesystem interface for basic file operations.

**Required Methods:**
```go
type Filer interface {
    OpenFile(name string, flag int, perm os.FileMode) (File, error)
    Mkdir(name string, perm os.FileMode) error
    Remove(name string) error
    Rename(oldpath, newpath string) error
    Stat(name string) (os.FileInfo, error)
    Chmod(name string, mode os.FileMode) error
    Chtimes(name string, atime time.Time, mtime time.Time) error
    Chown(name string, uid, gid int) error
}
```

### 5. FileSystem Interface (Full Filesystem Operations)
Extends Filer with convenience methods and path operations.

**Additional Methods:**
```go
type FileSystem interface {
    Filer  // Inherits all Filer methods
    
    Separator() uint8          // Path separator (usually '/')
    ListSeparator() uint8      // Path list separator (usually ':')
    Chdir(dir string) error    // Change working directory
    Getwd() (dir string, err error)  // Get working directory
    TempDir() string           // Get temp directory path
    Open(name string) (File, error)     // Convenience: OpenFile(name, O_RDONLY, 0)
    Create(name string) (File, error)   // Convenience: OpenFile(name, O_CREATE|O_TRUNC|O_WRONLY, 0666)
    MkdirAll(name string, perm os.FileMode) error  // Create directories recursively
    RemoveAll(path string) error        // Remove recursively
    Truncate(name string, size int64) error  // Truncate file by path
}
```

### 6. SymLinker Interface (Symbolic Link Operations)
Defines symbolic link handling.

**Required Methods:**
```go
type SymLinker interface {
    Lstat(name string) (os.FileInfo, error)     // Stat without following symlinks
    Lchown(name string, uid, gid int) error     // Chown the link itself, not target
    Readlink(name string) (string, error)       // Get symlink target
    Symlink(oldname, newname string) error      // Create symbolic link
}
```

### 7. SymlinkFileSystem Interface (Complete Filesystem)
The complete filesystem interface combining FileSystem and SymLinker.

```go
type SymlinkFileSystem interface {
    FileSystem
    SymLinker
}
```

### 8. Optional Role-Based Interfaces
For read-only or write-only implementations:

```go
type ReadOnlyFiler interface {
    Open(name string) (io.ReadCloser, error)
}

type WriteOnlyFiler interface {
    Open(name string) (io.WriteCloser, error)
}
```

---

## File Opening Flags (OpenFile)

The absfs package uses standard os package flags with aliases:

```go
const (
    // Access modes (mutually exclusive - must specify exactly one)
    O_RDONLY = os.O_RDONLY  // 0 - Open read-only
    O_WRONLY = os.O_WRONLY  // 1 - Open write-only
    O_RDWR   = os.O_RDWR    // 2 - Open read-write
    
    // Behavior flags (may be OR'd together)
    O_APPEND = os.O_APPEND  // Append data when writing
    O_CREATE = os.O_CREATE  // Create file if not exists
    O_EXCL   = os.O_EXCL    // With CREATE: fail if file exists
    O_SYNC   = os.O_SYNC    // Open for synchronous I/O
    O_TRUNC  = os.O_TRUNC   // Truncate file when opened
    
    // Utility mask
    O_ACCESS = 0x3          // Mask for access mode bits
)
```

**Example Usage:**
```go
// Create and truncate
file, err := fs.OpenFile("test.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

// Open for append
file, err := fs.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)

// Open read-only
file, err := fs.OpenFile("readonly.txt", os.O_RDONLY, 0)
```

---

## Permission Flags

For mode parameters (os.FileMode):

```go
const (
    // Base permission bits
    OS_READ  = 04    // Read
    OS_WRITE = 02    // Write
    OS_EX    = 01    // Execute
    
    // Shift amounts
    OS_USER_SHIFT  = 6   // Owner shift
    OS_GROUP_SHIFT = 3   // Group shift
    OS_OTH_SHIFT   = 0   // Other shift
    
    // Combined permissions
    OS_USER_RWX = 0700   // Owner read/write/execute
    OS_GROUP_RWX = 0070  // Group read/write/execute
    OS_OTH_RWX = 0007    // Other read/write/execute
)
```

**Common Mode Values:**
```go
0644  // User: rw, Group: r, Other: r (typical file)
0755  // User: rwx, Group: rx, Other: rx (typical directory)
0600  // User: rw, Group: none, Other: none (private file)
0700  // User: rwx, Group: none, Other: none (private directory)
```

---

## Error Handling Standards

All methods should return errors of specific types where applicable:

- **PathError** (`*os.PathError`): For single-file operations (Stat, Chmod, etc.)
- **LinkError** (`*os.LinkError`): For operations involving two paths (Rename, Symlink, Link)
- **standard errors**: For general errors

**Example:**
```go
// If Stat fails, return *os.PathError
info, err := fs.Stat(path)
if err != nil {
    // err should be *os.PathError
}

// If Rename fails, return *os.LinkError
err := fs.Rename(oldpath, newpath)
if err != nil {
    // err should be *os.LinkError
}
```

---

## Helper Utilities (basefs Package)

The `basefs` package provides utilities for wrapping filesystems:

### basefs.NewFS() - Path-Based Wrapper
Creates a SymlinkFileSystem wrapper that operates within a specific directory:

```go
// Original filesystem
fs, err := osfs.NewFS()

// Wrap with basefs to operate only within /usr/local/src
wrapped, err := basefs.NewFS(fs, "/usr/local/src")
```

**Benefits:**
- Isolates filesystem operations to a specific directory
- All paths are relative to the base directory
- Used by billyfs for chroot functionality

### Helper Functions
```go
// Get the base prefix of a wrapped filesystem
prefix := basefs.Prefix(fs)  // Returns "/usr/local/src" or ""

// Unwrap to get underlying filesystem
original := basefs.Unwrap(fs)  // Returns original fs if wrapped
```

---

## Extension Functions

### ExtendFiler() - Add FileSystem Convenience Methods
Wraps a Filer implementation to add FileSystem convenience functions:

```go
// Adds Open(), Create(), MkdirAll(), RemoveAll(), Truncate()
fs := absfs.ExtendFiler(myFiler)
```

### ExtendSeekable() - Add File Operations
Extends a Seekable file to add full File interface:

```go
// Adds ReadAt(), WriteAt(), WriteString(), Truncate(), Readdirnames()
file := absfs.ExtendSeekable(mySeekable)
```

---

## Types and Utilities

### Flags Type
```go
type Flags int

// Parse string representation
flags, err := absfs.ParseFlags("O_RDONLY|O_CREATE|O_EXCL")
// Returns flags with those values set

// Convert back to string
str := flags.String()  // "O_RDONLY|O_CREATE|O_EXCL"
```

### FileMode Parsing
```go
// Parse Unix-style mode strings
mode, err := absfs.ParseFileMode("rwxr-xr-x")  // Returns 0755
mode, err := absfs.ParseFileMode("rw-r--r--")  // Returns 0644
```

### InvalidFile
A no-op File implementation that returns errors on all operations, similar to what os.Open returns on error. Used internally by implementations when file operations fail.

### FastWalkFunc
```go
type FastWalkFunc func(string, os.FileMode) error
```
Function signature for fast directory traversal (available in SymlinkFileSystem).

---

## Implementation Requirements Checklist

### Minimum: Filer Implementation
- [ ] OpenFile() - Core file opening
- [ ] Mkdir() - Create single directory
- [ ] Remove() - Remove file/empty directory
- [ ] Rename() - Rename/move file
- [ ] Stat() - Get file information
- [ ] Chmod() - Change file mode
- [ ] Chtimes() - Change access/modification times
- [ ] Chown() - Change owner/group

### FileSystem Level (Filer + Convenience)
- [ ] All Filer methods
- [ ] Separator() - Return '/' or '\\'
- [ ] ListSeparator() - Return ':' or ';'
- [ ] Chdir() - Change working directory
- [ ] Getwd() - Get current working directory
- [ ] TempDir() - Get temp directory location
- [ ] Open() - Simple read operation
- [ ] Create() - Simple create/write operation
- [ ] MkdirAll() - Recursive directory creation
- [ ] RemoveAll() - Recursive deletion
- [ ] Truncate() - Resize file by path

### File Operations
- [ ] Name() - Return file path
- [ ] Read() - Sequential reading
- [ ] Write() - Sequential writing
- [ ] Close() - Release file handle
- [ ] Sync() - Flush to storage
- [ ] Stat() - Get file info
- [ ] Readdir() - List directory entries
- [ ] Seek() - Change file position
- [ ] ReadAt() - Read at offset
- [ ] WriteAt() - Write at offset
- [ ] WriteString() - Write string
- [ ] Truncate() - Resize file handle
- [ ] Readdirnames() - Get directory names only

### SymLinker Level (for full filesystem)
- [ ] Lstat() - Stat without following symlinks
- [ ] Lchown() - Chown the link itself
- [ ] Readlink() - Get symlink target
- [ ] Symlink() - Create symbolic link

---

## Implementation Patterns

### 1. Delegation Pattern (billyfs example)
Create wrapper types that delegate to an underlying absfs filesystem:

```go
type Filesystem struct {
    fs absfs.SymlinkFileSystem  // Embedded filesystem
}

type File struct {
    f absfs.File  // Embedded file
    mu sync.Mutex // Optional: add concurrency control
}

// Delegation methods
func (fs *Filesystem) Open(name string) (billy.File, error) {
    file, err := fs.fs.Open(name)
    if err != nil {
        return nil, err
    }
    return &File{f: file}, nil  // Wrap absfs.File
}
```

### 2. Path Translation Pattern
When implementing basefs or chroot functionality:

```go
func (fs *Filesystem) translatePath(path string) (string, error) {
    // Ensure path is within base directory
    // Convert relative to absolute
    // Validate no directory traversal
    return absolutePath, nil
}
```

### 3. File Handle Caching
Some implementations cache file metadata:

```go
func (f *File) Stat() (os.FileInfo, error) {
    if f.cachedInfo != nil && !isCacheStale() {
        return f.cachedInfo, nil
    }
    info, err := f.realStat()
    f.cachedInfo = info
    return info, err
}
```

### 4. Error Handling Pattern
Proper error wrapping:

```go
func (fs *Filesystem) Stat(name string) (os.FileInfo, error) {
    realPath := fs.translatePath(name)
    info, err := fs.underlying.Stat(realPath)
    if err != nil {
        // os package already returns *os.PathError
        return nil, err
    }
    return info, nil
}
```

---

## Common Usage Patterns

### Creating Wrapped Filesystem
```go
// Start with an abstract filesystem
fs, err := osfs.NewFS()  // OS filesystem
// or
fs, err := memfs.NewFS()  // In-memory filesystem

// Wrap with basefs for chroot/isolation
wrappedFS, err := basefs.NewFS(fs, "/base/path")

// Convert to billy.Filesystem for go-git
billyFS, err := billyfs.NewFS(wrappedFS, "/")
```

### Implementing Custom Filesystem
```go
// 1. Start with Filer implementation
type MyFS struct { /* ... */ }

// Implement all Filer methods
func (fs *MyFS) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) { /* ... */ }
func (fs *MyFS) Mkdir(name string, perm os.FileMode) error { /* ... */ }
// ... etc for all Filer methods

// 2. Add FileSystem convenience methods
func (fs *MyFS) Separator() uint8 { return '/' }
func (fs *MyFS) ListSeparator() uint8 { return ':' }
// ... etc

// 3. For full support, add SymLinker methods
func (fs *MyFS) Lstat(name string) (os.FileInfo, error) { /* ... */ }
func (fs *MyFS) Symlink(oldname, newname string) error { /* ... */ }
// ... etc
```

---

## Key Design Decisions

1. **Interface Composition**: Large functionality built from small, focused interfaces
2. **Optional Implementation**: Implementations can support subsets (ReadOnlyFiler vs Filer)
3. **Standard os.FileInfo**: Leverages Go's standard library types
4. **Error Type Standards**: Uses os.PathError and os.LinkError for consistency
5. **Path Separators Explicit**: Filesystem can define its own separator (/ vs \\)
6. **Basestation Wrapping**: basefs provides path-based isolation without reimplementation
7. **No Globals**: All operations are method-based, no global state
8. **Context-Free**: Methods don't require context.Context

---

## Migration/Comparison Guide for Billy to AbsFS

| Billy v5 Interface | AbsFS Equivalent | Notes |
|---|---|---|
| billy.File | absfs.File | Very similar, minor differences |
| billy.Filesystem | absfs.SymlinkFileSystem | Full equivalence |
| billy.Capability | N/A | absfs assumes all capabilities |
| billy.Chroot() | basefs wrapper | Different approach - wraps FS |
| billy.Root() | basefs.Prefix() | Get current base directory |

---

## References
- **absfs Package**: github.com/absfs/absfs
- **basefs Package**: github.com/absfs/basefs
- **billyfs Reference**: github.com/absfs/billyfs (adapter implementation)
- **Standard Library**: os, io, filepath packages
