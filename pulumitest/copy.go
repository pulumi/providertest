package pulumitest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pulumi/providertest/pulumitest/opttest"
)

// CopyToTempDir copies the program to a temporary directory.
// It returns a new PulumiTest instance for the copied program.
// This is used to avoid temporary files being written to the source directory.
func (a *PulumiTest) CopyToTempDir(t PT, opts ...opttest.Option) *PulumiTest {
	t.Helper()
	options := a.options.Copy()
	for _, opt := range opts {
		opt.Apply(options)
	}
	tempDir := tempDirWithoutCleanupOnFailedTest(t, "programDir", options.TempDir)

	// Maintain the directory name in the temp dir as this might be used for stack naming.
	sourceBase := filepath.Base(a.workingDir)
	destination := filepath.Join(tempDir, sourceBase)
	err := os.Mkdir(destination, 0755)
	if err != nil {
		ptFatal(t, err)
	}

	return a.CopyTo(t, destination, opts...)
}

// CopyTo copies the program to the specified directory.
// It returns a new PulumiTest instance for the copied program.
func (a *PulumiTest) CopyTo(t PT, dir string, opts ...opttest.Option) *PulumiTest {
	t.Helper()

	err := copyDirectory(a.workingDir, dir)
	if err != nil {
		ptFatal(t, err)
	}

	options := a.options.Copy()
	for _, opt := range opts {
		opt.Apply(options)
	}
	newTest := &PulumiTest{
		ctx:        a.ctx,
		workingDir: dir,
		options:    options,
	}
	pulumiTestInit(t, newTest, options)
	return newTest
}

func copyDirectory(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		owner, err := getFileOwner(fileInfo)
		if err != nil {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s': %w", sourcePath, err)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := copyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := copySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		// On Windows, os.Lchown always returns the syscall.EWINDOWS error, wrapped in
		// *PathError. But on Windows owner will be nil.
		if owner != nil {
			if err := os.Lchown(destPath, owner.Uid, owner.Gid); err != nil {
				return err
			}
		}

		fInfo, err := entry.Info()
		if err != nil {
			return err
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func exists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	switch {
	case err == nil:
		return true, nil
	case !os.IsNotExist(err):
		return false, err
	}
	return false, nil
}

func createIfNotExists(dir string, perm os.FileMode) error {
	exists, err := exists(dir)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func copySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
