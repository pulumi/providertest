package pulumitest

import "github.com/pulumi/pulumi/sdk/v3/go/auto"

func (a *PulumiTest) SetConfig(t PT, key, value string) {
	t.Helper()

	if a.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	err := a.currentStack.SetConfig(a.ctx, key, auto.ConfigValue{Value: value})
	if err != nil {
		ptFatalF(t, "failed to set config: %s", err)
	}
}
