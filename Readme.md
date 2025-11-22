# billyfs - Abstract File System Adapter

[![Go Reference](https://pkg.go.dev/badge/github.com/absfs/billyfs.svg)](https://pkg.go.dev/github.com/absfs/billyfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/absfs/billyfs)](https://goreportcard.com/report/github.com/absfs/billyfs)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/absfs/billyfs/blob/master/LICENSE)
[![Build Status](https://github.com/absfs/billyfs/workflows/Go/badge.svg)](https://github.com/absfs/billyfs/actions)

The `billyfs` package enables seamless conversion of any
`absfs.SymlinkFileSystem` into a file system compatible with
`github.com/go-git/go-billy/v5`. This integration allows developers to utilize
the robust features of the `go-git` package across a variety of file systems
that implement the `absfs.SymlinkFileSystem` interface. By bridging these
technologies, `billyfs` facilitates more flexible and efficient management of
file systems within Go applications, especially for those requiring git
functionality.

See the [absfs](github.com/absfs/absfs) package for more information about the
`absfs.SymlinkFileSystem` interface and the `absfs` abstract file system API.

## Installation

Install `billyfs` using the typical `go get` command:

```bash
$ go get github.com/absfs/billyfs
```


## Example Usage

The following example demonstrates how to use `billyfs` to create a UNIX-like
folder structure and manipulate files within a custom file system setup. This
example provides a practical insight into integrating `billyfs` into your
projects.

```go
package main

import (
    "fmt"
    slash "path"
    "sort"

    "github.com/absfs/absfs"
    "github.com/absfs/billyfs"
    "github.com/absfs/memfs"
    "github.com/absfs/osfs"
)

// Choose the file system type: "memfs" for in-memory or "osfs" for OS file
// system, or add any other file system that implements
// `the absfs.SymlinkFileSystem` interface.
const FS = "memfs"

// List of UNIX-like directories to create in the file system.
var UNIXLikeDirs = []string{
    "/bin", "/etc", "/tmp", "/var", "/opt", "/home", "/root", "/mnt", "/media",
    "/srv", "/lib", "/usr/local/src", "/usr/bin", "/usr/sbin", "/usr/lib",
    "/usr/include", "/usr/share", "/usr/src",
}

func main() {
    var fs absfs.SymlinkFileSystem
    var err error

    // Initialize the file system based on the FS constant.
    switch FS {
    case "memfs":
        fs, err = memfs.NewFS()
    case "osfs":
        fs, err = osfs.NewFS()
    }
    if err != nil {
        panic(err)
    }

    // Create a UNIX-like folder structure as a demonstration.
    for _, dir := range UNIXLikeDirs {
        err := fs.MkdirAll(dir, 0755)
        if err != nil {
            panic(err)
        }
    }

    // Use billyfs to adapt the absfs file system.
    bfs, err := billyfs.NewFS(fs, "/usr/local/src")
    if err != nil {
        panic(err)
    }

    // Create and remove a test file to demonstrate file manipulation.
    f, _ := bfs.Create("/test.txt")
    defer func() {
        err := f.Close()
        if err != nil {
            panic(err)
        }
        err = bfs.Remove("/test.txt")
        if err != nil {
            panic(err)
        }
    }()
    bfsdata := []byte("This demonstrates file manipulation within the adapted file system.\n")
    _, err = f.Write(bfsdata)
    if err != nil {
        panic(err)
    }

    // List files in the current directory using billyfs
    files, err := bfs.ReadDir("/")
    if err != nil {
        panic(err)
    }

    var names []string
    for _, file := range files {
        names = append(names, file.Name())
    }
    sort.Strings(names)

    fmt.Println("Files in /usr/local/src:")
    for _, name := range names {
        fmt.Printf("  - %s\n", name)
    }
}
```

This example showcases basic operations like creating directories, files, and
listing contents within a file system adapted by `billyfs`. Remember to adapt
the constants and variable values to fit your specific use case.

## API Reference

### Core Types

#### `type Filesystem`

The `Filesystem` type wraps an `absfs.SymlinkFileSystem` and implements the `billy.Filesystem` interface.

```go
type Filesystem struct {
    // contains filtered or unexported fields
}
```

### Functions

#### `NewFS`

```go
func NewFS(fs absfs.SymlinkFileSystem, dir string) (*Filesystem, error)
```

Creates a new billy-compatible filesystem from an `absfs.SymlinkFileSystem`. The `dir` parameter specifies the root directory for the billy filesystem and must be an absolute path that already exists.

**Parameters:**
- `fs`: An implementation of `absfs.SymlinkFileSystem`
- `dir`: Absolute path to the root directory

**Returns:** A `*Filesystem` or an error if the directory doesn't exist or is invalid.

### Methods

The `Filesystem` type implements all methods from the `billy.Filesystem` interface:

**Basic Operations:**
- `Create(filename string) (billy.File, error)` - Create a new file
- `Open(filename string) (billy.File, error)` - Open a file for reading
- `OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error)` - Open with specific flags
- `Stat(filename string) (os.FileInfo, error)` - Get file information
- `Rename(oldpath, newpath string) error` - Rename a file or directory
- `Remove(filename string) error` - Remove a file or directory
- `Join(elem ...string) string` - Join path elements
- `TempFile(dir, prefix string) (billy.File, error)` - Create a temporary file
- `ReadDir(path string) ([]os.FileInfo, error)` - Read directory contents
- `MkdirAll(filename string, perm os.FileMode) error` - Create directory tree
- `Lstat(filename string) (os.FileInfo, error)` - Get file info (doesn't follow symlinks)
- `Symlink(target, link string) error` - Create a symbolic link
- `Readlink(link string) (string, error)` - Read symbolic link target
- `Chroot(path string) (billy.Filesystem, error)` - Create a chrooted filesystem
- `Root() string` - Get the root path

## Limitations

1. **Path Handling**: The package uses the filesystem's separator (obtained via `fs.Separator()`) for path operations. This ensures consistency with the underlying filesystem but may differ from OS path separators.

2. **Absolute Paths Required**: The `NewFS` function requires an absolute path. Relative paths will be converted to absolute paths using `filepath.Abs()`.

3. **Existing Directory**: The root directory specified in `NewFS` must already exist in the underlying filesystem.

4. **Symlink Support**: The underlying filesystem must implement `absfs.SymlinkFileSystem`. Filesystems that don't support symlinks may return errors for symlink operations.

5. **Capabilities**: The `Capabilities()` method reports the filesystem as case-sensitive and supporting all Billy features. The actual capabilities depend on the underlying `absfs.SymlinkFileSystem` implementation.

## Troubleshooting

### "not a directory" Error

**Problem**: `NewFS` returns "not a directory" error.

**Solution**: Ensure the path provided to `NewFS` points to an existing directory, not a file.

```go
// Create the directory first if it doesn't exist
err := fs.MkdirAll("/my/root/path", 0755)
if err != nil {
    panic(err)
}

// Then create the billy filesystem
bfs, err := billyfs.NewFS(fs, "/my/root/path")
```

### Path Separator Issues

**Problem**: Paths work on some systems but not others.

**Solution**: Use the `Join` method provided by the filesystem instead of manually constructing paths:

```go
// Good - portable across filesystems
path := bfs.Join("dir", "subdir", "file.txt")

// Avoid - may not work on all filesystems
path := "/dir/subdir/file.txt"
```

### "invalid argument" Error

**Problem**: `NewFS` returns `os.ErrInvalid`.

**Solution**: The directory path cannot be empty. Provide a valid absolute path.

## Comparison with Other Adapters

### billyfs vs. Direct billy Implementation

- **billyfs**: Adapts any `absfs.SymlinkFileSystem` to billy, allowing you to use abstract filesystems (memory, OS, custom) with go-git
- **Direct billy**: Implementations like `memfs` or `osfs` from go-billy are tightly coupled to the billy interface

### When to Use billyfs

Use `billyfs` when:
- You need to use go-git with an existing `absfs.SymlinkFileSystem` implementation
- You want to switch between different filesystem backends (memory, OS, cloud storage) without changing your go-git code
- You're building a system that benefits from the abstraction layer provided by absfs

Use direct billy implementations when:
- You only need go-git functionality with standard filesystems
- You don't need the additional abstraction layer
- Performance is critical and you want to minimize adapter overhead

## Contributing

We warmly welcome contributions to the `billyfs` project. Whether you're fixing
bugs, adding new features, or improving the documentation, your help is greatly
appreciated. Please follow these steps to contribute:

1. Fork the repository.
2. Create a new branch for your feature or fix.
3. Commit your changes with clear, descriptive messages.
4. Push your changes and submit a pull request.

We will review your pull request as soon as possible. Thank you for your
contributing to the `absfs` ecosystem!

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/billyfs/blob/master/LICENSE) for more information.