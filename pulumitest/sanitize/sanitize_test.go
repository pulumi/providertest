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
	"bytes"
	"encoding/json"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeSecretsInObject(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		input := map[string]any{
			"secondaryAccessKey": map[string]any{
				secretSignature: "1b47061264138c4ac30d75fd1eb44270",
				"plaintext":     "secret",
			},
		}

		expected := map[string]any{
			"secondaryAccessKey": secretReplacement,
		}

		assert.Equal(t, expected, sanitizeSecretsInObject(input))
	})

	t.Run("nested", func(t *testing.T) {
		input := map[string]any{
			"bar": 1,
			"foo": map[string]any{
				"inner": map[string]any{
					"secondaryAccessKey": map[string]any{
						secretSignature: "1b47061264138c4ac30d75fd1eb44270",
						"plaintext":     "secret",
					},
				},
			},
		}

		expected := map[string]any{
			"bar": 1,
			"foo": map[string]any{
				"inner": map[string]any{
					"secondaryAccessKey": secretReplacement,
				},
			},
		}

		assert.Equal(t, expected, sanitizeSecretsInObject(input))
	})
}

func TestSanitizeSecretsInStackState(t *testing.T) {
	t.Parallel()

	var stack apitype.UntypedDeployment
	err := json.Unmarshal(realStack, &stack)
	require.NoError(t, err)

	sanitized, err := SanitizeSecretsInStackState(&stack)
	require.NoError(t, err)

	expected := bytes.ReplaceAll(realStack, []byte(`\"SECRET\"`), []byte(`\"`+plaintextSub+`\"`))
	var expectedDeployment apitype.UntypedDeployment
	err = json.Unmarshal(expected, &expectedDeployment)
	require.NoError(t, err)

	sanitizedPretty := prettyPrintJson(t, sanitized.Deployment)
	expectedPretty := prettyPrintJson(t, expectedDeployment.Deployment)
	assert.JSONEq(t, string(expectedPretty), string(sanitizedPretty))
}

func prettyPrintJson(t *testing.T, jsonStr []byte) []byte {
	var v any
	err := json.Unmarshal(jsonStr, &v)
	require.NoError(t, err)
	pretty, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err)
	return pretty
}

var realStack = []byte(`{
  "version": 3,
  "deployment": {
    "manifest": {
      "time": "2024-09-05T11:23:42.551264+02:00",
      "magic": "59ab42470ec682a2eb8566128a64ecaee8e5d25c6d5902576977eb325cf4d7b3",
      "version": "v3.130.0"
    },
    "secrets_providers": {
      "type": "passphrase",
      "state": {
        "salt": "v1:YMv/Yx+VlW0=:v1:sbgZHJ6QDAq8dzEQ:gbsJTyFyS7GU0svVisIL+uQyDJYqqA=="
      }
    },
    "resources": [
      {
        "urn": "urn:pulumi:test::storage::pulumi:pulumi:Stack::storage-test",
        "custom": false,
        "type": "pulumi:pulumi:Stack",
        "created": "2024-09-05T09:22:04.581633Z",
        "modified": "2024-09-05T09:22:04.581633Z"
      },
      {
        "urn": "urn:pulumi:test::storage::pulumi:providers:azure::default",
        "custom": true,
        "id": "515481f4-90eb-46e4-a36e-29ad4413fb22",
        "type": "pulumi:providers:azure",
        "inputs": {
          "subscriptionId": {
            "4dabf18193072939515e22adb298388d": "1b47061264138c4ac30d75fd1eb44270",
            "plaintext": "\"SECRET\""
          }
        },
        "outputs": {
          "subscriptionId": {
            "4dabf18193072939515e22adb298388d": "1b47061264138c4ac30d75fd1eb44270",
            "plaintext": "\"SECRET\""
          }
        },
        "created": "2024-09-05T09:22:05.050659Z",
        "modified": "2024-09-05T09:22:05.050659Z"
      },
      {
        "urn": "urn:pulumi:test::storage::azure:core/resourceGroup:ResourceGroup::exampleResourceGroup",
        "custom": true,
        "id": "/subscriptions/12345/resourceGroups/exampleresourcegroup35548da3",
        "type": "azure:core/resourceGroup:ResourceGroup",
        "inputs": {
          "__defaults": [
            "name"
          ],
          "location": "East US",
          "name": "exampleresourcegroup35548da3"
        },
        "outputs": {
          "__meta": "{\"e2bfb730-ecaa-11e6-8f88-34363bc7c4c0\":{\"create\":5400000000000,\"delete\":5400000000000,\"read\":300000000000,\"update\":5400000000000}}",
          "id": "/subscriptions/12345/resourceGroups/exampleresourcegroup35548da3",
          "location": "eastus",
          "managedBy": "",
          "name": "exampleresourcegroup35548da3",
          "tags": null
        },
        "parent": "urn:pulumi:test::storage::pulumi:pulumi:Stack::storage-test",
        "provider": "urn:pulumi:test::storage::pulumi:providers:azure::default::515481f4-90eb-46e4-a36e-29ad4413fb22",
        "propertyDependencies": {
          "location": []
        },
        "created": "2024-09-05T09:22:23.554439Z",
        "modified": "2024-09-05T09:22:23.554439Z"
      },
      {
        "urn": "urn:pulumi:test::storage::azure:storage/account:Account::exampleAccount",
        "custom": true,
        "id": "/subscriptions/12345/resourceGroups/exampleresourcegroup35548da3/providers/Microsoft.Storage/storageAccounts/exampleaccount4cb2982b",
        "type": "azure:storage/account:Account",
        "inputs": {
          "__defaults": [
            "accountKind",
            "allowNestedItemsToBePublic",
            "crossTenantReplicationEnabled",
            "defaultToOauthAuthentication",
            "dnsEndpointType",
            "infrastructureEncryptionEnabled",
            "isHnsEnabled",
            "localUserEnabled",
            "minTlsVersion",
            "name",
            "nfsv3Enabled",
            "publicNetworkAccessEnabled",
            "queueEncryptionKeyType",
            "sftpEnabled",
            "sharedAccessKeyEnabled",
            "tableEncryptionKeyType"
          ],
          "accountKind": "StorageV2",
          "accountReplicationType": "LRS",
          "accountTier": "Standard",
          "allowNestedItemsToBePublic": true,
          "crossTenantReplicationEnabled": true,
          "defaultToOauthAuthentication": false,
          "dnsEndpointType": "Standard",
          "infrastructureEncryptionEnabled": false,
          "isHnsEnabled": false,
          "localUserEnabled": true,
          "location": "eastus",
          "minTlsVersion": "TLS1_2",
          "name": "exampleaccount4cb2982b",
          "nfsv3Enabled": false,
          "publicNetworkAccessEnabled": true,
          "queueEncryptionKeyType": "Service",
          "resourceGroupName": "exampleresourcegroup35548da3",
          "sftpEnabled": false,
          "sharedAccessKeyEnabled": true,
          "tableEncryptionKeyType": "Service",
          "tags": {
            "environment": "staging"
          }
        },
        "outputs": {
          "__meta": "{\"e2bfb730-ecaa-11e6-8f88-34363bc7c4c0\":{\"create\":3600000000000,\"delete\":3600000000000,\"read\":300000000000,\"update\":3600000000000},\"schema_version\":\"4\"}",
          "accessTier": "Hot",
          "accountKind": "StorageV2",
          "accountReplicationType": "LRS",
          "accountTier": "Standard",
          "allowNestedItemsToBePublic": true,
          "allowedCopyScope": "",
          "azureFilesAuthentication": null,
          "blobProperties": {
            "changeFeedEnabled": false,
            "changeFeedRetentionInDays": 0,
            "containerDeleteRetentionPolicy": null,
            "corsRules": [],
            "defaultServiceVersion": "",
            "deleteRetentionPolicy": null,
            "lastAccessTimeEnabled": false,
            "restorePolicy": null,
            "versioningEnabled": false
          },
          "crossTenantReplicationEnabled": true,
          "customDomain": null,
          "customerManagedKey": null,
          "defaultToOauthAuthentication": false,
          "dnsEndpointType": "Standard",
          "edgeZone": "",
          "enableHttpsTrafficOnly": true,
          "httpsTrafficOnlyEnabled": true,
          "id": "/subscriptions/12345/resourceGroups/exampleresourcegroup35548da3/providers/Microsoft.Storage/storageAccounts/exampleaccount4cb2982b",
          "identity": null,
          "immutabilityPolicy": null,
          "infrastructureEncryptionEnabled": false,
          "isHnsEnabled": false,
          "largeFileShareEnabled": false,
          "localUserEnabled": true,
          "location": "eastus",
          "minTlsVersion": "TLS1_2",
          "name": "exampleaccount4cb2982b",
          "networkRules": null,
          "nfsv3Enabled": false,
          "primaryAccessKey": {
            "4dabf18193072939515e22adb298388d": "1b47061264138c4ac30d75fd1eb44270",
            "plaintext": "\"SECRET\""
          },
          "primaryBlobConnectionString": {
            "4dabf18193072939515e22adb298388d": "1b47061264138c4ac30d75fd1eb44270",
            "plaintext": "\"SECRET\""
          },
          "primaryBlobEndpoint": "https://exampleaccount4cb2982b.blob.core.windows.net/",
          "primaryBlobHost": "exampleaccount4cb2982b.blob.core.windows.net",
          "primaryBlobInternetEndpoint": "",
          "primaryBlobInternetHost": "",
          "primaryBlobMicrosoftEndpoint": "",
          "primaryBlobMicrosoftHost": "",
          "primaryConnectionString": {
            "4dabf18193072939515e22adb298388d": "1b47061264138c4ac30d75fd1eb44270",
            "plaintext": "\"SECRET\""
          },
          "primaryDfsEndpoint": "https://exampleaccount4cb2982b.dfs.core.windows.net/",
          "primaryDfsHost": "exampleaccount4cb2982b.dfs.core.windows.net",
          "primaryDfsInternetEndpoint": "",
          "primaryDfsInternetHost": "",
          "primaryDfsMicrosoftEndpoint": "",
          "primaryDfsMicrosoftHost": "",
          "primaryFileEndpoint": "https://exampleaccount4cb2982b.file.core.windows.net/",
          "primaryFileHost": "exampleaccount4cb2982b.file.core.windows.net",
          "primaryFileInternetEndpoint": "",
          "primaryFileInternetHost": "",
          "primaryFileMicrosoftEndpoint": "",
          "primaryFileMicrosoftHost": "",
          "primaryLocation": "eastus",
          "primaryQueueEndpoint": "https://exampleaccount4cb2982b.queue.core.windows.net/",
          "primaryQueueHost": "exampleaccount4cb2982b.queue.core.windows.net",
          "primaryQueueMicrosoftEndpoint": "",
          "primaryQueueMicrosoftHost": "",
          "primaryTableEndpoint": "https://exampleaccount4cb2982b.table.core.windows.net/",
          "primaryTableHost": "exampleaccount4cb2982b.table.core.windows.net",
          "primaryTableMicrosoftEndpoint": "",
          "primaryTableMicrosoftHost": "",
          "primaryWebEndpoint": "https://exampleaccount4cb2982b.z13.web.core.windows.net/",
          "primaryWebHost": "exampleaccount4cb2982b.z13.web.core.windows.net",
          "primaryWebInternetEndpoint": "",
          "primaryWebInternetHost": "",
          "primaryWebMicrosoftEndpoint": "",
          "primaryWebMicrosoftHost": "",
          "publicNetworkAccessEnabled": true,
          "queueEncryptionKeyType": "Service",
          "queueProperties": {
            "corsRules": [],
            "hourMetrics": {
              "enabled": true,
              "includeApis": true,
              "retentionPolicyDays": 7,
              "version": "1.0"
            },
            "logging": {
              "delete": false,
              "read": false,
              "retentionPolicyDays": 0,
              "version": "1.0",
              "write": false
            },
            "minuteMetrics": {
              "enabled": false,
              "includeApis": false,
              "retentionPolicyDays": 0,
              "version": "1.0"
            }
          },
          "resourceGroupName": "exampleresourcegroup35548da3",
          "routing": null,
          "sasPolicy": null,
          "secondaryAccessKey": {
            "4dabf18193072939515e22adb298388d": "1b47061264138c4ac30d75fd1eb44270",
            "plaintext": "\"SECRET\""
          },
          "secondaryBlobConnectionString": {
            "4dabf18193072939515e22adb298388d": "1b47061264138c4ac30d75fd1eb44270",
            "plaintext": "\"SECRET\""
          },
          "secondaryBlobEndpoint": "",
          "secondaryBlobHost": "",
          "secondaryBlobInternetEndpoint": "",
          "secondaryBlobInternetHost": "",
          "secondaryBlobMicrosoftEndpoint": "",
          "secondaryBlobMicrosoftHost": "",
          "secondaryConnectionString": {
            "4dabf18193072939515e22adb298388d": "1b47061264138c4ac30d75fd1eb44270",
            "plaintext": "\"SECRET\""
          },
          "secondaryDfsEndpoint": "",
          "secondaryDfsHost": "",
          "secondaryDfsInternetEndpoint": "",
          "secondaryDfsInternetHost": "",
          "secondaryDfsMicrosoftEndpoint": "",
          "secondaryDfsMicrosoftHost": "",
          "secondaryFileEndpoint": "",
          "secondaryFileHost": "",
          "secondaryFileInternetEndpoint": "",
          "secondaryFileInternetHost": "",
          "secondaryFileMicrosoftEndpoint": "",
          "secondaryFileMicrosoftHost": "",
          "secondaryLocation": "",
          "secondaryQueueEndpoint": "",
          "secondaryQueueHost": "",
          "secondaryQueueMicrosoftEndpoint": "",
          "secondaryQueueMicrosoftHost": "",
          "secondaryTableEndpoint": "",
          "secondaryTableHost": "",
          "secondaryTableMicrosoftEndpoint": "",
          "secondaryTableMicrosoftHost": "",
          "secondaryWebEndpoint": "",
          "secondaryWebHost": "",
          "secondaryWebInternetEndpoint": "",
          "secondaryWebInternetHost": "",
          "secondaryWebMicrosoftEndpoint": "",
          "secondaryWebMicrosoftHost": "",
          "sftpEnabled": false,
          "shareProperties": {
            "corsRules": [],
            "retentionPolicy": {
              "days": 7
            },
            "smb": null
          },
          "sharedAccessKeyEnabled": true,
          "staticWebsite": null,
          "tableEncryptionKeyType": "Service",
          "tags": {
            "environment": "staging"
          }
        },
        "parent": "urn:pulumi:test::storage::pulumi:pulumi:Stack::storage-test",
        "dependencies": [
          "urn:pulumi:test::storage::azure:core/resourceGroup:ResourceGroup::exampleResourceGroup"
        ],
        "provider": "urn:pulumi:test::storage::pulumi:providers:azure::default::515481f4-90eb-46e4-a36e-29ad4413fb22",
        "propertyDependencies": {
          "accountReplicationType": [],
          "accountTier": [],
          "location": [
            "urn:pulumi:test::storage::azure:core/resourceGroup:ResourceGroup::exampleResourceGroup"
          ],
          "resourceGroupName": [
            "urn:pulumi:test::storage::azure:core/resourceGroup:ResourceGroup::exampleResourceGroup"
          ],
          "tags": []
        },
        "additionalSecretOutputs": [
          "primaryAccessKey",
          "primaryBlobConnectionString",
          "primaryConnectionString",
          "secondaryAccessKey",
          "secondaryBlobConnectionString",
          "secondaryConnectionString"
        ],
        "created": "2024-09-05T09:23:38.680385Z",
        "modified": "2024-09-05T09:23:38.680385Z"
      }
   ]
  }
}`)
