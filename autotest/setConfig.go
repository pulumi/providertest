package autotest

import "github.com/pulumi/pulumi/sdk/v3/go/auto"

func (a *AutoTest) SetConfig(key, value string) {
	a.t.Helper()

	if a.currentStack == nil {
		a.t.Fatal("no current stack")
	}
	err := a.currentStack.SetConfig(a.ctx, key, auto.ConfigValue{Value: value})
	if err != nil {
		a.t.Fatalf("failed to set config: %s", err)
	}
}
