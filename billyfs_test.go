package billyfs_test

import (
	"testing"

	"github.com/absfs/billyfs"
	"github.com/absfs/osfs"

	"github.com/go-git/go-billy/v5"
)

func TestBillyfs(t *testing.T) {
	var bfs billy.Filesystem
	var err error
	fs, err := osfs.NewFS()
	if err != nil {
		t.Fatal(err)
	}

	bfs, err = billyfs.NewFS(fs, "/")
	if err != nil {
		t.Fatal(err)
	}
	_ = bfs
}
