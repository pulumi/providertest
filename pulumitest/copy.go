package pulumitest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pulumi/providertest/pulumitest/opttest"
)

// CopyToTempDir copies the program to a temporary directory.
// It returns a new PulumiTest instance for the copied program.
// This is used to avoid temporary files being written to the source directory.
func (a *PulumiTest) CopyToTempDir(opts ...opttest.Option) *PulumiTest {
	a.t.Helper()
	tempDir := a.t.TempDir()

	// Maintain the directory name in the temp dir as this might be used for stack naming.
	sourceBase := filepath.Base(a.source)
	destination := filepath.Join(tempDir, sourceBase)
	err := os.Mkdir(destination, 0755)
	if err != nil {
		a.t.Fatal(err)
	}

	return a.CopyTo(destination)
}

// CopyTo copies the program to the specified directory.
// It returns a new PulumiTest instance for the copied program.
func (a *PulumiTest) CopyTo(dir string, opts ...opttest.Option) *PulumiTest {
	a.t.Helper()

	err := copyDirectory(a.source, dir)
	if err != nil {
		a.t.Fatal(err)
	}

	options := a.options.Copy()
	for _, opt := range opts {
		opt.Apply(options)
	}
	newTest := &PulumiTest{
		t:       a.t,
		ctx:     a.ctx,
		source:  dir,
		options: options,
	}
	if !options.SkipInstall {
		newTest.Install()
	}
	if !options.SkipStackCreate {
		newTest.NewStack(options.StackName)
	}
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

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
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

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
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

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func createIfNotExists(dir string, perm os.FileMode) error {
	if exists(dir) {
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
