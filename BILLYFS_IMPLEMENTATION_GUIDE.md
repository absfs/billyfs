# ABSFS Compliance - BillyFS Implementation Reference

This document shows how billyfs implements the absfs interface requirements as a real-world example.

## Project Structure

```
billyfs/
├── billyfs.go      - Filesystem adapter (implements billy.Filesystem)
├── billyfile.go    - File adapter (implements billy.File)
└── go.mod          - Dependencies
```

## How BillyFS Works

BillyFS is a **two-level adapter**:
1. Takes an `absfs.SymlinkFileSystem` as input
2. Wraps it with `basefs.NewFS()` to provide path isolation
3. Wraps the result in a `billy.Filesystem` interface

```
┌──────────────────────────────────────────────────┐
│  Billy.Filesystem (go-git compatible)            │
│  └─ Filesystem struct                           │
│     └─ fs: absfs.SymlinkFileSystem             │
│        └─ (wrapped by basefs for chroot)       │
│           └─ Original absfs implementation       │
│              (osfs, memfs, custom, etc.)       │
└──────────────────────────────────────────────────┘
```

## Core Implementation Details

### File Wrapper (billyfile.go)

```go
type File struct {
    f  absfs.File
    mu sync.Mutex  // For thread-safety
}

// Delegates all billy.File methods to absfs.File:
Name()       → f.f.Name()
Read()       → f.f.Read()
Write()      → f.f.Write()
Close()      → f.f.Close()
Seek()       → f.f.Seek()
ReadAt()     → f.f.ReadAt()
WriteAt()    → f.f.WriteAt()
Truncate()   → f.f.Truncate()
Lock()       → f.mu.Lock()   (mutex-based)
Unlock()     → f.mu.Unlock() (mutex-based)
```

### Filesystem Wrapper Methods (billyfs.go)

#### Basic File Operations
```go
Create(filename string) → f.fs.Create(filename)
Open(filename string)   → f.fs.Open(filename)
OpenFile(filename, flag, perm) → f.fs.OpenFile(filename, flag, perm)
```

#### File Metadata
```go
Stat(filename string) → f.fs.Stat(filename)
Lstat(filename string) → f.fs.Lstat(filename)  // No symlink follow
```

#### File Modification
```go
Rename(oldpath, newpath string) → f.fs.Rename()
Remove(filename string) → f.fs.Remove()
Chmod(name string, mode) → f.fs.Chmod()
Chown(name, uid, gid) → f.fs.Chown()
Lchown(name, uid, gid) → f.fs.Lchown()  // Symlink itself
Chtimes(name, atime, mtime) → f.fs.Chtimes()
```

#### Directory Operations
```go
ReadDir(path string) → dir.Open() + dir.Readdir(0)
MkdirAll(filename, perm) → f.fs.MkdirAll()
```

#### Symlink Operations
```go
Symlink(target, link) → f.fs.Symlink()
Readlink(link) → f.fs.Readlink()
```

#### Chroot/Isolation
```go
Chroot(path string) → basefs.NewFS(f.fs, path)  // Creates sub-filesystem
Root() string → basefs.Prefix(f.fs)  // Gets current base path
```

#### Capabilities
```go
Capabilities() → billy.AllCapabilities  // Declares full support
```

#### Utilities
```go
Join(elem ...string) → filepath.Join()
TempFile(dir, prefix) → f.fs.Create() with random name
```

## Key Implementation Patterns

### 1. Constructor Pattern
```go
func NewFS(fs absfs.SymlinkFileSystem, dir string) (*Filesystem, error) {
    // Wrap with basefs for path isolation
    fs, err := basefs.NewFS(fs, dir)
    if err != nil {
        return nil, err
    }
    return &Filesystem{fs: fs}, nil
}
```

### 2. File Wrapping Pattern
When returning files to billy consumers:
```go
func (f *Filesystem) Open(filename string) (billy.File, error) {
    file, err := f.fs.Open(filename)
    if err != nil {
        return nil, err
    }
    // Wrap absfs.File in billy.File adapter
    return &File{f: file}, nil
}
```

### 3. Directory Listing Pattern
```go
func (f *Filesystem) ReadDir(path string) ([]os.FileInfo, error) {
    // Open as directory file
    dir, err := f.fs.Open(path)
    if err != nil {
        return nil, err
    }
    defer dir.Close()
    
    // Read all entries (0 = read all)
    return dir.Readdir(0)
}
```

### 4. Temporary File Pattern
```go
func (f *Filesystem) TempFile(dir string, prefix string) (billy.File, error) {
    // Generate unique name
    p := path.Join(f.fs.TempDir(), prefix+"_"+randSeq(5))
    
    // Create file
    file, err := f.fs.Create(p)
    if err != nil {
        return nil, err
    }
    return &File{f: file}, nil
}
```

## AbsFS Methods Used by BillyFS

| AbsFS Method | Used In | Purpose |
|---|---|---|
| Create() | Create() | Create new file |
| Open() | Open(), ReadDir() | Open for reading |
| OpenFile() | OpenFile() | Flexible open with flags |
| Stat() | Stat() | Get file info |
| Lstat() | Lstat() | Get symlink info |
| Chmod() | Chmod() | Change permissions |
| Chown() | Chown() | Change owner/group |
| Lchown() | Lchown() | Change symlink owner |
| Chtimes() | Chtimes() | Change times |
| Rename() | Rename() | Move/rename |
| Remove() | Remove() | Delete file/empty dir |
| MkdirAll() | MkdirAll() | Create dir hierarchy |
| Symlink() | Symlink() | Create symlink |
| Readlink() | Readlink() | Read symlink target |
| TempDir() | TempFile() | Get temp directory |

## File Interface Methods Used by BillyFS

| File Method | Used In | Purpose |
|---|---|---|
| Name() | Name() | Get file path |
| Read() | File.Read() | Read data |
| Write() | File.Write() | Write data |
| ReadAt() | File.ReadAt() | Read at offset |
| WriteAt() | File.WriteAt() | Write at offset |
| Seek() | File.Seek() | Move file pointer |
| Close() | File.Close() | Close file |
| Truncate() | File.Truncate() | Resize file |
| Readdir() | ReadDir() | List directory |

## What BillyFS Adds/Modifies

### Additions
1. **Thread-safety**: Wraps file operations with sync.Mutex
2. **Lock/Unlock**: Adds explicit file locking interface
3. **Billy compatibility**: Implements billy.File and billy.Filesystem interfaces
4. **All capabilities**: Declares support for all billy capabilities

### Differences from Raw AbsFS
- Path operations use `filepath.Join()` (GOROOT standard)
- TempFile creates files with random suffixes
- Chroot returns wrapped filesystem (not underlying)
- Root returns basefs prefix path

## Validation Checklist for New AbsFS Implementations

When implementing an absfs filesystem, billyfs shows you need:

### Required Methods (All 30 ✓)
- **UnSeekable** (7): Name, Read, Write, Close, Sync, Stat, Readdir
- **Seekable** (1): Seek
- **File** (4): ReadAt, WriteAt, WriteString, Truncate (+ Readdirnames if used)
- **Filer** (8): OpenFile, Mkdir, Remove, Rename, Stat, Chmod, Chtimes, Chown
- **FileSystem** (10): Separator, ListSeparator, Chdir, Getwd, TempDir, Open, Create, MkdirAll, RemoveAll, Truncate
- **SymLinker** (4): Lstat, Lchown, Readlink, Symlink

### Testing Pattern (from billyfs_test.go)
```go
func TestBillyfs(t *testing.T) {
    // 1. Create underlying absfs filesystem
    fs, err := osfs.NewFS()
    if err != nil {
        t.Fatal(err)
    }
    
    // 2. Wrap with billyfs
    bfs, err := billyfs.NewFS(fs, "/")
    if err != nil {
        t.Fatal(err)
    }
    
    // 3. Verify it's usable as billy.Filesystem
    _ = bfs  // Type check
}
```

## AbsFS Ecosystem

BillyFS demonstrates the adapter pattern in the absfs ecosystem:

```
Multiple AbsFS Implementations:
├── osfs - OS filesystem
├── memfs - In-memory filesystem
├── basefs - Path-based wrapper/chroot
├── fstesting - Testing utilities
└── Your custom implementation

↓

All work with billyfs to become billy.Filesystem
↓

All work with go-git and git operations
```

## Common Implementation Gotchas (Learned from BillyFS)

1. **Path Handling**: Use filepath operations, not string manipulation
2. **Error Wrapping**: Ensure proper error types (PathError, LinkError)
3. **Directory Readdir**: Returns os.FileInfo, not just names
4. **Symlink Handling**: Distinguish Stat vs Lstat, Chown vs Lchown
5. **TempDir**: Must return valid directory path, not create on demand
6. **Chroot/Basefs**: Use basefs wrapper rather than reimplementing
7. **File Closing**: Always defer Close() in file operations
8. **Thread Safety**: Consider mutex protection for concurrent access

