// Copyright 2016-2023, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package providertest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// Configures ambient plugin setup for provider tests. Ambient plugins are placed in PATH so that
// pulumi CLI would use them in preference to the plugins specified in the programs under test.
//
// See also PULUMI_IGNORE_AMBIENT_PLUGINS option in pulumi/pulumi.
//
// Manipulating proivder versions in this way is heavy-handed but easy to implement in the framework
// without placing constrainsts on programs under test.
type ambientPlugin struct {
	// Short name of the provider, such as "eks" or "aws".
	Provider string

	// Desired version, such as "5.42.0". Only one of Version, LocalPath should be set.
	Version string

	// Local path to a plugin binary, such as "../bin/pulumi-resource-eks".
	LocalPath string
}

// Builds a PATH environment variable value suitable for setting up the environment with the desired
// ambient plugins. Auto-installed via Pulumi CLI.
func pathWithAmbientPlugins(originalPATH string, plugins ...ambientPlugin) (string, error) {
	_, err := ensurePluginsInstalled(plugins)
	if err != nil {
		return "", err
	}
	newPaths := []string{}
	for _, p := range plugins {
		var newPath string
		if p.LocalPath != "" {
			newPath = filepath.Dir(p.LocalPath)
		} else if p.Version != "" {
			pi := pluginInfo{
				Name:    p.Provider,
				Kind:    "resource",
				Version: p.Version,
			}
			var err error
			newPath, err = pluginPath(pi)
			if err != nil {
				return "", err
			}
		}
		newPaths = append(newPaths, newPath)
	}
	oldPaths := strings.Split(originalPATH, string(os.PathListSeparator))
	allPaths := append(newPaths, oldPaths...)
	return strings.Join(allPaths, string(os.PathListSeparator)), nil
}

type pluginInfo struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Version string `json:"version"`
}

// Find the locally installed Plugin path.
func pluginPath(info pluginInfo) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	p := filepath.Join(usr.HomeDir, ".pulumi", "plugins",
		fmt.Sprintf("%s-%s-%s", info.Kind, info.Name, info.Version))
	return p, nil
}

// Use pulumi plugin install to ensure all requested plugins are installed, if not already.
func ensurePluginsInstalled(plugins []ambientPlugin) ([]pluginInfo, error) {
	matched := []pluginInfo{}
	installed, err := findInstalledPlugins()
	if err != nil {
		return nil, err
	}
	for _, p := range plugins {
		if p.Version != "" {
			isInstalledAlready := false
			for _, i := range installed {
				if i.Kind == "resource" && i.Name == p.Provider && i.Version == p.Version {
					isInstalledAlready = true
					matched = append(matched, i)
					break
				}
			}
			if !isInstalledAlready {
				info := pluginInfo{Kind: "resource", Name: p.Provider, Version: p.Version}
				if err := installPlugin(info); err != nil {
					return nil, err
				}
				installed = append(installed, info)
				matched = append(matched, info)
			}
		}
	}
	return matched, nil
}

// Call pulumi plugin install.
func installPlugin(p pluginInfo) error {
	pulumi, err := exec.LookPath("pulumi")
	if err != nil {
		return fmt.Errorf("cannot find pulumi CLI in PATH: %w", err)
	}
	cmd := exec.Command(pulumi, "plugin", "install", p.Kind, p.Name, p.Version)
	return cmd.Run()
}

// Call pulumi plugin ls to find currently installed plugins.
func findInstalledPlugins() ([]pluginInfo, error) {
	pulumi, err := exec.LookPath("pulumi")
	if err != nil {
		return nil, fmt.Errorf("cannot find pulumi CLI in PATH: %w", err)
	}
	var buf bytes.Buffer
	cmd := exec.Command(pulumi, "plugin", "ls", "--json")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	var plugins []pluginInfo
	if err := json.Unmarshal(buf.Bytes(), &plugins); err != nil {
		return nil, err
	}
	return plugins, nil
}
