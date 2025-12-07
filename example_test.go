package billyfs_test

import (
	"fmt"
	"io"
	"os"

	"github.com/absfs/billyfs"
	"github.com/absfs/osfs"
)

// ExampleNewFS demonstrates creating a new billyfs adapter.
func ExampleNewFS() {
	// Create an osfs filesystem
	fs, err := osfs.NewFS()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Wrap it with billyfs to get a billy.Filesystem interface
	tmpDir := os.TempDir()
	bfs, err := billyfs.NewFS(fs, tmpDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Now bfs implements billy.Filesystem
	root := bfs.Root()
	fmt.Println("Root is set:", root != "")
	// Output: Root is set: true
}

// Example_create demonstrates creating and writing files.
func Example_create() {
	fs, _ := osfs.NewFS()
	bfs, _ := billyfs.NewFS(fs, os.TempDir())

	// Create a new file
	f, err := bfs.Create("example_billyfs.txt")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer f.Close()
	defer bfs.Remove("example_billyfs.txt")

	// Write to the file
	n, _ := f.Write([]byte("Hello, billyfs!"))
	fmt.Println("Wrote", n, "bytes")
	// Output: Wrote 15 bytes
}

// Example_readDir demonstrates reading directory contents.
func Example_readDir() {
	fs, _ := osfs.NewFS()
	bfs, _ := billyfs.NewFS(fs, os.TempDir())

	// Create a subdirectory for our test
	bfs.MkdirAll("example_readdir_test", 0755)
	defer bfs.Remove("example_readdir_test")

	// Create some test files in it
	f1, _ := bfs.Create("example_readdir_test/file1.txt")
	f1.Close()
	f2, _ := bfs.Create("example_readdir_test/file2.txt")
	f2.Close()
	defer bfs.Remove("example_readdir_test/file1.txt")
	defer bfs.Remove("example_readdir_test/file2.txt")

	// Read our test directory
	entries, err := bfs.ReadDir("example_readdir_test")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Found files:", len(entries))
	// Output: Found files: 2
}

// Example_mkdirAll demonstrates creating nested directories.
func Example_mkdirAll() {
	fs, _ := osfs.NewFS()
	bfs, _ := billyfs.NewFS(fs, os.TempDir())

	// Create nested directories
	err := bfs.MkdirAll("example_billyfs_dir/nested/deep", 0755)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer bfs.Remove("example_billyfs_dir/nested/deep")
	defer bfs.Remove("example_billyfs_dir/nested")
	defer bfs.Remove("example_billyfs_dir")

	// Verify it exists
	info, err := bfs.Stat("example_billyfs_dir/nested/deep")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Is directory:", info.IsDir())
	// Output: Is directory: true
}

// Example_openAndRead demonstrates opening and reading files.
func Example_openAndRead() {
	fs, _ := osfs.NewFS()
	bfs, _ := billyfs.NewFS(fs, os.TempDir())

	// Create and write a file
	f, _ := bfs.Create("example_read.txt")
	f.Write([]byte("Example content"))
	f.Close()
	defer bfs.Remove("example_read.txt")

	// Open and read the file
	f2, err := bfs.Open("example_read.txt")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer f2.Close()

	data, _ := io.ReadAll(f2)
	fmt.Println("Content:", string(data))
	// Output: Content: Example content
}

// Example_capabilities demonstrates checking filesystem capabilities.
func Example_capabilities() {
	fs, _ := osfs.NewFS()
	bfs, _ := billyfs.NewFS(fs, os.TempDir())

	caps := bfs.Capabilities()
	fmt.Println("Has all capabilities:", caps != 0)
	// Output: Has all capabilities: true
}
