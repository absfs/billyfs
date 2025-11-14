# ABSFS Standards Documentation Index

This directory contains comprehensive documentation about ABSFS (Abstract File System) interface standards and how to implement compliant filesystems.

## Quick Start

**Start here:** [ABSFS Standards Summary](ABSFS_STANDARDS_SUMMARY.md)
- Executive overview
- Interface hierarchy
- Implementation levels
- Critical standards
- Key design principles

## Documentation Files

### 1. ABSFS_STANDARDS_SUMMARY.md
**Purpose:** Executive overview and getting started guide
**Contains:**
- What is ABSFS and why it matters
- Complete interface hierarchy with visual diagram
- Implementation levels (1-6) quick reference
- File operations breakdown by category
- Critical standards (error handling, modes, paths, symlinks, etc.)
- ABSFS ecosystem and role of basefs/billyfs
- Implementation checklist
- Standards violations to avoid
- Key design principles
- Billy vs AbsFS comparison

**Read this first** for a high-level understanding.

### 2. ABSFS_INTERFACE_REFERENCE.md
**Purpose:** Complete technical reference for all ABSFS interfaces
**Contains:**
- Detailed interface definitions with signatures
- UnSeekable, Seekable, File interface specifications
- Filer, FileSystem, SymLinker specifications
- File opening flags (O_RDONLY, O_CREATE, etc.)
- Permission flags and constants
- Error handling standards
- Helper utilities (basefs, extension functions)
- Type definitions and utilities (Flags, ParseFileMode, InvalidFile, FastWalkFunc)
- Implementation requirements checklist
- Implementation patterns (delegation, path translation, caching, error handling)
- Common usage patterns
- Key design decisions
- Migration/comparison guide

**Reference this** while implementing for precise method signatures and requirements.

### 3. BILLYFS_IMPLEMENTATION_GUIDE.md
**Purpose:** Real-world example showing ABSFS implementation
**Contains:**
- How billyfs works as a two-level adapter
- File wrapper implementation details
- Filesystem wrapper method mappings
- Key implementation patterns (constructor, file wrapping, directory listing, temp files)
- All absfs methods used by billyfs (with mapping table)
- All file interface methods used (with mapping table)
- What billyfs adds/modifies
- Validation checklist for new implementations
- ABSFS ecosystem visualization
- Common implementation gotchas

**Study this** to understand patterns used in real implementations.

### 4. ABSFS_COMPLIANCE_CHECKLIST.md
**Purpose:** Validation guide for implementing ABSFS-compliant filesystems
**Contains:**
- Interface implementation matrix (all methods with signatures)
- Implementation levels 1-6 with examples
- Pre-implementation decision checklist
- File type implementation verification checklist
- Filesystem type implementation verification checklist
- Error handling verification checklist
- Path handling verification checklist
- Metadata verification checklist
- Testing verification checklist
- Common pitfalls to avoid (8 categories with examples)
- Integration testing pattern
- Quick reference: method count by level
- Standards summary

**Use this** when validating your implementation.

## Implementation Path

### For Understanding ABSFS
1. Read ABSFS_STANDARDS_SUMMARY.md (overview)
2. Skim ABSFS_INTERFACE_REFERENCE.md (what exists)
3. Review BILLYFS_IMPLEMENTATION_GUIDE.md (how it's done)

### For Implementing a Filesystem
1. Read ABSFS_STANDARDS_SUMMARY.md to determine your level
2. Use ABSFS_INTERFACE_REFERENCE.md as technical reference
3. Study BILLYFS_IMPLEMENTATION_GUIDE.md for patterns
4. Follow ABSFS_COMPLIANCE_CHECKLIST.md for validation

### For Validating Compliance
1. Reference ABSFS_COMPLIANCE_CHECKLIST.md
2. Cross-check against ABSFS_INTERFACE_REFERENCE.md
3. Compare patterns with BILLYFS_IMPLEMENTATION_GUIDE.md

## Key Concepts Quick Reference

### Interface Hierarchy
```
UnSeekable (7) → Seekable (+1) → File (+4)
Filer (8) → FileSystem (+10) → SymlinkFileSystem (+4)
```

### Implementation Levels
- **Level 1:** ReadOnlyFiler (1 method) - Archives, read-only
- **Level 2:** WriteOnlyFiler (1 method) - Logging, write-only
- **Level 3:** Filer (8 methods) - Basic operations
- **Level 4:** FileSystem (18 methods) - Standard operations
- **Level 5:** SymlinkFileSystem (22 methods) - POSIX-like
- **Level 6:** SymlinkFileSystem + File (34+ methods) - Production

### Critical Standards
1. **Error Handling:** PathError (single-file), LinkError (two-path)
2. **File Mode:** O_RDONLY/WRONLY/RDWR (exclusive), O_APPEND/CREATE/EXCL/SYNC/TRUNC
3. **Permissions:** 0644 (file), 0755 (dir), POSIX bits (user/group/other)
4. **Symlinks:** Stat follows, Lstat doesn't; Chown target, Lchown link
5. **Directories:** Mkdir (single), MkdirAll (recursive), Remove (empty), RemoveAll (contents)
6. **Metadata:** FileInfo with Name/Size/Mode/ModTime/IsDir/Sys
7. **Paths:** Use Separator() dynamically, normalize consistently
8. **Thread-Safety:** Consider sync.Mutex for concurrent access

### BillyFS Patterns
- **Delegation:** Wrap underlying filesystem, delegate all operations
- **File Wrapping:** Return wrapped File objects from filesystem methods
- **Thread-Safety:** Use sync.Mutex at file level
- **Path-Based Isolation:** Use basefs.NewFS for chroot functionality
- **Error Handling:** Preserve error types from underlying filesystem

## Common Questions Answered

**Q: Which level should I implement?**
A: Determine your needs - Level 3 for basic operations, Level 5 for POSIX compatibility, Level 6 for production use.

**Q: Do I need to implement everything?**
A: No - implement the interface level you need. Smaller interfaces are valid implementations.

**Q: How do I validate my implementation?**
A: Use ABSFS_COMPLIANCE_CHECKLIST.md or wrap it with billyfs - if billyfs works, you're compliant.

**Q: What about thread safety?**
A: Optional by design - add sync.Mutex if concurrent access needed.

**Q: Can I mix interface levels?**
A: Yes - implement Filer + SymLinker for example, using basefs wrapper for convenience.

**Q: How does billyfs relate to absfs?**
A: BillyFS is an adapter - it wraps absfs.SymlinkFileSystem and presents it as billy.Filesystem for go-git compatibility.

## Document Statistics

- **ABSFS_STANDARDS_SUMMARY.md:** Executive summary (10-15 min read)
- **ABSFS_INTERFACE_REFERENCE.md:** Complete reference (30-45 min read)
- **BILLYFS_IMPLEMENTATION_GUIDE.md:** Implementation example (20-30 min read)
- **ABSFS_COMPLIANCE_CHECKLIST.md:** Validation guide (30-60 min reference)

**Total time to understand ABSFS:** 2-3 hours
**Time to implement basic filesystem:** 4-8 hours depending on complexity

## See Also

- **BillyFS Source:** billyfs.go, billyfile.go (actual implementation)
- **Project README:** Readme.md (project overview)
- **Go Modules:** go.mod (dependencies: absfs, basefs, osfs, go-billy)

## Standards Compliance

These documents describe the ABSFS ecosystem as of:
- **absfs:** v0.0.0-20230318165928-6f31c6ac7458
- **basefs:** v0.0.0-20240302110531-053a96f8cf86
- **billyfs:** github.com/absfs/billyfs (current)

Standards match Go's os package semantics and POSIX conventions.

---

For additional information about specific interfaces or patterns, refer to the appropriate document in the "Documentation Files" section above.
