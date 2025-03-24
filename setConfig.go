package pulumitest

import "github.com/pulumi/pulumi/sdk/v3/go/auto"

// SetConfig is a quick helper to set a config value on the current stack.
// If needing advanced options, use `CurrentStack()` to access all config methods.
func (pt *PulumiTest) SetConfig(t PT, key, value string) {
	t.Helper()

	if pt.currentStack == nil {
		ptFatal(t, "no current stack")
	}
	err := pt.currentStack.SetConfig(pt.ctx, key, auto.ConfigValue{Value: value})
	if err != nil {
		ptFatalF(t, "failed to set config: %s", err)
	}
}
