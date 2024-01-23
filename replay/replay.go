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

package replay

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	jsonpb "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

// Replay executes a request from a provider operation log against an in-memory resource provider server and asserts
// that the server's response matches the logged response.
//
// The jsonLog parameter is a verbatim JSON string such as this one:
//
//	{
//	  "method": "/pulumirpc.ResourceProvider/Create",
//	  "request": {
//	    "urn": "urn:pulumi:dev::repro-pulumi-random::random:index/randomString:RandomString::s",
//	    "properties": {
//	      "length": 1
//	    }
//	  },
//	  "response": {
//	    "id": "*",
//	    "properties": {
//	      "__meta": "{\"schema_version\":\"2\"}",
//	      "id": "*",
//	      "result": "*",
//	      "length": 1,
//	      "lower": true,
//	      "minLower": 0,
//	      "minNumeric": 0,
//	      "minSpecial": 0,
//	      "minUpper": 0,
//	      "number": true,
//	      "numeric": true,
//	      "special": true,
//	      "upper": true
//	    }
//	  }
//	}
//
// The format is the JSON encoding of the gRPC protocol used by Pulumi ResourceProvider service.
//
//	https://github.com/pulumi/pulumi/blob/master/proto/pulumi/provider.proto#L27
//
// Conveniently, the format matches what Pulumi CLI emits when invoked with PULUMI_DEBUG_GPRC:
//
//	PULUMI_DEBUG_GPRC=$PWD/log.json pulumi up
//
// This allows quickly turning fragments of the program execution trace into test cases.
//
// Instead of direct JSON equality, Replay uses AssertJSONMatchesPattern to compare the actual and expected responses.
// This allows patterns such as "*". In the above example, the random provider will generate new strings with every
// invocation and they would fail a strict equality check. Using "*" allows the test to succeed while ignoring the
// randomness.
//
// Beware possible side-effects: although Replay executes in-memory without actual gRPC sockets, replaying against an
// actual resource provider will side-effect. For example, replaying Create calls against pulumi-aws provider may try to
// create resorces in AWS. This is not an issue with side-effect-free providers such as pulumi-random, or for methods
// that do not involve cloud interaction such as Diff.
//
// Replay does not assume that the provider is a bridged provider and can be generally useful.
func Replay(t *testing.T, server pulumirpc.ResourceProviderServer, jsonLog string) {
	ctx := context.Background()
	var entry jsonLogEntry
	err := json.Unmarshal([]byte(jsonLog), &entry)
	assert.NoError(t, err)

	switch entry.Method {

	case "/pulumirpc.ResourceProvider/GetSchema":
		replay(t, entry, new(pulumirpc.GetSchemaRequest), server.GetSchema, nil)

	case "/pulumirpc.ResourceProvider/CheckConfig":
		replay(t, entry, new(pulumirpc.CheckRequest), server.CheckConfig, normCheckResponse)

	case "/pulumirpc.ResourceProvider/DiffConfig":
		replay(t, entry, new(pulumirpc.DiffRequest), server.DiffConfig, nil)

	case "/pulumirpc.ResourceProvider/Configure":
		replay(t, entry, new(pulumirpc.ConfigureRequest), server.Configure, nil)

	case "/pulumirpc.ResourceProvider/Invoke":
		replay(t, entry, new(pulumirpc.InvokeRequest), server.Invoke, normInvokeResponse)

	// TODO StreamInvoke might need some special handling as it is a streaming RPC method.

	case "/pulumirpc.ResourceProvider/Call":
		replay(t, entry, new(pulumirpc.CallRequest), server.Call, normCallResponse)

	case "/pulumirpc.ResourceProvider/Check":
		replay(t, entry, new(pulumirpc.CheckRequest), server.Check, normCheckResponse)

	case "/pulumirpc.ResourceProvider/Diff":
		replay(t, entry, new(pulumirpc.DiffRequest), server.Diff, nil)

	case "/pulumirpc.ResourceProvider/Create":
		replay(t, entry, new(pulumirpc.CreateRequest), server.Create, nil)

	case "/pulumirpc.ResourceProvider/Read":
		replay(t, entry, new(pulumirpc.ReadRequest), server.Read, nil)

	case "/pulumirpc.ResourceProvider/Update":
		replay(t, entry, new(pulumirpc.UpdateRequest), server.Update, nil)

	case "/pulumirpc.ResourceProvider/Delete":
		replay(t, entry, new(pulumirpc.DeleteRequest), server.Delete, nil)

	case "/pulumirpc.ResourceProvider/Construct":
		replay(t, entry, new(pulumirpc.ConstructRequest), server.Construct, nil)

	case "/pulumirpc.ResourceProvider/Cancel":
		_, err := server.Cancel(ctx, &emptypb.Empty{})
		assert.NoError(t, err)

	// TODO GetPluginInfo is a bit odd in that it has an Empty request, need to generealize replay() function.
	//
	// rpc GetPluginInfo(google.protobuf.Empty) returns (PluginInfo) {}

	case "/pulumirpc.ResourceProvider/Attach":
		replay(t, entry, new(pulumirpc.PluginAttach), server.Attach, nil)

	case "/pulumirpc.ResourceProvider/GetMapping":
		replay(t, entry, new(pulumirpc.GetMappingRequest), server.GetMapping, nil)

	case "/pulumirpc.ResourceProvider/GetMappings":
		replay(t, entry, new(pulumirpc.GetMappingsRequest), server.GetMappings, nil)

	default:
		t.Errorf("Unknown method: %s", entry.Method)
	}
}

// ReplaySequence is exactly like Replay, but expects jsonLog to encode a sequence of events `[e1, e2, e3]`, and will
// call Replay on each of those events in the given order.
func ReplaySequence(t *testing.T, server pulumirpc.ResourceProviderServer, jsonLog string) {
	var entries []jsonLogEntry
	err := json.Unmarshal([]byte(jsonLog), &entries)
	assert.NoError(t, err)
	for _, e := range entries {
		bytes, err := json.Marshal(e)
		assert.NoError(t, err)
		Replay(t, server, string(bytes))
	}
}

func replay[Req protoreflect.ProtoMessage, Resp protoreflect.ProtoMessage](
	t *testing.T,
	entry jsonLogEntry,
	req Req,
	serve func(context.Context, Req) (Resp, error),
	normalizeResponse func(Resp),
) {
	ctx := context.Background()

	err := jsonpb.Unmarshal([]byte(entry.Request), req)
	assert.NoError(t, err)

	resp, err := serve(ctx, req)
	if done := assertErrorMatchesSpec(t, entry.Errors, err); done {
		return
	}

	if normalizeResponse != nil {
		normalizeResponse(resp)
	}

	bytes, err := jsonpb.Marshal(resp)
	assert.NoError(t, err)

	var expected, actual json.RawMessage = entry.Response, bytes
	AssertJSONMatchesPattern(t, expected, actual)
}

func assertErrorMatchesSpec(t *testing.T, expectedErrors []string, err error) bool {
	switch {
	case len(expectedErrors) == 0:
		require.NoError(t, err)
		return false
	case len(expectedErrors) == 1:
		require.Error(t, err)
		e := expectedErrors[0]
		if e != "*" {
			require.Equal(t, e, err.Error())
		}
		return true
	default:
		// The cases where there are actual multiple errors returned from gRPC intereceptor
		// seem a little unclear, the code is in interceptors.go:
		//
		// https://github.com/pulumi/pulumi/blob/master/pkg/util/rpcdebug/interceptors.go#L34
		//
		// It seems we have at most one logical error coming from the provider, but we also
		// may record errors coming from interceptor-related issues in the error list.
		//
		// If this reasoning is correct, when reusing GRPC logs as expectations, a
		// reasonable default assert is to check that the actual error message is contained
		// in the list of expectations.
		require.Error(t, err)
		require.Contains(t, expectedErrors, err.Error())
		return true
	}
}

// ReplayFile executes ReplaySequence on all pulumirpc.ResourceProvider events found in the file produced with
// PULUMI_DEBUG_GPRC. For example:
//
//	PULUMI_DEBUG_GPRC=testdata/log.json pulumi up
//
// This produces the testdata/log.json file, which can then be used for Replay-style testing:
//
//	ReplayFile(t, server, "testdata/log.json")
func ReplayFile(t *testing.T, server pulumirpc.ResourceProviderServer, traceFile string) {
	bytes, err := os.ReadFile(traceFile)
	require.NoError(t, err)

	var entries []jsonLogEntry
	err = json.Unmarshal(bytes, &entries)
	require.NoError(t, err)

	count := 0
	for _, entry := range entries {
		if entry.Method == "" {
			continue
		}

		if !strings.HasPrefix(entry.Method, "/pulumirpc.ResourceProvider") {
			continue
		}
		// TODO support replaying all these method calls.
		switch entry.Method {
		case "/pulumirpc.ResourceProvider/StreamInvoke":
			continue
		case "/pulumirpc.ResourceProvider/GetPluginInfo":
			continue
		default:
			entryBytes, err := json.Marshal(entry)
			require.NoError(t, err)
			Replay(t, server, string(entryBytes))
			count++
		}
	}
	assert.Greater(t, count, 0)
}

// See also: https://github.com/pulumi/pulumi/blob/master/pkg/util/rpcdebug/logformat.go#L28
type jsonLogEntry struct {
	Method   string          `json:"method"`
	Request  json.RawMessage `json:"request,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
	Errors   []string        `json:"errors,omitempty"`
}

func normInvokeResponse(resp *pulumirpc.InvokeResponse) {
	if resp == nil {
		return
	}
	sortCheckFailures(resp.Failures)
}

func normCallResponse(resp *pulumirpc.CallResponse) {
	if resp == nil {
		return
	}
	sortCheckFailures(resp.Failures)
}

func normCheckResponse(resp *pulumirpc.CheckResponse) {
	if resp == nil {
		return
	}
	sortCheckFailures(resp.Failures)
}

func sortCheckFailures(cf []*pulumirpc.CheckFailure) {
	sort.SliceStable(cf, func(i, j int) bool {
		a, b := cf[i], cf[j]
		if a.Property < b.Property {
			return true
		}
		if a.Reason < b.Reason {
			return true
		}
		return false
	})
}
