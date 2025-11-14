# GitHub Issues for BillyFS Quality Improvements

Generated from comprehensive quality review - organized by priority and scope.

---

## 🔴 Critical Priority - Bugs (Fix Immediately)

### Issue #1: Enable and fix input validation in NewFS constructor

**Labels**: `bug`, `critical`, `security`

**Title**: Enable input validation in NewFS constructor

**Body**:
```markdown
## Problem
The `NewFS` constructor has essential validation code commented out (lines 25-38 in `billyfs.go`), allowing invalid inputs that can cause runtime errors or security issues.

## Current Code
```go
func NewFS(fs absfs.SymlinkFileSystem, dir string) (*Filesystem, error) {
    //if dir == "" {
    //    return nil, os.ErrInvalid
    //}
    // ... more commented validation
```

## Expected Behavior
- Should validate that `dir` is not empty
- Should validate that `dir` is an absolute path
- Should validate that `dir` exists in the filesystem
- Should validate that `dir` is actually a directory (not a file)

## Acceptance Criteria
- [ ] Uncomment validation code
- [ ] Ensure all validation checks work correctly
- [ ] Add tests for each validation failure case
- [ ] Document validation requirements in godoc

## Related
Part of quality review findings - see PROJECT_QUALITY_REVIEW.md Issue #1
```

---

### Issue #2: Fix TempFile to honor directory parameter

**Labels**: `bug`, `critical`

**Title**: TempFile ignores dir parameter, always uses TempDir()

**Body**:
```markdown
## Problem
The `TempFile` method completely ignores its `dir` parameter and always creates files in the system temp directory. This violates the billy.Filesystem interface contract.

## Current Code
```go
func (f *Filesystem) TempFile(dir string, prefix string) (billy.File, error) {
    rand.Seed(time.Now().UnixNano())
    p := path.Join(f.fs.TempDir(), prefix+"_"+randSeq(5))  // ❌ Ignores 'dir'
    // ...
}
```

## Expected Behavior
Per billy.Filesystem specification:
- If `dir` is non-empty, create temp file in that directory
- If `dir` is empty, use default temp directory
- File should be created with unique name

## Acceptance Criteria
- [ ] Honor `dir` parameter when provided
- [ ] Fall back to TempDir() only when `dir` is empty
- [ ] Add test cases for both scenarios
- [ ] Verify compatibility with go-git usage patterns

## Related
Part of quality review findings - see PROJECT_QUALITY_REVIEW.md Issue #2
```

---

### Issue #3: Replace deprecated rand.Seed with modern approach

**Labels**: `bug`, `medium`, `technical-debt`

**Title**: Replace deprecated rand.Seed in TempFile implementation

**Body**:
```markdown
## Problem
`rand.Seed()` is deprecated as of Go 1.20+ and is called on every TempFile invocation, which is inefficient and not thread-safe.

## Current Code
```go
func (f *Filesystem) TempFile(dir string, prefix string) (billy.File, error) {
    rand.Seed(time.Now().UnixNano())  // ❌ Deprecated, called repeatedly
    // ...
}
```

## Solution Options
1. Use Go 1.20+ automatic seeding (remove explicit seed call)
2. Use `rand.New(rand.NewSource(time.Now().UnixNano()))` for local generator
3. Use `crypto/rand` for stronger uniqueness guarantees

## Acceptance Criteria
- [ ] Remove or replace deprecated `rand.Seed()` call
- [ ] Ensure thread-safety
- [ ] Improve uniqueness (consider using more than 5 random chars)
- [ ] Add collision handling or document collision risk
- [ ] Verify no deprecation warnings on Go 1.22+

## Related
Part of quality review findings - see PROJECT_QUALITY_REVIEW.md Issue #3
```

---

## 🔴 Critical Priority - Testing

### Issue #4: Add comprehensive test suite for all filesystem operations

**Labels**: `testing`, `critical`, `good-first-issue`

**Title**: Add comprehensive test coverage for filesystem operations

**Body**:
```markdown
## Problem
Current test coverage is ~2% (only one smoke test). No functional testing exists for any filesystem operations.

## Current State
- 1 test function (TestBillyfs) - only validates initialization
- 0 tests for file operations
- 0 tests for error conditions
- 0 tests for concurrent access

## Required Test Coverage

### Basic Operations (Priority 1)
- [ ] TestFilesystem_Create
- [ ] TestFilesystem_Open
- [ ] TestFilesystem_OpenFile (with various flags)
- [ ] TestFilesystem_Stat
- [ ] TestFilesystem_Rename
- [ ] TestFilesystem_Remove

### File Operations (Priority 1)
- [ ] TestFile_Read
- [ ] TestFile_Write
- [ ] TestFile_Seek
- [ ] TestFile_ReadAt
- [ ] TestFile_WriteAt
- [ ] TestFile_Truncate

### Directory Operations (Priority 2)
- [ ] TestFilesystem_MkdirAll
- [ ] TestFilesystem_ReadDir

### Symlink Operations (Priority 2)
- [ ] TestFilesystem_Symlink
- [ ] TestFilesystem_Readlink
- [ ] TestFilesystem_Lstat

### Permissions (Priority 3)
- [ ] TestFilesystem_Chmod
- [ ] TestFilesystem_Chown
- [ ] TestFilesystem_Lchown
- [ ] TestFilesystem_Chtimes

### Chroot (Priority 2)
- [ ] TestFilesystem_Chroot
- [ ] TestFilesystem_Root
- [ ] TestFilesystem_Chroot_PathIsolation

### TempFile (Priority 2)
- [ ] TestFilesystem_TempFile
- [ ] TestFilesystem_TempFile_WithDir
- [ ] TestFilesystem_TempFile_Uniqueness

### Error Conditions (Priority 1)
- [ ] TestNewFS_InvalidPath
- [ ] TestNewFS_NonExistentPath
- [ ] TestNewFS_PathIsFile
- [ ] TestFilesystem_NonExistentFile
- [ ] TestFilesystem_PermissionDenied

### Integration (Priority 2)
- [ ] TestWithMemFS (test with in-memory filesystem)
- [ ] TestWithOSFS (test with OS filesystem)
- [ ] TestConcurrentAccess (test thread-safety)

## Target Coverage
Minimum 70% code coverage before merging.

## References
See PROJECT_QUALITY_REVIEW.md Appendix for detailed test plan.
```

---

### Issue #5: Add example tests for documentation

**Labels**: `testing`, `documentation`, `good-first-issue`

**Title**: Add runnable example tests

**Body**:
```markdown
## Problem
No example tests exist. Go's example tests serve dual purpose:
1. Verify code works as documented
2. Appear in godoc as usage examples

## Required Examples
- [ ] ExampleNewFS - basic initialization
- [ ] ExampleFilesystem_Create - creating and writing files
- [ ] ExampleFilesystem_Chroot - path isolation
- [ ] ExampleFilesystem_TempFile - temporary file creation
- [ ] ExampleFilesystem_ReadDir - listing directories
- [ ] ExampleFilesystem_Symlink - working with symlinks

## Reference
See absfs/osfs and absfs/memfs for example patterns.

## Acceptance Criteria
- [ ] Examples compile and run successfully
- [ ] Examples appear in godoc output
- [ ] Examples demonstrate common use cases
```

---

## 🟠 High Priority - Infrastructure

### Issue #6: Set up GitHub Actions CI/CD pipeline

**Labels**: `infrastructure`, `ci-cd`, `high-priority`

**Title**: Add GitHub Actions workflow for automated testing

**Body**:
```markdown
## Problem
No CI/CD automation exists. All testing and quality checks are manual.

## Required Workflow

Create `.github/workflows/test.yml` with:

### Test Matrix
- [ ] Test on Go versions: 1.20, 1.21, 1.22, 1.23
- [ ] Test on platforms: Linux, macOS, Windows
- [ ] Run all tests with race detector
- [ ] Generate code coverage report

### Quality Checks
- [ ] Run `go vet`
- [ ] Run `golangci-lint`
- [ ] Check code formatting (`gofmt`)

### Coverage Reporting
- [ ] Upload coverage to codecov.io or similar
- [ ] Enforce minimum coverage threshold (70%)
- [ ] Block PRs that decrease coverage

## Acceptance Criteria
- [ ] Workflow runs on every push and PR
- [ ] All checks must pass before merge
- [ ] Badge added to README showing build status
- [ ] Coverage badge added to README
```

---

### Issue #7: Add linting configuration and enforcement

**Labels**: `infrastructure`, `code-quality`, `medium-priority`

**Title**: Configure golangci-lint for code quality

**Body**:
```markdown
## Problem
No automated linting exists to enforce code quality standards.

## Tasks
- [ ] Create `.golangci.yml` configuration
- [ ] Enable recommended linters (gosec, errcheck, govet, staticcheck, etc.)
- [ ] Configure appropriate severity levels
- [ ] Integrate into GitHub Actions workflow
- [ ] Fix any existing linting issues
- [ ] Document linting requirements in CONTRIBUTING.md

## Recommended Linters
- gosec (security)
- errcheck (error handling)
- govet (suspicious constructs)
- staticcheck (bugs and performance)
- ineffassign (unused assignments)
- unconvert (unnecessary conversions)

## Acceptance Criteria
- [ ] Linting runs in CI
- [ ] All linting issues resolved
- [ ] Configuration committed to repository
```

---

## 🟡 Medium Priority - Code Quality

### Issue #8: Add error context wrapping throughout

**Labels**: `enhancement`, `error-handling`, `medium-priority`

**Title**: Wrap errors with context for better debugging

**Body**:
```markdown
## Problem
All errors from underlying filesystem are passed through without adding context. This makes debugging difficult in layered filesystem stacks.

## Current Pattern
```go
func (f *Filesystem) Create(filename string) (billy.File, error) {
    file, err := f.fs.Create(filename)
    if err != nil {
        return nil, err  // ❌ No context
    }
    return &File{f: file}, nil
}
```

## Improved Pattern
```go
func (f *Filesystem) Create(filename string) (billy.File, error) {
    file, err := f.fs.Create(filename)
    if err != nil {
        return nil, fmt.Errorf("billyfs.Create(%q): %w", filename, err)
    }
    return &File{f: file}, nil
}
```

## Acceptance Criteria
- [ ] All error returns include context
- [ ] Use `%w` verb to preserve error chains
- [ ] Include operation name and key parameters
- [ ] Verify error wrapping doesn't break error type checks
- [ ] Add tests that verify error messages
```

---

### Issue #9: Fix path separator handling in Join method

**Labels**: `bug`, `cross-platform`, `low-priority`

**Title**: Use filesystem separator instead of OS separator

**Body**:
```markdown
## Problem
`Join` method uses `filepath.Join` which is OS-dependent, not filesystem-dependent. This can cause issues with non-OS filesystems like memfs.

## Current Code
```go
func (f *Filesystem) Join(elem ...string) string {
    return filepath.Join(elem...)
}
```

## Solution
Use the underlying filesystem's separator:
```go
func (f *Filesystem) Join(elem ...string) string {
    sep := f.fs.Separator()
    // Custom join logic or use path.Join for unix-style
    return path.Join(elem...)  // For unix-style paths
}
```

## Acceptance Criteria
- [ ] Join respects underlying filesystem separator
- [ ] Test with both OS and non-OS filesystems
- [ ] Document behavior in godoc
```

---

### Issue #10: Remove unused fstesting dependency or add tests

**Labels**: `dependencies`, `testing`, `low-priority`

**Title**: Resolve unused fstesting dependency

**Body**:
```markdown
## Problem
The `github.com/absfs/fstesting` package is imported in go.mod but never used.

## Options
1. Use fstesting utilities to improve test suite (recommended)
2. Remove dependency if not needed

## If Using fstesting
- [ ] Review fstesting utilities available
- [ ] Integrate into test suite
- [ ] Document usage

## If Removing
- [ ] Remove from go.mod
- [ ] Run `go mod tidy`
- [ ] Verify tests still pass
```

---

## 🟡 Medium Priority - Documentation

### Issue #11: Complete README example and add troubleshooting

**Labels**: `documentation`, `medium-priority`

**Title**: Complete README.md and add missing sections

**Body**:
```markdown
## Problems
1. README example is truncated at line 104 ("Additional code to demonstrate...")
2. Missing sections: API reference, limitations, troubleshooting, comparison to alternatives

## Required Sections

### Complete Example
- [ ] Finish the truncated example code
- [ ] Add expected output
- [ ] Test that example actually runs

### Add Sections
- [ ] API Reference (brief method listing)
- [ ] Limitations (what billyfs doesn't support)
- [ ] Troubleshooting (common issues)
- [ ] When to Use (comparison to native billy implementations)

### Add Badges
- [ ] Go version badge
- [ ] Build status badge
- [ ] Code coverage badge
- [ ] Go Report Card badge
- [ ] License badge

## Acceptance Criteria
- [ ] README is complete and professional
- [ ] All code examples compile and run
- [ ] Documentation matches actual implementation
```

---

### Issue #12: Add package-level documentation (doc.go)

**Labels**: `documentation`, `godoc`, `medium-priority`

**Title**: Add doc.go with package-level documentation

**Body**:
```markdown
## Problem
No package-level godoc comment exists. The first comment in billyfs.go is on the Filesystem struct, not the package.

## Required Content
Create `doc.go` with:
- Package purpose and overview
- Key concepts (adapter pattern)
- Usage examples
- Links to billy and absfs interfaces

## Example Structure
```go
// Package billyfs provides an adapter to use absfs.SymlinkFileSystem
// implementations with the go-billy filesystem interface.
//
// This package enables seamless integration between the absfs ecosystem
// and go-git, allowing any absfs-compatible filesystem (memfs, osfs, etc.)
// to be used with git operations.
//
// Basic usage:
//   fs, _ := osfs.NewFS()
//   bfs, _ := billyfs.NewFS(fs, "/path/to/root")
//   // Use bfs with go-git
//
// For more information about absfs, see: https://github.com/absfs/absfs
// For more information about billy, see: https://github.com/go-git/go-billy
package billyfs
```

## Acceptance Criteria
- [ ] Package documentation appears in godoc
- [ ] Documentation is clear and helpful
- [ ] Includes practical examples
```

---

### Issue #13: Add CHANGELOG.md

**Labels**: `documentation`, `low-priority`

**Title**: Create CHANGELOG.md for version tracking

**Body**:
```markdown
## Problem
No changelog exists to track project evolution and breaking changes.

## Tasks
- [ ] Create CHANGELOG.md following Keep a Changelog format
- [ ] Document historical changes based on git history
- [ ] Set up process to update changelog with each release

## Format
Follow https://keepachangelog.com/en/1.0.0/ standard with sections:
- Added
- Changed
- Deprecated
- Removed
- Fixed
- Security
```

---

## 🟢 Low Priority - Nice to Have

### Issue #14: Add security policy and testing

**Labels**: `security`, `documentation`, `low-priority`

**Title**: Add SECURITY.md and security-focused tests

**Body**:
```markdown
## Tasks
- [ ] Create SECURITY.md with security policy
- [ ] Document how to report security issues
- [ ] Add security-focused tests (path traversal, etc.)
- [ ] Enable Dependabot security alerts
- [ ] Document security considerations in README

## Test Cases
- [ ] Path traversal attempts
- [ ] TempFile predictability
- [ ] Permission escalation scenarios
- [ ] Race condition testing
```

---

### Issue #15: Add fuzzing tests

**Labels**: `testing`, `security`, `low-priority`

**Title**: Add fuzzing tests for robustness

**Body**:
```markdown
## Problem
No fuzzing tests exist to discover edge cases and potential crashes.

## Tasks
- [ ] Add fuzz tests for path handling
- [ ] Add fuzz tests for file operations
- [ ] Add fuzz tests for TempFile generation
- [ ] Integrate with OSS-Fuzz if appropriate

## Reference
Use Go 1.18+ native fuzzing support.
```

---

### Issue #16: Add benchmark tests

**Labels**: `testing`, `performance`, `low-priority`

**Title**: Add benchmark tests for performance tracking

**Body**:
```markdown
## Problem
No performance benchmarks exist to track adapter overhead or prevent regressions.

## Required Benchmarks
- [ ] BenchmarkFilesystem_Create
- [ ] BenchmarkFilesystem_Open
- [ ] BenchmarkFilesystem_Read
- [ ] BenchmarkFilesystem_Write
- [ ] BenchmarkFilesystem_ReadDir
- [ ] Compare with native billy implementations

## Acceptance Criteria
- [ ] Benchmarks run in CI
- [ ] Results tracked over time
- [ ] Regressions flagged
```

---

## 📋 Meta Issues

### Issue #17: Create CONTRIBUTING.md guide

**Labels**: `documentation`, `community`, `low-priority`

**Title**: Expand contribution guidelines

**Body**:
```markdown
## Problem
Contribution guidelines are minimal (4 lines in README).

## Required Content
- [ ] Development setup instructions
- [ ] How to run tests locally
- [ ] Code style guidelines
- [ ] How to submit PRs
- [ ] Review process
- [ ] Testing requirements
- [ ] Linting requirements
```

---

## Summary

**Total Issues**: 17
- 🔴 Critical Bugs: 3
- 🔴 Critical Testing: 2
- 🟠 High Priority Infrastructure: 2
- 🟡 Medium Priority Code Quality: 3
- 🟡 Medium Priority Documentation: 3
- 🟢 Low Priority: 4

**Suggested Creation Order**:
1. Issues #1-3 (Critical bugs)
2. Issue #6 (CI/CD - enables automation)
3. Issue #4 (Comprehensive tests)
4. Issue #5 (Example tests)
5. Remaining issues as time permits

**Quick Wins** (good-first-issue):
- Issue #5: Add example tests
- Issue #10: Remove unused dependency
- Issue #13: Add CHANGELOG.md
