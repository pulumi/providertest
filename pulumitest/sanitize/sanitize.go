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

func SanitizeSecretsInGrpcLog(log json.RawMessage) json.RawMessage {
	var data map[string]any
	if err := json.Unmarshal(log, &data); err != nil {
		return log
	}

	sanitized := sanitizeSecretsInObject(data, sanitizeGrpcSecret)
	sanitizedBytes, err := json.Marshal(sanitized)
	if err != nil {
		return log
	}
	return sanitizedBytes
}

func sanitizeGrpcSecret(secretObj map[string]any) map[string]any {
	// gRPC logs contain a field called "value" which is any JSON value
	secretObj["value"] = sanitizeStringsRecursively(secretObj["value"])
	return secretObj
}

func sanitizeSecretsInResources(resources []apitype.ResourceV3) {
	for i, r := range resources {
		r.Inputs = sanitizeSecretsInObject(r.Inputs, sanitizeStateSecret)
		r.Outputs = sanitizeSecretsInObject(r.Outputs, sanitizeStateSecret)
		resources[i] = r
	}
}

func sanitizeStateSecret(secretObj map[string]any) map[string]any {
	// State file secrets either have a plaintext field which is any serialized JSON value
	// of a cyphertext field which is and encrypted version of the JSON value.
	plaintext, hasPlaintext := secretObj["plaintext"]
	if !hasPlaintext {
		return secretObj
	}
	plaintextString, isString := plaintext.(string)
	if !isString {
		return secretObj
	}
	var jsonValue any
	err := json.Unmarshal([]byte(plaintextString), &jsonValue)
	if err != nil {
		return secretObj
	}
	sanitized := sanitizeStringsRecursively(jsonValue)
	sanitizedBytes, err := json.Marshal(sanitized)
	if err != nil {
		return secretObj
	}
	secretObj["plaintext"] = string(sanitizedBytes)
	return secretObj
}

func sanitizeSecretsInObject(obj map[string]any, sanitizeSecret func(map[string]any) map[string]any) map[string]any {
	copy := map[string]any{}
	for k, v := range obj {
		innerObj, ok := v.(map[string]any)
		if ok {
			_, hasSecretSignature := innerObj[secretSignature]
			if hasSecretSignature {
				copy[k] = sanitizeSecret(innerObj)
			} else {
				copy[k] = sanitizeSecretsInObject(innerObj, sanitizeSecret)
			}
		} else {
			copy[k] = v
		}
	}
	return copy
}

func sanitizeStringsRecursively(value any) any {
	switch typedValue := value.(type) {
	case string:
		return plaintextSub
	case []any:
		sanitizedSlice := make([]any, len(typedValue))
		for i, v := range typedValue {
			sanitizedSlice[i] = sanitizeStringsRecursively(v)
		}
		return sanitizedSlice
	case map[string]any:
		sanitizedMap := make(map[string]any, len(typedValue))
		for k, v := range typedValue {
			sanitizedMap[k] = sanitizeStringsRecursively(v)
		}
		return sanitizedMap
	}
	return value
}
