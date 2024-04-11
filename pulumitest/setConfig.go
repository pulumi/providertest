package pulumitest

import "github.com/pulumi/pulumi/sdk/v3/go/auto"

func (a *PulumiTest) SetConfig(key, value string) {
	a.t.Helper()

	if a.currentStack == nil {
		a.fatal("no current stack")
	}
	err := a.currentStack.SetConfig(a.ctx, key, auto.ConfigValue{Value: value})
	if err != nil {
		a.fatalf("failed to set config: %s", err)
	}
}
