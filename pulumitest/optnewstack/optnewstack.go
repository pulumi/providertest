package optnewstack

import "github.com/pulumi/pulumi/sdk/v3/go/auto"

// DisableAttach will configure the provider binary in the program's Pulumi.yaml rather than attaching the running provider.
func DisableAutoDestroy() NewStackOpt {
	return optionFunc(func(o *NewStackOptions) {
		o.SkipDestroy = true
	})
}

func EnableAutoDestroy() NewStackOpt {
	return optionFunc(func(o *NewStackOptions) {
		o.SkipDestroy = false
	})
}

func WithOpts(opts ...auto.LocalWorkspaceOption) NewStackOpt {
	return optionFunc(func(o *NewStackOptions) {
		o.Opts = opts
	})
}

type NewStackOptions struct {
	SkipDestroy bool
	Opts        []auto.LocalWorkspaceOption
}

type NewStackOpt interface {
	Apply(*NewStackOptions)
}

func Defaults() NewStackOptions {
	return NewStackOptions{
		SkipDestroy: false,
	}
}

type optionFunc func(*NewStackOptions)

func (o optionFunc) Apply(opts *NewStackOptions) {
	o(opts)
}
