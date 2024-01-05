package pulumitest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pulumi/providertest/pulumitest/opttest"
)

type ConvertResult struct {
	// PulumiTest instance for the converted program.
	PulumiTest *PulumiTest
	// Combined output of the `pulumi convert` command.
	Output string
}

// Convert a program to a given language.
// It returns a new PulumiTest instance for the converted program which will be outputted into a temporary directory.
func (a *PulumiTest) Convert(language string, opts ...opttest.Option) ConvertResult {
	a.t.Helper()

	tempDir := a.t.TempDir()
	base := filepath.Base(a.source)
	targetDir := filepath.Join(tempDir, fmt.Sprintf("%s-%s", base, language))
	err := os.Mkdir(targetDir, 0755)
	if err != nil {
		a.t.Fatal(err)
	}

	a.t.Logf("converting to %s", language)
	cmd := exec.Command("pulumi", "convert", "--language", language, "--generate-only", "--out", targetDir)
	cmd.Dir = a.source
	out, err := cmd.CombinedOutput()
	if err != nil {
		a.t.Fatalf("failed to convert directory: %s\n%s", err, out)
	}

	options := a.options.Copy()
	for _, opt := range opts {
		opt.Apply(options)
	}

	convertedTest := &PulumiTest{
		t:       a.t,
		ctx:     a.ctx,
		source:  targetDir,
		options: options,
	}
	if !options.SkipInstall {
		convertedTest.Install()
	}
	if !options.SkipStackCreate {
		convertedTest.NewStack(options.StackName)
	}
	return ConvertResult{
		PulumiTest: convertedTest,
		Output:     string(out),
	}
}
