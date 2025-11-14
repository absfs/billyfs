# ABSFS Filesystem Implementation Compliance Checklist

Quick reference for verifying that a filesystem implementation meets absfs standards.

---

## Interface Implementation Matrix

### UnSeekable (Base File Operations) - 7 methods

| Method | Signature | Purpose | Must Have |
|--------|-----------|---------|-----------|
| Name | `() string` | Return file path | YES |
| Read | `(b []byte) (int, error)` | Read bytes sequentially | YES |
| Write | `(b []byte) (int, error)` | Write bytes sequentially | YES |
| Close | `() error` | Release file handle | YES |
| Sync | `() error` | Flush to storage | YES |
| Stat | `() os.FileInfo, error` | Get file metadata | YES |
| Readdir | `(int) ([]os.FileInfo, error)` | List directory entries | YES |

### Seekable (File Positioning) - 1 method

| Method | Signature | Purpose | Must Have |
|--------|-----------|---------|-----------|
| Seek | `(offset int64, whence int) (int64, error)` | Move file position | YES |

### File Interface - 5 additional methods

| Method | Signature | Purpose | Must Have |
|--------|-----------|---------|-----------|
| ReadAt | `(b []byte, off int64) (n int, err error)` | Read at offset | YES |
| WriteAt | `(b []byte, off int64) (n int, err error)` | Write at offset | YES |
| WriteString | `(s string) (n int, err error)` | Write string | YES |
| Truncate | `(size int64) error` | Resize file | YES |
| Readdirnames | `(n int) ([]string, error)` | Get directory names | NO (optional) |

### Filer (Core Filesystem) - 8 methods

| Method | Signature | Purpose | Must Have |
|--------|-----------|---------|-----------|
| OpenFile | `(name string, flag int, perm os.FileMode) (File, error)` | Core open | YES |
| Mkdir | `(name string, perm os.FileMode) error` | Create directory | YES |
| Remove | `(name string) error` | Delete file/empty dir | YES |
| Rename | `(oldpath, newpath string) error` | Move/rename | YES |
| Stat | `(name string) (os.FileInfo, error)` | Get file info | YES |
| Chmod | `(name string, mode os.FileMode) error` | Change mode | YES |
| Chtimes | `(name string, atime, mtime time.Time) error` | Change times | YES |
| Chown | `(name string, uid, gid int) error` | Change owner | YES |

### FileSystem (Convenience Methods) - 10 methods

| Method | Signature | Purpose | Must Have |
|--------|-----------|---------|-----------|
| Separator | `() uint8` | Path separator | YES |
| ListSeparator | `() uint8` | Path list separator | YES |
| Chdir | `(dir string) error` | Change directory | YES |
| Getwd | `() (string, error)` | Get current directory | YES |
| TempDir | `() string` | Get temp directory | YES |
| Open | `(name string) (File, error)` | Open read-only | YES (convenience) |
| Create | `(name string) (File, error)` | Create file | YES (convenience) |
| MkdirAll | `(name string, perm os.FileMode) error` | Create recursively | YES |
| RemoveAll | `(path string) error` | Delete recursively | YES |
| Truncate | `(name string, size int64) error` | Resize by path | YES |

### SymLinker (Symbolic Links) - 4 methods

| Method | Signature | Purpose | Must Have |
|--------|-----------|---------|-----------|
| Lstat | `(name string) (os.FileInfo, error)` | Stat link itself | YES |
| Lchown | `(name string, uid, gid int) error` | Chown link | YES |
| Readlink | `(name string) (string, error)` | Get link target | YES |
| Symlink | `(oldname, newname string) error` | Create link | YES |

### SymlinkFileSystem (Complete) = FileSystem + SymLinker
- Total: 22 methods required
- Total with file: 29+ methods

---

## Implementation Levels

### Level 1: ReadOnlyFiler (Minimal)
```
ReadOnlyFiler
└─ Open() - (name string) (io.ReadCloser, error)
```
Use when: Read-only virtual filesystems, archives

### Level 2: WriteOnlyFiler (Minimal)
```
WriteOnlyFiler
└─ Open() - (name string) (io.WriteCloser, error)
```
Use when: Write-only logging filesystems

### Level 3: Filer (Basic)
```
Filer (8 methods)
├─ OpenFile()
├─ Mkdir()
├─ Remove()
├─ Rename()
├─ Stat()
├─ Chmod()
├─ Chtimes()
└─ Chown()
```
Use when: Simple file operations only

### Level 4: FileSystem (Full Convenience)
```
FileSystem extends Filer (18 methods total)
├─ All Filer methods
├─ Separator()
├─ ListSeparator()
├─ Chdir()
├─ Getwd()
├─ TempDir()
├─ Open()         [convenience]
├─ Create()       [convenience]
├─ MkdirAll()     [convenience]
├─ RemoveAll()    [convenience]
└─ Truncate()     [convenience]
```
Use when: Standard filesystem operations

### Level 5: SymlinkFileSystem (Complete)
```
SymlinkFileSystem = FileSystem + SymLinker (22 methods total)
├─ All FileSystem methods
├─ Lstat()
├─ Lchown()
├─ Readlink()
└─ Symlink()
```
Use when: Full POSIX filesystem support

### Level 6: Full File Interface (Production)
```
File interface (29+ methods total)
├─ UnSeekable (7)
├─ Seekable (1 additional)
├─ File additions (4 additional)
└─ SymlinkFileSystem filesystem operations
```
Use when: Complete filesystem abstraction needed

---

## Pre-Implementation Checklist

Before implementing a filesystem, determine:

- [ ] What level of compatibility is needed?
- [ ] Do you need symlink support? (YES → SymlinkFileSystem)
- [ ] Do you need directory enumeration? (YES → Readdir/Readdirnames)
- [ ] Do you need file seeking? (YES → Full File interface)
- [ ] Do you need random access? (YES → ReadAt/WriteAt)
- [ ] Do you need current directory tracking? (YES → Chdir/Getwd)
- [ ] Do you need ownership tracking? (YES → Chown/Lchown)
- [ ] Do you need permission handling? (YES → Chmod)
- [ ] Do you need time tracking? (YES → Chtimes)
- [ ] Do you need temp directory? (YES → TempDir)

---

## Implementation Verification Checklist

### File Type Implementation

```go
type File struct {
    // Your underlying file representation
}

// Check implementation:
- [ ] Name() implemented
- [ ] Read() implemented
- [ ] Write() implemented
- [ ] Close() implemented
- [ ] Sync() implemented (may be no-op)
- [ ] Stat() implemented
- [ ] Readdir() implemented
- [ ] Seek() implemented
- [ ] ReadAt() implemented
- [ ] WriteAt() implemented
- [ ] WriteString() implemented
- [ ] Truncate() implemented
- [ ] Readdirnames() implemented (optional)
```

### Filesystem Type Implementation

```go
type FileSystem struct {
    // Your underlying filesystem representation
}

// Check Filer implementation:
- [ ] OpenFile() implemented
- [ ] Mkdir() implemented
- [ ] Remove() implemented
- [ ] Rename() implemented
- [ ] Stat() implemented
- [ ] Chmod() implemented
- [ ] Chtimes() implemented
- [ ] Chown() implemented

// Check FileSystem implementation:
- [ ] Separator() returns uint8 ('/' or '\\')
- [ ] ListSeparator() returns uint8 (':' or ';')
- [ ] Chdir() implemented
- [ ] Getwd() implemented
- [ ] TempDir() returns string
- [ ] Open() implemented (reads file)
- [ ] Create() implemented (creates/truncates file)
- [ ] MkdirAll() implemented (recursive)
- [ ] RemoveAll() implemented (recursive)
- [ ] Truncate() implemented (resize by path)

// Check SymLinker implementation (if needed):
- [ ] Lstat() implemented
- [ ] Lchown() implemented
- [ ] Readlink() implemented
- [ ] Symlink() implemented
```

### Error Handling Verification

```go
// Check each method's error handling:
- [ ] PathError for single-file operations
  - Stat(), Chmod(), Chtimes(), Chown(), Lstat(), Lchown(), Readlink()
  
- [ ] LinkError for two-path operations
  - Rename(), Symlink()
  
- [ ] Standard errors for directory operations
  - Mkdir(), Remove(), MkdirAll(), RemoveAll()
  
- [ ] io.EOF for Readdir/Readdirnames at end
  
- [ ] Proper error wrapping (don't lose context)
```

### Path Handling Verification

```go
// Verify path operations:
- [ ] All paths are properly normalized
- [ ] Separator is applied consistently
- [ ] Absolute vs relative paths handled correctly
- [ ] symlink targets preserved as-is (no normalization)
- [ ] No directory traversal vulnerabilities (if sandboxed)
- [ ] Hidden files handled correctly
- [ ] Special paths (., ..) handled correctly
- [ ] Case sensitivity appropriate for platform
```

### Metadata Verification

```go
// FileInfo implementation:
- [ ] Name() returns basename
- [ ] Size() returns file size
- [ ] Mode() returns proper os.FileMode
- [ ] ModTime() returns modification time
- [ ] IsDir() returns true for directories
- [ ] Sys() returns platform-specific info (or nil)

// Time handling:
- [ ] Access time preserved if supported
- [ ] Modification time preserved if supported
- [ ] Times are system-appropriate precision
- [ ] Zero time handled properly
```

### Testing Verification

```go
// Create test cases for:
- [ ] Opening non-existent file (should error)
- [ ] Creating file (should succeed)
- [ ] Writing to file (should succeed)
- [ ] Reading back data (should match)
- [ ] Directory creation (should succeed)
- [ ] File removal (should succeed)
- [ ] Directory removal empty (should succeed)
- [ ] Directory removal non-empty (should error or RemoveAll)
- [ ] Renaming file (should succeed)
- [ ] Renaming to existing file (behavior defined?)
- [ ] Stat operations (should return valid FileInfo)
- [ ] Chmod operations (if supported)
- [ ] Ownership operations (if supported)
- [ ] Symlink operations (if supported)
- [ ] Concurrent access (thread-safety)
- [ ] Large file handling
- [ ] Special filenames (spaces, unicode, special chars)
```

---

## Common Pitfalls to Avoid

### 1. Error Types
```go
// BAD: Generic error
return fmt.Errorf("file not found")

// GOOD: Proper error type
return &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
```

### 2. File Mode Handling
```go
// BAD: Ignoring mode in Create
file, _ := fs.OpenFile(name, os.O_CREATE, 0)

// GOOD: Use provided mode
file, _ := fs.OpenFile(name, os.O_CREATE, perm)
```

### 3. Directory vs File Operations
```go
// BAD: Treating as single operation
_ = fs.Remove("/path")  // Works for both files and dirs

// GOOD: Distinguish operations
_ = fs.Remove(path)     // Files and empty dirs
_ = fs.RemoveAll(path)  // Files, dirs, contents
```

### 4. Symlink Following
```go
// BAD: Following symlinks when shouldn't
_ = fs.Stat(symlink)   // Follows link

// GOOD: Distinguishing operations
_ = fs.Stat(path)      // Follows symlinks
_ = fs.Lstat(path)     // Doesn't follow symlinks
```

### 5. Concurrent Access
```go
// BAD: No synchronization
func (f *File) Read(p []byte) (int, error) {
    return f.file.Read(p)  // Unsafe!
}

// GOOD: Protected access
func (f *File) Read(p []byte) (int, error) {
    f.mu.Lock()
    defer f.mu.Unlock()
    return f.file.Read(p)
}
```

### 6. Temporary Directory
```go
// BAD: Temp directory doesn't exist
func (fs *FS) TempDir() string {
    return "/tmp"  // Must exist!
}

// GOOD: Return actual temp dir
func (fs *FS) TempDir() string {
    return "/tmp"  // Pre-verified to exist
}
```

### 7. File Closing
```go
// BAD: Not closing files
dir, _ := fs.Open(path)
entries := dir.Readdir(0)

// GOOD: Always close
dir, _ := fs.Open(path)
defer dir.Close()
entries := dir.Readdir(0)
```

### 8. Seek Behavior
```go
// BAD: Seek resets on Read
file, _ := fs.Open("test.txt")
file.Seek(5, io.SeekStart)
file.Read(p)  // Might not start at 5!

// GOOD: Seek position is maintained
file, _ := fs.Open("test.txt")
file.Seek(5, io.SeekStart)
file.Read(p)  // Starts at 5
```

---

## Integration Testing Pattern

```go
import (
    "github.com/absfs/billyfs"
    "github.com/go-git/go-billy/v5"
)

func TestMyFilesystem(t *testing.T) {
    // Create your filesystem
    fs := myfs.New()
    
    // Wrap with billyfs to test absfs compliance
    bfs, err := billyfs.NewFS(fs, "/")
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify billy interface works (absfs compliance test)
    if err := testBillyOperations(bfs); err != nil {
        t.Fatal(err)
    }
}

func testBillyOperations(fs billy.Filesystem) error {
    // If it works with billy, it's absfs-compliant
    if _, err := fs.Create("/test.txt"); err != nil {
        return err
    }
    return fs.Remove("/test.txt")
}
```

---

## Quick Reference: Method Count by Level

| Level | Interface | Methods |
|-------|-----------|---------|
| 1 | ReadOnlyFiler | 1 |
| 2 | WriteOnlyFiler | 1 |
| 3 | Filer | 8 |
| 3+ | Filer + UnSeekable | 15 |
| 4 | FileSystem | 18 |
| 4+ | FileSystem + File | 30 |
| 5 | SymlinkFileSystem | 22 |
| 6 | SymlinkFileSystem + File | 34+ |

---

## Standards Summary

1. **Use standard Go types**: os.FileMode, os.FileInfo, time.Time
2. **Use standard error types**: os.PathError, os.LinkError
3. **Use standard constants**: os.O_RDONLY, os.O_CREATE, etc.
4. **Use interface composition**: Build larger interfaces from smaller ones
5. **Be consistent**: Behavior should match os package semantics
6. **Handle edge cases**: Empty paths, non-existent files, invalid modes
7. **Thread-safe**: Use locks if concurrent access expected
8. **Proper cleanup**: Files must be closable, resources must be released

