# Contributing to billyfs

Thank you for your interest in contributing to billyfs! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project follows the standard open source code of conduct. Be respectful and professional in all interactions.

## How to Contribute

### Reporting Issues

If you find a bug or have a feature request:

1. Check the [issue tracker](https://github.com/absfs/billyfs/issues) to see if it's already reported
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - Go version and OS information
   - Code samples if applicable

### Submitting Changes

1. **Fork the repository** and create a new branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the coding standards below

3. **Test your changes:**
   ```bash
   go test -v ./...
   ```

4. **Commit your changes** with clear, descriptive commit messages:
   ```bash
   git commit -m "Add feature: brief description"
   ```

5. **Push to your fork** and submit a pull request

## Coding Standards

### Code Quality

- **Formatting:** All code must pass `gofmt`. Run `gofmt -w .` before committing
- **Vetting:** Code must pass `go vet ./...` without warnings
- **Testing:** New features must include tests

### Testing Requirements

- All new code must have test coverage
- Run tests with race detection: `go test -race ./...`
- Ensure tests pass on Linux, macOS, and Windows
- Run benchmarks to check for performance regressions: `go test -bench=. -benchmem ./...`
- Run fuzz tests locally before submitting changes to path handling: `go test -fuzz=. -fuzztime=30s`

### Code Style

- Follow standard Go conventions and idioms
- Keep functions focused and reasonably sized
- Use meaningful variable and function names
- Add comments for exported functions and types
- Include package-level documentation

### Documentation

- Update README.md if adding new features
- Add godoc examples for new public APIs
- Update CHANGELOG.md following [Keep a Changelog](https://keepachangelog.com/) format
- Document any breaking changes clearly

## Development Workflow

### Setting Up Development Environment

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/billyfs.git
cd billyfs

# Add upstream remote
git remote add upstream https://github.com/absfs/billyfs.git

# Install dependencies
go mod download

# Run tests
go test ./...
```

### Running Benchmarks

```bash
# Run all benchmarks with memory allocation stats
go test -bench=. -benchmem ./...

# Run a specific benchmark
go test -bench=BenchmarkWrite -benchmem ./...
```

### Running Fuzz Tests

```bash
# Run all fuzz tests for 30 seconds each
go test -fuzz=FuzzCreateFile -fuzztime=30s
go test -fuzz=FuzzReadWrite -fuzztime=30s
go test -fuzz=FuzzSymlink -fuzztime=30s

# Run a specific fuzz test longer for deeper coverage
go test -fuzz=FuzzReadWrite -fuzztime=5m
```

### Before Submitting PR

Run this checklist:

- [ ] Code is formatted with `gofmt -w .`
- [ ] All tests pass: `go test -v ./...`
- [ ] No `go vet` warnings: `go vet ./...`
- [ ] Race detector passes: `go test -race ./...`
- [ ] Added tests for new functionality
- [ ] Updated documentation (README, CHANGELOG, godoc)
- [ ] Commit messages are clear and descriptive
- [ ] Branch is up to date with main

### Pull Request Guidelines

- Keep PRs focused on a single feature or fix
- Write clear PR descriptions explaining what and why
- Reference related issues in PR description
- Respond to review feedback promptly
- Ensure CI checks pass

## Architecture Guidelines

### billy Interface Compliance

billyfs adapts `absfs.SymlinkFileSystem` to the `go-billy/v5` `billy.Filesystem` interface. Any changes must maintain full compliance with both interfaces.

### Path Handling

billyfs uses `path` (not `path/filepath`) for all path operations to ensure consistent Unix-style separators regardless of platform. This is consistent with the absfs convention of using `/` as the path separator.

### Code Organization

- `billyfs.go` - Core filesystem adapter implementation
- `billyfile.go` - File wrapper implementation
- `billyfs_test.go` - Filesystem test suite
- `file_test.go` - File operation tests
- `example_test.go` - Godoc examples
- `benchmark_test.go` - Performance benchmarks
- `fuzz_test.go` - Fuzz tests

## Questions?

If you have questions about contributing, feel free to:

- Open an issue for discussion
- Check the [absfs documentation](https://github.com/absfs/absfs)
- Review existing code and tests for patterns

Thank you for contributing to billyfs!
