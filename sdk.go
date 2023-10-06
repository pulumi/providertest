package providertest

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func (pt *ProviderTest) RunSdk(t *testing.T, language string) {
	t.Helper()

	convertedProgram, err := pt.ConvertProgram(t, language)
	if err != nil {
		t.Errorf("failed to convert program: %v", err)
		return
	}
	ctx := context.Background()
	stackName := fmt.Sprintf("test-%s", language)
	stack, err := auto.NewStackLocalSource(ctx, stackName, convertedProgram.Dir)
	if err != nil {
		t.Error(err)
		return
	}
	t.Cleanup(func() {
		stack.Destroy(ctx)
	})

	_, err = stack.Preview(ctx)
	if err != nil {
		t.Error(err)
		return
	}
}

type ConvertedProgram struct {
	Dir      string
	EditDirs []EditDir
}

func (pt *ProviderTest) ConvertProgram(t *testing.T, language string) (*ConvertedProgram, error) {
	t.Helper()

	convertedDir := t.TempDir()

	t.Logf("converting to %s", language)
	cmd := exec.Command("pulumi", "convert", "--language", language, "--generate-only", "--out", convertedDir)
	cmd.Dir = pt.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to convert directory: %s\n%s", err, out)
	}

	convertedEditDirs := make([]EditDir, len(pt.editDirs))
	for i, ed := range pt.editDirs {
		convertedEditDir := t.TempDir()
		convertedEditDirs[i].dir = convertedEditDir
		convertedEditDirs[i].clean = ed.clean

		t.Logf("converting to %s", language)
		cmd := exec.Command("pulumi", "convert", "--language", language, "--generate-only", "--out", convertedDir)
		cmd.Dir = ed.dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to convert edit directory %d: %s\n%s", i+1, err, out)
		}
	}

	return &ConvertedProgram{
		Dir:      convertedDir,
		EditDirs: convertedEditDirs,
	}, nil
}
