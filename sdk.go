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
	Dir         string
	UpdateSteps []UpdateStep
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

	convertedEditDirs := make([]UpdateStep, len(pt.updateSteps))
	for i, updateStep := range pt.updateSteps {
		convertedEditDirs[i].pt = pt
		if updateStep.dir == nil {
			return nil, fmt.Errorf("update step %d has no changes specified", i+1)
		}
		convertedEditDir := t.TempDir()
		convertedEditDirs[i].dir = &convertedEditDir
		convertedEditDirs[i].clean = updateStep.clean

		t.Logf("converting to %s", language)
		cmd := exec.Command("pulumi", "convert", "--language", language, "--generate-only", "--out", convertedDir)
		cmd.Dir = *updateStep.dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to convert update step %d: %s\n%s", i+1, err, out)
		}
	}

	return &ConvertedProgram{
		Dir:         convertedDir,
		UpdateSteps: convertedEditDirs,
	}, nil
}
