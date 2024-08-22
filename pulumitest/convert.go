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
func (a *PulumiTest) Convert(t PT, language string, opts ...opttest.Option) ConvertResult {
	t.Helper()

	tempDir := t.TempDir()
	base := filepath.Base(a.workingDir)
	targetDir := filepath.Join(tempDir, fmt.Sprintf("%s-%s", base, language))
	err := os.Mkdir(targetDir, 0755)
	if err != nil {
		ptFatal(t, err)
	}

	ptLogF(t, "converting to %s", language)
	cmd := exec.Command("pulumi", "convert", "--language", language, "--generate-only", "--out", targetDir)
	cmd.Dir = a.workingDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		ptFatalF(t, "failed to convert directory: %s\n%s", err, out)
	}

	options := a.options.Copy()
	for _, opt := range opts {
		opt.Apply(options)
	}

	convertedTest := &PulumiTest{
		ctx:        a.ctx,
		workingDir: targetDir,
		options:    options,
	}
	pulumiTestInit(t, convertedTest, options)
	return ConvertResult{
		PulumiTest: convertedTest,
		Output:     string(out),
	}
}
