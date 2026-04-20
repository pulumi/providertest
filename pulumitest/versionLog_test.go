package pulumitest_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/stretchr/testify/assert"
)

type logCaptureT struct {
	*testing.T
	mu   sync.Mutex
	logs []string
}

func (t *logCaptureT) Log(args ...any) {
	line := fmt.Sprint(args...)
	t.mu.Lock()
	t.logs = append(t.logs, line)
	t.mu.Unlock()
	t.T.Log(args...)
}

func (t *logCaptureT) logged(substr string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, line := range t.logs {
		if strings.Contains(line, substr) {
			return true
		}
	}
	return false
}

func TestLogsPulumiVersionOnStackCreation(t *testing.T) {
	t.Parallel()
	ct := &logCaptureT{T: t}

	pulumitest.NewPulumiTest(ct, "testdata/yaml_program")

	assert.True(t,
		ct.logged("pulumi version:") || ct.logged("failed to get pulumi version"),
		"expected pulumi version success or failure to be logged")
	assert.True(t,
		ct.logged("pulumi plugins:") || ct.logged("failed to list pulumi plugins"),
		"expected pulumi plugins success or failure to be logged")
}

func TestDisablePulumiVersionLogSuppressesOutput(t *testing.T) {
	t.Parallel()
	ct := &logCaptureT{T: t}

	pulumitest.NewPulumiTest(ct, "testdata/yaml_program", opttest.DisablePulumiVersionLog())

	assert.False(t, ct.logged("pulumi version:"), "expected pulumi version log to be suppressed")
	assert.False(t, ct.logged("pulumi plugins:"), "expected pulumi plugins log to be suppressed")
	assert.False(t, ct.logged("failed to get pulumi version"), "expected pulumi version failure log to be suppressed")
	assert.False(t, ct.logged("failed to list pulumi plugins"), "expected pulumi plugins failure log to be suppressed")
}
