// Copyright 2016-2024, Pulumi Corporation.
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

package sanitize

import (
	"encoding/json"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

const plaintextSub = "REDACTED BY PROVIDERTEST"
const secretSignature = "4dabf18193072939515e22adb298388d"

// SanitizeSecretsInStackState sanitizes secrets in the stack state by replacing them with a placeholder.
// secrets are identified by their magic signature, copied from pulumi/pulumi.
func SanitizeSecretsInStackState(stack *apitype.UntypedDeployment) (*apitype.UntypedDeployment, error) {
	var d apitype.DeploymentV3
	err := json.Unmarshal(stack.Deployment, &d)
	if err != nil {
		return nil, err
	}

	sanitizeSecretsInResources(d.Resources)

	marshaledDeployment, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	return &apitype.UntypedDeployment{
		Version:    stack.Version,
		Deployment: json.RawMessage(marshaledDeployment),
	}, nil
}

func sanitizeSecretsInResources(resources []apitype.ResourceV3) {
	for i, r := range resources {
		r.Inputs = sanitizeSecretsInObject(r.Inputs)
		r.Outputs = sanitizeSecretsInObject(r.Outputs)
		resources[i] = r
	}
}

var secretReplacement = map[string]any{
	secretSignature: "1b47061264138c4ac30d75fd1eb44270",
	"plaintext":     plaintextSub,
}

func sanitizeSecretsInObject(obj map[string]any) map[string]any {
	copy := map[string]any{}
	for k, v := range obj {
		innerObj, ok := v.(map[string]any)
		if ok {
			_, hasSecret := innerObj[secretSignature]
			if hasSecret {
				copy[k] = secretReplacement
			} else {
				copy[k] = sanitizeSecretsInObject(innerObj)
			}
		} else {
			copy[k] = v
		}
	}
	return copy
}
