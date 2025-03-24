//go:build windows
// +build windows

package pulumitest

import (
	"os"
)

type owner struct {
	Uid int
	Gid int
}

func getFileOwner(fileInfo os.FileInfo) (*owner, error) {
	return nil, nil
}
