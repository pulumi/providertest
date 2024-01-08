package optrun

import (
	"path/filepath"

	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/deepcopy"
)

// StackName sets the default stack name to use when running the program under test.
func WithCache(path ...string) Option {
	return optionFunc(func(o *Options) {
		o.CachePath = filepath.Join(path...)
		o.EnableCache = true
	})
}

func WithOpts(opts ...opttest.Option) Option {
	return optionFunc(func(o *Options) {
		o.OptTest = append(o.OptTest, opts...)
	})
}

type Options struct {
	OptTest     []opttest.Option
	CachePath   string
	EnableCache bool
}

// Copy creates a deep copy of the current options.
func (o *Options) Copy() *Options {
	newOptions := deepcopy.Copy(*o).(Options)
	return &newOptions
}

// Defaults sets all options back to their defaults.
// This can be useful when using CopyToTempDir or Convert but not wanting to inherit any options from the previous PulumiTest.
func Defaults() Option {
	return optionFunc(func(o *Options) {
		o.EnableCache = false
		o.CachePath = ""
	})
}

func DefaultOptions() *Options {
	o := &Options{}
	Defaults().Apply(o)
	return o
}

type Option interface {
	Apply(*Options)
}

type optionFunc func(*Options)

func (o optionFunc) Apply(opts *Options) {
	o(opts)
}
