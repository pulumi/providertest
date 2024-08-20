package pulumiyaml

import "fmt"

type Program struct {
	// Defaults to "test"
	Name string `yaml:"name,omitempty"`
	// Defaults to "yaml"
	Runtime     string              `yaml:"runtime,omitempty"`
	Description string              `yaml:"description,omitempty"`
	Config      map[string]any      `yaml:"config"`
	Plugins     Plugins             `yaml:"plugins"`
	Resources   map[string]Resource `yaml:"resources"`
}

func NewProgram() *Program {
	return &Program{
		Name:      "test",
		Runtime:   "yaml",
		Config:    make(map[string]any),
		Resources: make(map[string]Resource),
	}
}

// AddResource adds a resource to the program by name.
// Returns an error if the resource already exists in the program.
func (p *Program) AddResource(name string, resource Resource) error {
	if p.Resources == nil {
		p.Resources = make(map[string]Resource)
	}
	if _, ok := p.Resources[name]; ok {
		return fmt.Errorf("resource %q already exists in program", name)
	}
	p.Resources[name] = resource
	return nil
}

// RemoveResource removes a resource from the program by name and returns it, if found.
func (p *Program) RemoveResource(name string) (*Resource, error) {
	if p.Resources == nil {
		return nil, fmt.Errorf("resource %q not found in program", name)
	}
	removed, found := p.Resources[name]
	if !found {
		return nil, fmt.Errorf("resource %q not found in program", name)
	}
	delete(p.Resources, name)
	return &removed, nil
}

// GetResource returns a resource from the program by name.
func (p *Program) GetResource(name string) (*Resource, error) {
	resource, found := p.Resources[name]
	if !found {
		return nil, fmt.Errorf("resource %q not found in program", name)
	}
	return &resource, nil
}

// UpdateResource updates a resource in the program by name.
// Returns the old resource and true if the resource was found and updated.
func (p *Program) UpdateResource(name string, resource Resource) (*Resource, error) {
	oldResource, found := p.Resources[name]
	if !found {
		return nil, fmt.Errorf("resource %q not found in program", name)
	}
	p.Resources[name] = resource
	return &oldResource, nil
}

type Resource struct {
	Type            string          `yaml:"type"`
	DefaultProvider bool            `yaml:"defaultProvider,omitempty"`
	Properties      map[string]any  `yaml:"properties"`
	Options         ResourceOptions `yaml:"options,omitempty"`
	Get             ResourceGetter  `yaml:"get,omitempty"`
}

type ResourceOptions struct {
	AdditionalSecretOutputs []string       `yaml:"additionalSecretOutputs,omitempty"`
	Aliases                 []string       `yaml:"aliases,omitempty"`
	CustomTimeouts          CustomTimeouts `yaml:"customTimeouts,omitempty"`
	DeleteBeforeReplace     bool           `yaml:"deleteBeforeReplace,omitempty"`
	DependsOn               []any          `yaml:"dependsOn,omitempty"`
	IgnoreChanges           []string       `yaml:"ignoreChanges,omitempty"`
	Import                  string         `yaml:"import,omitempty"`
	Parent                  any            `yaml:"parent,omitempty"`
	Protect                 bool           `yaml:"protect,omitempty"`
	Provider                any            `yaml:"provider,omitempty"`
	Providers               map[string]any `yaml:"providers,omitempty"`
	Version                 string         `yaml:"version,omitempty"`
	PluginDownloadUrl       string         `yaml:"pluginDownloadUrl,omitempty"`
	ReplaceOnChanges        []string       `yaml:"replaceOnChanges,omitempty"`
	RetainOnDelete          bool           `yaml:"retainOnDelete,omitempty"`
}

type CustomTimeouts struct {
	Create string `yaml:"create,omitempty"`
	Delete string `yaml:"delete,omitempty"`
	Update string `yaml:"update,omitempty"`
}

type ResourceGetter struct {
	Id    string         `yaml:"id"`
	State map[string]any `yaml:"state"`
}

type Plugins struct {
	Providers []Plugin `yaml:"providers"`
	Analyzers []Plugin `yaml:"analyzers"`
	Languages []Plugin `yaml:"languages"`
}

type Plugin struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Version string `yaml:"version,omitempty"`
}
