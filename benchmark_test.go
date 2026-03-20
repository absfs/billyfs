package billyfs_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/absfs/billyfs"
	"github.com/absfs/osfs"
)

// newBenchFS creates a filesystem for benchmarking
func newBenchFS(b *testing.B) *billyfs.Filesystem {
	b.Helper()
	tmpDir := b.TempDir()

	fs, err := osfs.NewFS()
	if err != nil {
		b.Fatalf("failed to create osfs: %v", err)
	}

	bfs, err := billyfs.NewFS(fs, tmpDir)
	if err != nil {
		b.Fatalf("failed to create billyfs: %v", err)
	}

	return bfs
}

// BenchmarkCreate measures file creation performance
func BenchmarkCreate(b *testing.B) {
	bfs := newBenchFS(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, err := bfs.Create("bench_file.txt")
		if err != nil {
			b.Fatal(err)
		}
		f.Close()
		bfs.Remove("bench_file.txt")
	}
}

// BenchmarkWrite measures write performance at various data sizes
func BenchmarkWrite(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"4KB", 4 * 1024},
		{"64KB", 64 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			bfs := newBenchFS(b)
			data := make([]byte, size.size)

			b.ReportAllocs()
			b.SetBytes(int64(size.size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, _ := bfs.Create("bench_write.txt")
				f.Write(data)
				f.Close()
			}
			b.StopTimer()
			bfs.Remove("bench_write.txt")
		})
	}
}

// BenchmarkRead measures read performance at various data sizes
func BenchmarkRead(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"4KB", 4 * 1024},
		{"64KB", 64 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			bfs := newBenchFS(b)

			// Setup: create file with data
			f, _ := bfs.Create("bench_read.txt")
			data := make([]byte, size.size)
			f.Write(data)
			f.Close()

			b.ReportAllocs()
			b.SetBytes(int64(size.size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, _ := bfs.Open("bench_read.txt")
				io.ReadAll(f)
				f.Close()
			}
			b.StopTimer()
			bfs.Remove("bench_read.txt")
		})
	}
}

// BenchmarkStat measures stat performance
func BenchmarkStat(b *testing.B) {
	bfs := newBenchFS(b)

	f, _ := bfs.Create("bench_stat.txt")
	f.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfs.Stat("bench_stat.txt")
	}
	b.StopTimer()
	bfs.Remove("bench_stat.txt")
}

// BenchmarkReadDir measures directory read performance with varying directory sizes
func BenchmarkReadDir(b *testing.B) {
	sizes := []struct {
		name  string
		files int
	}{
		{"10_files", 10},
		{"100_files", 100},
		{"1000_files", 1000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			bfs := newBenchFS(b)

			bfs.MkdirAll("bench_dir", 0755)
			for i := 0; i < size.files; i++ {
				f, _ := bfs.Create(fmt.Sprintf("bench_dir/file%04d.txt", i))
				f.Close()
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bfs.ReadDir("bench_dir")
			}
		})
	}
}

// BenchmarkMkdirAll measures directory creation performance
func BenchmarkMkdirAll(b *testing.B) {
	bfs := newBenchFS(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfs.MkdirAll("a/b/c/d", 0755)
		bfs.Remove("a/b/c/d")
		bfs.Remove("a/b/c")
		bfs.Remove("a/b")
		bfs.Remove("a")
	}
}

// BenchmarkOpen measures file open performance
func BenchmarkOpen(b *testing.B) {
	bfs := newBenchFS(b)

	f, _ := bfs.Create("bench_open.txt")
	f.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := bfs.Open("bench_open.txt")
		f.Close()
	}
	b.StopTimer()
	bfs.Remove("bench_open.txt")
}

// BenchmarkRename measures rename performance
func BenchmarkRename(b *testing.B) {
	bfs := newBenchFS(b)

	f, _ := bfs.Create("bench_rename_a.txt")
	f.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			bfs.Rename("bench_rename_a.txt", "bench_rename_b.txt")
		} else {
			bfs.Rename("bench_rename_b.txt", "bench_rename_a.txt")
		}
	}
}

// BenchmarkJoin measures path join performance
func BenchmarkJoin(b *testing.B) {
	bfs := newBenchFS(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfs.Join("a", "b", "c", "d", "e")
	}
}

// BenchmarkAdapterOverhead compares billyfs adapter operations against
// direct os operations to measure the adapter's overhead.
func BenchmarkAdapterOverhead(b *testing.B) {
	b.Run("Create", func(b *testing.B) {
		b.Run("billyfs", func(b *testing.B) {
			bfs := newBenchFS(b)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, _ := bfs.Create("overhead.txt")
				f.Close()
				bfs.Remove("overhead.txt")
			}
		})

		b.Run("os_direct", func(b *testing.B) {
			dir := b.TempDir()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, _ := os.Create(dir + "/overhead.txt")
				f.Close()
				os.Remove(dir + "/overhead.txt")
			}
		})
	})

	b.Run("Stat", func(b *testing.B) {
		b.Run("billyfs", func(b *testing.B) {
			bfs := newBenchFS(b)
			f, _ := bfs.Create("overhead.txt")
			f.Close()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bfs.Stat("overhead.txt")
			}
		})

		b.Run("os_direct", func(b *testing.B) {
			dir := b.TempDir()
			f, _ := os.Create(dir + "/overhead.txt")
			f.Close()
			path := dir + "/overhead.txt"
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				os.Stat(path)
			}
		})
	})

	b.Run("ReadWrite_4KB", func(b *testing.B) {
		data := make([]byte, 4096)

		b.Run("billyfs", func(b *testing.B) {
			bfs := newBenchFS(b)
			b.ReportAllocs()
			b.SetBytes(4096)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, _ := bfs.Create("overhead.txt")
				f.Write(data)
				f.Close()
				f, _ = bfs.Open("overhead.txt")
				io.ReadAll(f)
				f.Close()
			}
		})

		b.Run("os_direct", func(b *testing.B) {
			dir := b.TempDir()
			path := dir + "/overhead.txt"
			b.ReportAllocs()
			b.SetBytes(4096)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, _ := os.Create(path)
				f.Write(data)
				f.Close()
				f, _ = os.Open(path)
				io.ReadAll(f)
				f.Close()
			}
		})
	})
}
