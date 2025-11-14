# ABSFS Filesystem Standards - Executive Summary

## What is ABSFS?

ABSFS (Abstract File System) is a Go package that defines a comprehensive set of interfaces for implementing custom filesystems. It provides:

1. **Interface hierarchy** - From minimal (ReadOnlyFiler) to complete (SymlinkFileSystem)
2. **Standard semantics** - Matches Go's os package behavior
3. **Composable design** - Build from small, focused interfaces
4. **Multiple implementation levels** - Choose what features you need

## Key Interfaces (Interface Hierarchy)

```
┌─ UnSeekable (7 methods)
   ├─ Name(), Read(), Write(), Close(), Sync(), Stat(), Readdir()
│
├─ Seekable (+ Seek method)
│  └─ Used by File interface
│
├─ File (+ ReadAt, WriteAt, WriteString, Truncate, Readdirnames)
│  └─ Complete file operations
│
├─ Filer (8 methods)
│  ├─ OpenFile(), Mkdir(), Remove(), Rename()
│  └─ Stat(), Chmod(), Chtimes(), Chown()
│
├─ FileSystem (extends Filer, +10 methods)
│  ├─ Separator(), ListSeparator(), Chdir(), Getwd(), TempDir()
│  ├─ Open(), Create(), MkdirAll(), RemoveAll(), Truncate()
│  └─ (Convenience methods often have default implementations)
│
├─ SymLinker (4 methods)
│  ├─ Lstat(), Lchown(), Readlink(), Symlink()
│  └─ Symbolic link support
│
└─ SymlinkFileSystem = FileSystem + SymLinker
   └─ Complete POSIX-like filesystem (22 methods total)
```

## Implementation Levels Quick Reference

| Level | Interface | Methods | Use Case |
|-------|-----------|---------|----------|
| 1 | ReadOnlyFiler | 1 | Read-only archives, virtual filesystems |
| 2 | WriteOnlyFiler | 1 | Write-only logging, event streams |
| 3 | Filer | 8 | Basic file operations only |
| 4 | FileSystem | 18 | Standard filesystem with conveniences |
| 5 | SymlinkFileSystem | 22 | Full POSIX-like filesystem |
| 6 | SymlinkFileSystem + File | 34+ | Production-grade with all features |

## File Operations Breakdown

### Core File Methods (7 - UnSeekable)
```go
Name()          // Get file path
Read()          // Sequential read
Write()         // Sequential write
Close()         // Release handle
Sync()          // Flush to storage
Stat()          // Get metadata
Readdir()       // List directory
```

### Positioning (1 - Seekable)
```go
Seek()          // Move file pointer
```

### Advanced File (4 - File interface additions)
```go
ReadAt()        // Random read
WriteAt()       // Random write
WriteString()   // String write
Truncate()      // Resize
```

### Filesystem Core (8 - Filer)
```go
OpenFile()      // Open with flags
Mkdir()         // Create directory
Remove()        // Delete file/empty dir
Rename()        // Move/rename
Stat()          // Get file info
Chmod()         // Change permissions
Chtimes()       // Change times
Chown()         // Change owner
```

### Filesystem Convenience (10 - FileSystem additions)
```go
Separator()     // Path separator ('/' or '\\')
ListSeparator() // Path list separator (':' or ';')
Chdir()         // Change working directory
Getwd()         // Get current directory
TempDir()       // Get temp directory location
Open()          // Open read-only (convenience)
Create()        // Create/truncate (convenience)
MkdirAll()      // Recursive mkdir
RemoveAll()     // Recursive remove
Truncate()      // Truncate by path
```

### Symlink Operations (4 - SymLinker)
```go
Lstat()         // Stat without following symlinks
Lchown()        // Chown the symlink itself
Readlink()      // Get symlink target
Symlink()       // Create symlink
```

## Critical Standards

### 1. Error Handling
- Use `*os.PathError` for single-file operations (Stat, Chmod, etc.)
- Use `*os.LinkError` for two-path operations (Rename, Symlink)
- Return `io.EOF` when reading directories/files at end
- Preserve error context with proper wrapping

### 2. File Mode Semantics
- `OpenFile()` flags: O_RDONLY, O_WRONLY, O_RDWR (mutually exclusive)
- Mode flags: O_APPEND, O_CREATE, O_EXCL, O_SYNC, O_TRUNC
- Permissions: 0644 (typical file), 0755 (typical dir), 0600 (private)
- Permissions use POSIX bits: owner/group/other × read/write/execute

### 3. Path Handling
- Use Separator() to determine path separator dynamically
- Normalize paths consistently
- Handle relative vs absolute paths correctly
- Preserve symlink targets as-is (don't normalize)

### 4. Metadata Requirements
- All FileInfo must implement: Name(), Size(), Mode(), ModTime(), IsDir(), Sys()
- Mode() must return proper os.FileMode (includes type bits)
- Directory entries (Readdir) must return sorted FileInfo

### 5. Symlink Rules
- `Stat()` follows symlinks; `Lstat()` doesn't
- `Chown()` changes target owner; `Lchown()` changes link owner
- `Readlink()` returns target as string (no normalization)
- Symlink targets can be relative or absolute

### 6. Directory Operations
- `Mkdir()` creates single directory
- `MkdirAll()` creates parents as needed
- `Remove()` removes files and empty directories
- `RemoveAll()` removes recursively
- Readdir() returns entries in directory order

### 7. Temporal Operations
- `Chtimes()` changes access and modification times
- Times can be less precise on some filesystems
- Zero time should be handled gracefully

### 8. Thread Safety (Implementation Choice)
- Consider using sync.Mutex for concurrent access
- File operations should be safe for concurrent reads
- Write operations should be serialized or protected

## ABSFS in Ecosystem

```
Your Custom Filesystem
├─ Implements absfs.SymlinkFileSystem
│
├─ Can be wrapped with basefs
│  └─ Provides path isolation/chroot
│
└─ Can be wrapped with billyfs
   └─ Becomes go-billy/v5 compatible
      └─ Works with go-git package
```

## BillyFS as Reference Implementation

BillyFS shows you:
1. **How to use absfs** - Wraps absfs.SymlinkFileSystem
2. **How to adapt interfaces** - Maps absfs to billy.File/Filesystem
3. **Common patterns** - File wrapping, delegation, error handling
4. **Thread safety** - Uses sync.Mutex for file-level locking
5. **Path-based isolation** - Uses basefs.NewFS for chroot

Key pattern: Delegation
```go
type File struct {
    f absfs.File
}

type Filesystem struct {
    fs absfs.SymlinkFileSystem
}

// Methods delegate to underlying implementation
```

## Implementation Checklist

### Before You Start
- [ ] Determine required interface level (which features needed?)
- [ ] Identify your underlying storage (OS, memory, custom, etc.)
- [ ] Plan error handling strategy
- [ ] Consider thread-safety requirements
- [ ] Identify metadata support (owner, times, modes, etc.)

### File Type
- [ ] Implement UnSeekable (7 methods)
- [ ] Add Seekable (Seek method)
- [ ] Add File methods (ReadAt, WriteAt, WriteString, Truncate)
- [ ] Optional: Readdirnames()

### Filesystem Type
- [ ] Implement Filer (8 methods)
- [ ] Add FileSystem methods (path, directory, convenience)
- [ ] Optional: SymLinker for symlink support
- [ ] Error handling (PathError, LinkError)

### Testing
- [ ] File creation, writing, reading
- [ ] Directory creation and listing
- [ ] Error cases (non-existent files, etc.)
- [ ] Symlink operations (if supported)
- [ ] Concurrent access (if needed)

## Standards Violations to Avoid

```go
// BAD: Generic errors instead of proper types
return fmt.Errorf("not found")          // Should return *os.PathError

// BAD: Ignoring file mode parameters
OpenFile(name, flags, 0)                // Should use provided perm

// BAD: Unsafe concurrent access
func (f *File) Read(p []byte) (int, error) {
    return f.file.Read(p)               // Unsafe!
}

// BAD: Following symlinks when shouldn't
Stat(symlink)                           // Use Lstat instead
Chown(symlink, uid, gid)                // Use Lchown instead

// BAD: Not closing files
dir := fs.Open(path)
dir.Readdir(0)                          // Should defer Close()

// BAD: Removing directories with files
fs.Remove("/dir")                       // Use RemoveAll for contents

// BAD: TempDir returning non-existent path
return "/tmp-that-doesnt-exist"         // Must return existing directory
```

## Key Design Principles

1. **Composition > Inheritance** - Build from small interfaces
2. **Standard Library Compatibility** - Match os package behavior
3. **Optional Features** - Implementations choose what to support
4. **Explicit Semantics** - Separate Stat/Lstat, Chown/Lchown
5. **Proper Error Types** - PathError vs LinkError
6. **No Global State** - Everything is method-based
7. **Clear Responsibilities** - Each interface has defined scope
8. **Platform Awareness** - Separator/ListSeparator allow OS customization

## Comparison: Billy vs AbsFS

| Aspect | Billy v5 | AbsFS | Notes |
|--------|----------|-------|-------|
| Interfaces | 3 main | 7 composable | AbsFS more modular |
| File interface | Single | Multiple levels | AbsFS supports subsets |
| Capabilities | Declared | All assumed | AbsFS assumes full support |
| Chroot | Method-based | Wrapper-based | AbsFS uses basefs wrapper |
| Symlinks | Full support | Full support | Both support symlinks |
| Error types | Limited | Standard os types | AbsFS more precise |

## Documentation Files Provided

1. **absfs_standards.md** - Complete interface reference
   - All interface definitions
   - Complete method signatures
   - Error handling standards
   - Implementation patterns
   - Common usage patterns

2. **billyfs_implementation_reference.md** - Real-world example
   - How billyfs implements absfs
   - Pattern demonstrations
   - Method mappings
   - Common gotchas

3. **absfs_compliance_checklist.md** - Validation guide
   - Implementation matrix
   - Level-by-level requirements
   - Verification checklists
   - Common pitfalls
   - Testing patterns

## Next Steps

1. **Choose your level** - Determine which interface level you need
2. **Review implementations** - Study osfs, memfs, or billyfs
3. **Plan your structure** - Design File and Filesystem types
4. **Implement incrementally** - Start with Filer, add FileSystem, then SymLinker
5. **Test thoroughly** - Use patterns from absfs ecosystem
6. **Validate with billyfs** - Wrap with billyfs to test compliance

## Resources

- **absfs**: github.com/absfs/absfs (v0.0.0-20230318165928-6f31c6ac7458)
- **basefs**: github.com/absfs/basefs (v0.0.0-20240302110531-053a96f8cf86)
- **billyfs**: github.com/absfs/billyfs (reference adapter)
- **osfs**: github.com/absfs/osfs (OS filesystem implementation)
- **Billy**: github.com/go-git/go-billy/v5 (git filesystem)

---

## Summary

ABSFS provides a well-designed, modular interface system for implementing filesystems in Go. By following the standards documented here, you can:

- Build compliant filesystem implementations
- Ensure compatibility with go-git and other tools
- Support multiple implementation levels
- Maintain consistent error handling
- Enable easy testing and validation

The key to successful ABSFS implementation is understanding the interface hierarchy, implementing appropriate levels for your needs, and following the error handling and semantics standards established by the os package.
