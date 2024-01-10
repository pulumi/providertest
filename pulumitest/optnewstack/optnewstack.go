package optnewstack

import "github.com/pulumi/pulumi/sdk/v3/go/auto"

// DisableAutoDestroy will skip running `pulumi destroy` at the end of the test.
func DisableAutoDestroy() NewStackOpt {
	return optionFunc(func(o *NewStackOptions) {
		o.SkipDestroy = true
	})
}

// EnableAutoDestroy will ensure run `pulumi destroy` at the end of the test.
func EnableAutoDestroy() NewStackOpt {
	return optionFunc(func(o *NewStackOptions) {
		o.SkipDestroy = false
	})
}

// WithOpts adds additional workspace options for the context of the run.
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
