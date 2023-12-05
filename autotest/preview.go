package autotest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func (a *AutoTest) Preview(stack *auto.Stack) auto.PreviewResult {
	a.t.Helper()

	a.t.Log("previewing update")
	result, err := stack.Preview(a.ctx)
	if err != nil {
		a.t.Fatalf("failed to preview update: %s", err)
	}
	return result
}
