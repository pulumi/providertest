package autotest

import (
	"context"
	"path/filepath"
	"testing"
)

func TestNewStack(t *testing.T) {
	stack := NewStack(t, context.Background(), filepath.Join("testdata", "python_program"), "")
	t.Log(stack.Name())
}
