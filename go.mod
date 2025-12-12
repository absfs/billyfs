module github.com/absfs/billyfs

go 1.23.0

require (
	github.com/absfs/absfs v0.0.0-20251208232938-aa0ca30de832
	github.com/absfs/basefs v0.0.0-20251207003147-287779ebb14c
	github.com/absfs/osfs v0.1.0-fastwalk
	github.com/go-git/go-billy/v5 v5.7.0
)

replace (
	github.com/absfs/absfs => ../absfs
	github.com/absfs/basefs => ../basefs
	github.com/absfs/fstesting => ../fstesting
	github.com/absfs/fstools => ../fstools
	github.com/absfs/osfs => ../osfs
)
