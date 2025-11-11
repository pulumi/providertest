package opttest

// DependencyEdit represents a dependency version edit for a package.
type DependencyEdit struct {
	PackageName string
	Version     string
}

// EditDependency sets a specific version for a dependency in the program under test.
// This is language-dependent and will modify the appropriate dependency file based on the project structure:
// - Node.js: package.json (npm/yarn)
// - Python: requirements.txt or setup.py (pip)
// - Go: go.mod (via go get)
// - .NET: .csproj files (NuGet)
// - YAML: provider version config
//
// Note: Using EditDependency with LocalProviderPath may cause conflicts, as the local provider
// path may override the version from the SDK. A warning will be logged if this combination is detected.
func EditDependency(packageName, version string) Option {
	return optionFunc(func(o *Options) {
		o.DependencyEdits = append(o.DependencyEdits, DependencyEdit{
			PackageName: packageName,
			Version:     version,
		})
	})
}
