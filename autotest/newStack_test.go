package autotest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStackPreview(t *testing.T) {
	test := NewAutoTest(t, filepath.Join("testdata", "python_program"))
	test = test.CopyToTempDir()
	test.Install()
	stack := test.NewStack("")
	t.Log(stack.Name())
	preview := test.Preview(stack)
	assert.Equal(t, 1, len(preview.ChangeSummary))
}
