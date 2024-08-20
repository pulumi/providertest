package pulumitest

import (
	"os"
	"path/filepath"

	"github.com/pulumi/providertest/pulumiyaml"
	"gopkg.in/yaml.v3"
)

func (pt *PulumiTest) WriteProject(t PT, program *pulumiyaml.Program) {
	t.Helper()
	yamlBytes, err := yaml.Marshal(program)
	if err != nil {
		ptFatalF(t, "failed to marshal yaml program: %v", err)
	}
	yamlPath := filepath.Join(pt.workingDir, "Pulumi.yaml")
	err = os.WriteFile(yamlPath, yamlBytes, 0644)
	if err != nil {
		ptFatalF(t, "failed to write yaml program: %v", err)
	}
}

func (pt *PulumiTest) ReadProject(t PT) *pulumiyaml.Program {
	t.Helper()
	yamlPath := filepath.Join(pt.workingDir, "Pulumi.yaml")
	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		ptFatalF(t, "failed to read yaml program: %v", err)
	}
	var program pulumiyaml.Program
	err = yaml.Unmarshal(yamlBytes, &program)
	if err != nil {
		ptFatalF(t, "failed to unmarshal yaml program: %v", err)
	}
	return &program
}
