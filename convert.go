package pulumitest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pulumi/providertest/pulumitest/opttest"
)

// ConvertResult encapsulates the result of a conversion operation.
type ConvertResult struct {
	// PulumiTest instance for the converted program.
	PulumiTest *PulumiTest
	// Combined output of the `pulumi convert` command.
	Output string
}

// Create a new test by converting a program into a specific language.
// It returns a new PulumiTest instance for the converted program which will be outputted into a temporary directory.
func Convert(t PT, source, language string, opts ...opttest.Option) ConvertResult {
	t.Helper()

	pulumiTest := PulumiTest{
		ctx:        testContext(t),
		workingDir: source,
		options:    opttest.DefaultOptions(),
	}

	return pulumiTest.Convert(t, language, opts...)
}

// Convert a program to a given language.
// It returns a new PulumiTest instance for the converted program which will be outputted into a temporary directory.
func (pt *PulumiTest) Convert(t PT, language string, opts ...opttest.Option) ConvertResult {
	t.Helper()

	options := pt.options.Copy()
	for _, opt := range opts {
		opt.Apply(options)
	}

	tempDir := tempDirWithoutCleanupOnFailedTest(t, "converted", options.TempDir)
	base := filepath.Base(pt.workingDir)
	targetDir := filepath.Join(tempDir, fmt.Sprintf("%s-%s", base, language))
	err := os.Mkdir(targetDir, 0755)
	if err != nil {
		ptFatal(t, err)
	}

	ptLogF(t, "converting to %s", language)
	cmd := exec.Command("pulumi", "convert", "--language", language, "--generate-only", "--out", targetDir)
	cmd.Dir = pt.workingDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		ptFatalF(t, "failed to convert directory: %s\n%s", err, out)
	}

	convertedTest := &PulumiTest{
		ctx:        pt.ctx,
		workingDir: targetDir,
		options:    options,
	}
	pulumiTestInit(t, convertedTest, options)
	return ConvertResult{
		PulumiTest: convertedTest,
		Output:     string(out),
	}
}
