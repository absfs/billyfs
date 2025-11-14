# ABSFS Quick Reference Card

## Essential Interfaces at a Glance

### File Interfaces (11 methods total)
```go
// Core (UnSeekable)
Name() string
Read([]byte) (int, error)
Write([]byte) (int, error)
Close() error
Sync() error
Stat() (os.FileInfo, error)
Readdir(int) ([]os.FileInfo, error)

// Positioning (Seekable)
Seek(int64, int) (int64, error)

// Advanced (File)
ReadAt([]byte, int64) (int, error)
WriteAt([]byte, int64) (int, error)
WriteString(string) (int, error)
Truncate(int64) error
```

### Filesystem Interfaces (22 methods total)

#### Filer (8 core methods)
```go
OpenFile(string, int, os.FileMode) (File, error)
Mkdir(string, os.FileMode) error
Remove(string) error
Rename(string, string) error
Stat(string) (os.FileInfo, error)
Chmod(string, os.FileMode) error
Chtimes(string, time.Time, time.Time) error
Chown(string, int, int) error
```

#### FileSystem additions (10 convenience methods)
```go
Separator() uint8
ListSeparator() uint8
Chdir(string) error
Getwd() (string, error)
TempDir() string
Open(string) (File, error)           // Convenience
Create(string) (File, error)         // Convenience
MkdirAll(string, os.FileMode) error
RemoveAll(string) error
Truncate(string, int64) error
```

#### SymLinker (4 methods)
```go
Lstat(string) (os.FileInfo, error)
Lchown(string, int, int) error
Readlink(string) (string, error)
Symlink(string, string) error
```

## Critical Constants

### File Open Flags
```go
O_RDONLY    // Read only (default)
O_WRONLY    // Write only
O_RDWR      // Read and write
O_APPEND    // Append mode
O_CREATE    // Create if not exist
O_EXCL      // Fail if exists (with CREATE)
O_SYNC      // Synchronous I/O
O_TRUNC     // Truncate
```

### File Permissions (Octal)
```go
0644        // rw-r--r-- (typical file)
0755        // rwxr-xr-x (typical directory)
0600        // rw------- (private)
0700        // rwx------ (private directory)
```

## Error Handling (IMPORTANT!)

```go
// Single-file operations → *os.PathError
Stat(), Chmod(), Chtimes(), Chown(), Lstat(), Lchown(), Readlink()

// Two-path operations → *os.LinkError
Rename(), Symlink()

// Example:
return &os.PathError{
    Op:   "stat",
    Path: name,
    Err:  os.ErrNotExist,
}
```

## Symlink Rules

| Operation | Follows Symlink? | Method |
|-----------|------------------|--------|
| Get file info | Yes | Stat() |
| Get link info | No | Lstat() |
| Change owner | Yes (target) | Chown() |
| Change link owner | No (link) | Lchown() |
| Read link target | N/A | Readlink() |
| Create link | N/A | Symlink() |

## Implementation Levels Quick Lookup

| Level | Name | Methods | Typical Use |
|-------|------|---------|------------|
| 1 | ReadOnlyFiler | 1 | Read-only virtual FS |
| 2 | WriteOnlyFiler | 1 | Write-only logging |
| 3 | Filer | 8 | Basic file ops |
| 4 | FileSystem | 18 | Standard operations |
| 5 | SymlinkFileSystem | 22 | POSIX-like, full-featured |
| 6 | + File interface | 34+ | Production systems |

## Implementation Checklist (Minimal)

For Level 5 (SymlinkFileSystem):

File type:
- [ ] Name, Read, Write, Close, Sync, Stat, Readdir (UnSeekable)
- [ ] Seek (Seekable)
- [ ] ReadAt, WriteAt, WriteString, Truncate (File)

Filesystem type:
- [ ] OpenFile, Mkdir, Remove, Rename, Stat, Chmod, Chtimes, Chown (Filer)
- [ ] Separator, ListSeparator, Chdir, Getwd, TempDir (FileSystem)
- [ ] Open, Create, MkdirAll, RemoveAll, Truncate (FileSystem)
- [ ] Lstat, Lchown, Readlink, Symlink (SymLinker)

Validation:
- [ ] Errors are proper types (PathError, LinkError)
- [ ] File modes respected (not ignored)
- [ ] Paths normalized and handled safely
- [ ] Thread-safety considered (add sync.Mutex if needed)
- [ ] FileInfo includes all required fields
- [ ] Symlink operations work correctly

## Common Patterns

### File Wrapping (Delegation)
```go
type File struct {
    f absfs.File
    mu sync.Mutex
}

func (f *File) Read(p []byte) (int, error) {
    f.mu.Lock()
    defer f.mu.Unlock()
    return f.f.Read(p)
}
```

### Filesystem Wrapping
```go
type Filesystem struct {
    fs absfs.SymlinkFileSystem
}

func (fs *Filesystem) Open(name string) (File, error) {
    file, err := fs.fs.Open(name)
    if err != nil {
        return nil, err
    }
    return &File{f: file}, nil
}
```

### Using basefs for Chroot
```go
fs, _ := osfs.NewFS()
wrapped, _ := basefs.NewFS(fs, "/base/dir")  // Operate in /base/dir
// All paths now relative to /base/dir
```

### Testing with BillyFS
```go
// If it works with billyfs, it's ABSFS-compliant
bfs, err := billyfs.NewFS(yourFS, "/")
if err == nil {
    // You're compatible!
}
```

## Do's and Don'ts

### DO
- Use standard os.FileInfo and os.FileMode
- Return proper error types (*os.PathError, *os.LinkError)
- Respect file mode parameters
- Close files with defer
- Distinguish Stat/Lstat and Chown/Lchown
- Use Separator() dynamically
- Thread-protect concurrent access
- Normalize paths consistently

### DON'T
- Return generic fmt.Errorf() errors
- Ignore file mode parameters
- Follow symlinks in Lstat/Lchown
- Forget to close file handles
- Treat all paths as absolute
- Create TempDir on demand (must exist)
- Mix up directory operations (Remove vs RemoveAll)
- Leave unsynchronized concurrent access

## BillyFS Key Takeaway

BillyFS shows the pattern:
1. Take absfs.SymlinkFileSystem input
2. Wrap with basefs.NewFS for isolation
3. Wrap File/Filesystem objects for interface adaptation
4. Add thread-safety with sync.Mutex
5. Return properly wrapped objects

This adapter pattern allows any ABSFS-compliant filesystem to work with go-git!

## Validation Command (Pseudo-code)

```go
fs := myfs.NewFS()           // Your filesystem
bfs, _ := billyfs.NewFS(fs, "/")

// If you can do this, you're ABSFS-compliant:
bfs.Create("/test.txt")      // Works?
bfs.Open("/test.txt")        // Works?
bfs.Remove("/test.txt")      // Works?
bfs.Symlink("/a", "/b")      // Works?
bfs.Lstat("/b")              // Works?
```

## Files for Reference

- **ABSFS_STANDARDS_SUMMARY.md** - Full overview (start here)
- **ABSFS_INTERFACE_REFERENCE.md** - Complete specifications
- **BILLYFS_IMPLEMENTATION_GUIDE.md** - Real example
- **ABSFS_COMPLIANCE_CHECKLIST.md** - Validation guide
- **ABSFS_DOCUMENTATION_INDEX.md** - Navigation guide

## Resources

- Go Package: github.com/absfs/absfs
- BillyFS: github.com/absfs/billyfs
- Go-Billy: github.com/go-git/go-billy/v5

---

Print this page and keep it handy while implementing!
