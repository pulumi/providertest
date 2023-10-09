package providertest

import (
	"context"
	"crypto/rand"
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
	stack := convertedProgram.NewStack(t, ctx)
	err = stack.SetAllConfig(ctx, pt.GetConfig())
	if err != nil {
		t.Error(err)
		return
	}
	convertedProgram.RestorePackages(t, ctx)

	_, err = stack.Preview(ctx)
	if err != nil {
		t.Error(err)
		return
	}
}

func (pt *ProviderTest) GetConfig() auto.ConfigMap {
	config := auto.ConfigMap{}
	for k, v := range pt.config {
		config[k] = auto.ConfigValue{Value: v}
	}
	return config
}

func (convertedProgram *ConvertedProgram) NewStack(t *testing.T, ctx context.Context) *auto.Stack {
	t.Helper()

	rand := RandomString()
	stackName := fmt.Sprintf("%s-%s", convertedProgram.Language, rand)
	stack, err := auto.NewStackLocalSource(ctx, stackName, convertedProgram.Dir)

	if convertedProgram.Language == "python" {
		workspace := stack.Workspace()
		settings, err := workspace.ProjectSettings(ctx)
		if err != nil {
			t.Fatalf("failed loading project settings: %v", err)
		}
		settings.Runtime.SetOption("virtualenv", "venv")
		err = workspace.SaveProjectSettings(ctx, settings)
		if err != nil {
			t.Fatalf("failed saving project settings: %v", err)
		}
	}
	if err != nil {
		t.Error(err)
		return nil
	}
	t.Cleanup(func() {
		stack.Destroy(ctx)
		stack.Workspace().RemoveStack(ctx, stackName)
	})
	return &stack
}

func RandomString() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (convertedProgram *ConvertedProgram) RestorePackages(t *testing.T, ctx context.Context) {
	t.Helper()

	t.Logf("restoring packages for %s", convertedProgram.Language)
	var commands []*exec.Cmd
	switch convertedProgram.Language {
	case "csharp":
		commands = append(commands, exec.Command("dotnet", "restore"))
	case "go":
		commands = append(commands, exec.Command("go", "mod", "download"))
	case "python":
		commands = append(commands,
			exec.Command("python3", "-m", "venv", "venv"),
			exec.Command("venv/bin/pip", "install", "-r", "requirements.txt"))
	case "typescript":
		commands = append(commands, exec.Command("npm", "install"))
	default:
		t.Errorf("unknown language %s", convertedProgram.Language)
		return
	}
	for _, cmd := range commands {
		t.Logf("running %s", cmd.String())
		cmd.Dir = convertedProgram.Dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("failed to restore packages: %s\n%s", err, out)
		} else {
			t.Log(string(out))
		}
	}
}

type ConvertedProgram struct {
	Language    string
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
		Language:    language,
	}, nil
}
