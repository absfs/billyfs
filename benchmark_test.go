package billyfs_test

import (
	"io"
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

// BenchmarkWrite measures write performance
func BenchmarkWrite(b *testing.B) {
	bfs := newBenchFS(b)
	data := make([]byte, 1024) // 1KB of data

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := bfs.Create("bench_write.txt")
		f.Write(data)
		f.Close()
	}
	b.StopTimer()
	bfs.Remove("bench_write.txt")
}

// BenchmarkRead measures read performance
func BenchmarkRead(b *testing.B) {
	bfs := newBenchFS(b)

	// Setup: create file with data
	f, _ := bfs.Create("bench_read.txt")
	data := make([]byte, 1024)
	f.Write(data)
	f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := bfs.Open("bench_read.txt")
		io.ReadAll(f)
		f.Close()
	}
	b.StopTimer()
	bfs.Remove("bench_read.txt")
}

// BenchmarkStat measures stat performance
func BenchmarkStat(b *testing.B) {
	bfs := newBenchFS(b)

	// Setup
	f, _ := bfs.Create("bench_stat.txt")
	f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfs.Stat("bench_stat.txt")
	}
	b.StopTimer()
	bfs.Remove("bench_stat.txt")
}

// BenchmarkReadDir measures directory read performance
func BenchmarkReadDir(b *testing.B) {
	bfs := newBenchFS(b)

	// Setup: create directory with files
	bfs.MkdirAll("bench_dir", 0755)
	for i := 0; i < 100; i++ {
		f, _ := bfs.Create(bfs.Join("bench_dir", "file"+string(rune('0'+i%10))+".txt"))
		f.Close()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfs.ReadDir("bench_dir")
	}
}

// BenchmarkMkdirAll measures directory creation performance
func BenchmarkMkdirAll(b *testing.B) {
	bfs := newBenchFS(b)

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

	// Setup
	f, _ := bfs.Create("bench_open.txt")
	f.Close()

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfs.Join("a", "b", "c", "d", "e")
	}
}
