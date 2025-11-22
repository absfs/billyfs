# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Package-level documentation in doc.go with usage examples and key concepts
- Comprehensive README with API reference, limitations, troubleshooting, and comparison sections
- GitHub Actions CI/CD workflow for automated testing
- Security policy (SECURITY.md) with vulnerability reporting guidelines
- Security-focused tests for path traversal, permissions, and race conditions
- Runnable example tests for godoc documentation
- Comprehensive test coverage for all filesystem operations
- Build status, coverage, Go Report Card, and license badges to README

### Changed
- Use filesystem separator instead of OS separator in Join method for better cross-platform compatibility
- Wrap filesystem errors with context for improved debugging
- Replace deprecated rand.Seed with thread-safe RNG implementation
- Update minimum Go version to 1.22 for dependency compatibility

### Fixed
- Windows path handling: use filepath instead of path package
- Windows compatibility: use temp directories instead of hardcoded '/'
- Windows absolute path requirements in tests
- macOS temp directory variations in TestTempFile
- TempFile now properly respects the dir parameter
- Example_newFS uses temp directory instead of hardcoded '/tmp'
- Defensive absolute path conversion in NewFS for Windows compatibility
- Cross-platform path compatibility issues

### Security
- Added input validation in NewFS constructor
- Implemented security-focused tests for path traversal attacks
- Added tests for TempFile predictability issues
- Added permission escalation tests
- Added race condition detection tests

## [0.1.0] - Initial Release

### Added
- Core Filesystem type implementing billy.Filesystem interface
- NewFS function to create billy-compatible filesystem from absfs.SymlinkFileSystem
- File type implementing billy.File interface
- Support for all billy.Filesystem methods:
  - Basic operations: Create, Open, OpenFile, Stat, Rename, Remove
  - Path operations: Join
  - Temporary files: TempFile
  - Directory operations: ReadDir, MkdirAll
  - Symlink operations: Lstat, Symlink, Readlink
  - Advanced: Chroot, Root, Capabilities
- MIT License
- Basic README with installation and usage instructions

[Unreleased]: https://github.com/absfs/billyfs/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/absfs/billyfs/releases/tag/v0.1.0
