/*
Package billyfs provides an adapter that converts any absfs.SymlinkFileSystem
implementation into a filesystem compatible with the go-billy interface from
github.com/go-git/go-billy/v5.

# Overview

The billyfs package bridges two popular filesystem abstraction libraries in Go:
  - absfs: A flexible abstract filesystem interface (https://github.com/absfs/absfs)
  - go-billy: The filesystem interface used by go-git (https://github.com/go-git/go-billy)

This adapter allows you to use any absfs-compatible filesystem (memory-backed,
OS-based, cloud storage, etc.) with libraries that depend on the billy interface,
particularly the go-git library for Git operations.

# Key Concepts

The package uses the adapter pattern to wrap an absfs.SymlinkFileSystem and
expose it through the billy.Filesystem interface. This provides several benefits:

  1. Flexibility: Switch between different filesystem backends without changing
     your go-git code
  2. Testing: Use in-memory filesystems for fast, isolated tests
  3. Abstraction: Build applications that work across multiple storage backends
  4. Compatibility: Leverage both absfs and billy ecosystems in a single application

# Usage Example

Here's a simple example of creating a billy-compatible filesystem:

	package main

	import (
		"fmt"
		"github.com/absfs/billyfs"
		"github.com/absfs/memfs"
	)

	func main() {
		// Create an absfs filesystem (in-memory for this example)
		afs, err := memfs.NewFS()
		if err != nil {
			panic(err)
		}

		// Create the root directory
		err = afs.MkdirAll("/my/root", 0755)
		if err != nil {
			panic(err)
		}

		// Adapt it to a billy filesystem
		bfs, err := billyfs.NewFS(afs, "/my/root")
		if err != nil {
			panic(err)
		}

		// Now you can use bfs with go-git or any billy-compatible library
		file, err := bfs.Create("example.txt")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		_, err = file.Write([]byte("Hello, billyfs!"))
		if err != nil {
			panic(err)
		}

		fmt.Println("File created successfully")
	}

# Using with go-git

The primary use case for billyfs is integrating absfs filesystems with go-git:

	import (
		"github.com/absfs/billyfs"
		"github.com/absfs/memfs"
		"github.com/go-git/go-git/v5"
		"github.com/go-git/go-git/v5/storage/memory"
	)

	func cloneToMemory(url string) error {
		// Create an in-memory filesystem
		afs, _ := memfs.NewFS()
		afs.MkdirAll("/repo", 0755)

		// Adapt to billy
		bfs, _ := billyfs.NewFS(afs, "/repo")

		// Clone a repository into the memory filesystem
		_, err := git.Clone(memory.NewStorage(), bfs, &git.CloneOptions{
			URL: url,
		})
		return err
	}

# Adapter Pattern

The billyfs package implements the adapter pattern by:
  - Wrapping an absfs.SymlinkFileSystem instance
  - Implementing all billy.Filesystem interface methods
  - Delegating calls to the underlying absfs filesystem
  - Handling path and error translation between the two interfaces

This design keeps the package simple and maintainable while providing full
compatibility with both interfaces.

# Links

  - absfs interface: https://pkg.go.dev/github.com/absfs/absfs
  - billy interface: https://pkg.go.dev/github.com/go-git/go-billy/v5
  - go-git library: https://pkg.go.dev/github.com/go-git/go-git/v5

# Thread Safety

Thread safety depends on the underlying absfs.SymlinkFileSystem implementation.
The billyfs adapter itself does not add or remove any synchronization. If your
absfs filesystem is thread-safe, the resulting billy filesystem will be as well.
*/
package billyfs
