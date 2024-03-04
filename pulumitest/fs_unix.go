//go:build !windows
// +build !windows

package pulumitest

import (
	"fmt"
	"os"
	"syscall"
)

type owner struct {
	Uid int
	Gid int
}

func getFileOwner(fileInfo os.FileInfo) (*owner, error) {
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {

		return nil, fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", fileInfo.Name())
	}
	return &owner{
		Uid: int(stat.Uid),
		Gid: int(stat.Gid),
	}, nil
}
