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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	testutils "github.com/pulumi/pulumi-terraform-bridge/testing/x"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"gopkg.in/yaml.v3"
)

type UpgradeTestMode int

const (
	UpgradeTestMode_Quick UpgradeTestMode = iota
	UpgradeTestMode_PreviewOnly
)

func (m UpgradeTestMode) String() string {
	switch m {
	case UpgradeTestMode_PreviewOnly:
		return "PreviewOnly"
	case UpgradeTestMode_Quick:
		return "Quick"
	}
	return "<unknown>"
}

func VerifyUpgrade(t *testing.T) *providerUpgradeBuilder {
	return &providerUpgradeBuilder{
		tt: t,
		modes: map[UpgradeTestMode]string{
			UpgradeTestMode_Quick:       "",
			UpgradeTestMode_PreviewOnly: "",
		},
		config: map[string]string{},
	}
}

type providerUpgradeBuilder struct {
	tt                     *testing.T
	resourceProviderServer pulumirpc.ResourceProviderServer
	name                   string
	program                string
	modes                  map[UpgradeTestMode]string // skip reason by mode
	baselineVersion        string
	config                 map[string]string
}

func (b *providerUpgradeBuilder) Skip(
	m UpgradeTestMode,
	reason string,
) *providerUpgradeBuilder {
	require.NotEmpty(b.tt, reason, "reason cannot be empty")
	b.modes[m] = reason
	return b
}

func (b *providerUpgradeBuilder) WithBaselineVersion(v string) *providerUpgradeBuilder {
	b.baselineVersion = v
	return b
}

func (b *providerUpgradeBuilder) WithConfig(key, value string) *providerUpgradeBuilder {
	b.config[key] = value
	return b
}

func (b *providerUpgradeBuilder) WithProgram(dir string) *providerUpgradeBuilder {
	require.NotEmptyf(b.tt, dir, "dir cannot be empty")
	require.Truef(b.tt, dirExists(b.tt, dir), "no such directory")
	b.program = dir
	return b
}

func (b *providerUpgradeBuilder) WithProviderName(name string) *providerUpgradeBuilder {
	require.NotEmptyf(b.tt, name, "name cannot be empty, "+
		"expecting a provider name like `gcp` or `aws`")
	b.name = name
	return b
}

func (b *providerUpgradeBuilder) WithResourceProviderServer(
	s pulumirpc.ResourceProviderServer,
) *providerUpgradeBuilder {
	require.NotNil(b.tt, s)
	b.resourceProviderServer = s
	return b
}

func (b *providerUpgradeBuilder) Run() {
	require.NotEmptyf(b.tt, b.program, "WithProgram call is required")
	b.verifyVersion()

	acceptEnvVar := "PULUMI_ACCEPT"
	accept := cmdutil.IsTruthy(os.Getenv(acceptEnvVar))
	if accept {
		b.tt.Logf("Recording baseline behavior as requested by "+
			"setting %q env var", acceptEnvVar)
		b.providerUpgradeRecordBaselines(b.tt)
	}
	b.tt.Run("Quick", func(t *testing.T) {
		if skip := b.modes[UpgradeTestMode_Quick]; skip != "" {
			t.Skip(skip)
		}
		b.checkProviderUpgradeQuick(t)
	})
	b.tt.Run("PreviewOnly", func(t *testing.T) {
		if skip := b.modes[UpgradeTestMode_PreviewOnly]; skip != "" {
			t.Skip(skip)
		}
		if testing.Short() {
			t.Skipf("Skipping in -short mode")
			return
		}

		if accept {
			t.Skipf("Skipping because baselines were just pre-recorded")
		} else {
			b.checkProviderUpgradePreviewOnly(t)
		}
	})
}

func (b *providerUpgradeBuilder) checkProviderUpgradeQuick(t *testing.T) {
	require.NotNilf(b.tt, b.resourceProviderServer, "WithResourceProviderServer is required")
	info := b.newProviderUpgradeInfo(t)

	bytes, err := os.ReadFile(info.grpcFile)
	require.NoErrorf(t, err,
		"No pre-recorded gRPC log found, try to run with PULUMI_ACCEPT=1 to record")

	eng := &mockPulumiEngine{
		provider:              b.resourceProviderServer,
		lastCheckRequestByURN: map[string]*pulumirpc.CheckRequest{},
	}

	for _, line := range strings.Split(string(bytes), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = ignoreStables(t, line)
		eng.replayGRPCLog(t, line)
	}
	require.NotEmptyf(t, eng.verifiedDiffResourceCounter, "Need at least one replay test")
}

// Verifies provider upgrades by replaying Diff calls. This is slighly involved. The available
// information is Check and Diff calls recorded on vPrev version of the provider, and a vNext
// in-memory version of the provider available to test. The calls cannot be replayed directly,
// instead Check and Diff calls are paired to do something equivalent to this:
//
//	rawInputs := vPrev.Check.inputs
//	diffNew := vNext.Diff(vPrev.State, vNext.Check(rawInputs))
//	diffOld := vPrev.Diff(vPrev.State, vPrev.Check(rawInputs))
//	assert.Equal(t, diffOld, diffNew)
//
// Essentially the pre-recorded Check calls are used to extract a gRPC representation of raw
// resource inputs coming from the user program. This could have been parsed from YAML programs but
// would require interpolating variables and converting to gRPC-compatible form, parsing Check is
// easier.
//
// Then it is asserted that the vNext version of Diff behaves consistently with the vPrev.Diff on
// old state and inputs. This simulates the scenario of updating the provider while not making any
// changes to the program.
type mockPulumiEngine struct {
	// vNext in-memory provider
	provider                    pulumirpc.ResourceProviderServer
	lastCheckRequestByURN       map[string]*pulumirpc.CheckRequest
	verifiedDiffResourceCounter int
}

func (e *mockPulumiEngine) replayGRPCLog(t *testing.T, jsonLog string) {
	var entry jsonLogEntry
	err := json.Unmarshal([]byte(jsonLog), &entry)
	require.NoError(t, err)

	switch entry.Method {
	case "/pulumirpc.ResourceProvider/Check":
		req := unmarshalProto(t, entry.Request, new(pulumirpc.CheckRequest))
		e.recordCheck(t, req)
	case "/pulumirpc.ResourceProvider/Diff":
		req := unmarshalProto(t, entry.Request, new(pulumirpc.DiffRequest))
		e.fixupDiff(t, req)
		entry.Request = marshalProto(t, req)
		b, err := json.Marshal(entry)
		require.NoError(t, err)
		testutils.Replay(t, e.provider, string(b))
		e.verifiedDiffResourceCounter++
		t.Logf("Replayed Diff on %v", req.Urn)
	}
}

func (e *mockPulumiEngine) recordCheck(t *testing.T, checkReq *pulumirpc.CheckRequest) {
	e.lastCheckRequestByURN[checkReq.Urn] = checkReq
}

func (e *mockPulumiEngine) fixupDiff(t *testing.T, diffReq *pulumirpc.DiffRequest) {
	ctx := context.Background()
	lastCheck, ok := e.lastCheckRequestByURN[diffReq.Urn]
	require.Truef(t, ok, "Diff called for %q but there is no recent Check for this URN",
		diffReq.Urn)

	// Assuming here that CheckRequest does not depend on the provider version, so that
	// replaying a pre-recorded Check request from old provider on the new RC provider is
	// reasonable.
	checkResp, err := e.provider.Check(ctx, lastCheck)
	require.NoError(t, err)

	// Emulate the real engine would be passing checked inputs into the News field of the
	// DiffRequest and then replay this updated request against the provider.
	diffReq.News = checkResp.GetInputs()
}

type jsonLogEntry struct {
	Method   string          `json:"method"`
	Request  json.RawMessage `json:"request,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
}

func unmarshalProto[T protoreflect.ProtoMessage](t *testing.T, data json.RawMessage, req T) T {
	err := jsonpb.Unmarshal([]byte(data), req)
	require.NoError(t, err)
	return req
}

func marshalProto[T protoreflect.ProtoMessage](t *testing.T, req T) json.RawMessage {
	bytes, err := jsonpb.Marshal(req)
	require.NoError(t, err)
	return bytes
}

func ignoreStables(t *testing.T, grpcLogEntry string) string {
	var v map[string]any
	err := json.Unmarshal([]byte(grpcLogEntry), &v)
	require.NoError(t, err)
	if r, ok := v["response"]; ok {
		r := r.(map[string]any)
		if _, ok := r["stables"]; ok {
			r["stables"] = "*"
		}
	}
	out, err := json.Marshal(v)
	require.NoError(t, err)
	return string(out)
}

type providerUpgradeInfo struct {
	recordingDir string
	grpcFile     string
	stateFile    string
}

func (b *providerUpgradeBuilder) newProviderUpgradeInfo(t *testing.T) providerUpgradeInfo {
	info := providerUpgradeInfo{}
	n := strings.ReplaceAll(t.Name(), "Quick", ".")
	info.recordingDir = filepath.Join("testdata", "recorded", n, b.baselineVersion)
	var err error
	info.grpcFile, err = filepath.Abs(filepath.Join(info.recordingDir, "grpc.json"))
	require.NoError(t, err)
	info.stateFile, err = filepath.Abs(filepath.Join(info.recordingDir, "state.json"))
	require.NoError(t, err)
	return info
}

func (b *providerUpgradeBuilder) checkProviderUpgradePreviewOnly(t *testing.T) {
	info := b.newProviderUpgradeInfo(t)
	t.Logf("Baseline provider version: %s", b.baselineVersion)

	opts := integration.ProgramTestOptions{
		Dir:    b.program,
		Env:    []string{},
		Config: b.config,

		// Skips are required by programTestHelper.previewOnlyUpgradeTest
		SkipUpdate:       true,
		SkipRefresh:      true,
		SkipExportImport: true,
	}

	ambientProvider, _ := exec.LookPath(b.providerBinary())
	require.NotEmptyf(t, ambientProvider, "expected to find a release candidate provider "+
		"binary in PATH, try to call `make provider` and `export PATH=$PWD/bin:$PATH`")

	pth := newProgramTestHelper(t, opts)
	t.Logf("%v", pth)
	err := pth.previewOnlyUpgradeTest(info.stateFile)
	require.NoError(t, err)
}

func (b *providerUpgradeBuilder) providerBinary() string {
	return fmt.Sprintf("pulumi-resource-%s", b.name)
}

// Preview-only integration test.

type programTestHelper struct {
	t         *testing.T
	opts      integration.ProgramTestOptions
	pt        *integration.ProgramTester
	stackName string
}

func newProgramTestHelper(t *testing.T, opts integration.ProgramTestOptions) *programTestHelper {
	require.Falsef(t, opts.RunUpdateTest, "RunUpdateTest is not supported")
	require.Emptyf(t, opts.StackName, "Custom StackName is not supported")
	// Allocate stack name.
	stackName := opts.GetStackName()
	require.NotEmptyf(t, opts.StackName,
		"Expected GetStackName() to allocate a random stack name")
	pt := integration.ProgramTestManualLifeCycle(t, &opts)
	return &programTestHelper{
		t:         t,
		opts:      opts,
		pt:        pt,
		stackName: string(stackName),
	}
}

func (pth *programTestHelper) previewOnlyUpgradeTest(stateFile string) error {
	t := pth.t
	pt := pth.pt
	opts := pth.opts
	return pth.lifecycleInitAndDestroy(func() error {
		t.Logf("Importing pre-recorded stateFile from the baseline provider version")
		fixedStateFile := pth.fixupStackName(stateFile)
		if err := pt.RunPulumiCommand("stack", "import",
			"--file", fixedStateFile); err != nil {
			return err
		}

		t.Logf("Running preview using the new provider version")
		// Only run preview. There is no dedicated API for that so instead we check that
		// flags disable everything else. This runs preview twice unfortunately, it's the
		// second one that needs to run. The second preview is gated by
		// SkipEmptyPreviewUpdate and is checking that there are no unexpected updates.
		//
		// If this code could run just pt.PreviewAndUpdate that would be better but it needs
		// to access pt.dir which is kept private.
		require.Falsef(t, opts.SkipPreview,
			"previewOnlyUpgradeTest is incompatible with SkipPreview")
		require.True(t, opts.SkipUpdate, "expecting SkipUpdate: true")
		require.True(t, opts.SkipRefresh, "expecting SkipRefresh: true")
		require.True(t, opts.SkipExportImport, "expecting SkipExportImport: true")
		require.Falsef(t, opts.SkipEmptyPreviewUpdate,
			"expecting SkipEmptyPreviewUpdate: false")
		require.Emptyf(t, opts.EditDirs,
			"previewOnlyUpgradeTest is incompatible with EditDirs")
		if err := pt.TestPreviewUpdateAndEdits(); err != nil {
			return fmt.Errorf("running test preview: %w", err)
		}
		return nil
	})
}

func (pth *programTestHelper) fixupStackName(stateFile string) string {
	t := pth.t
	stackName := pth.stackName
	tempDir := t.TempDir()
	state := readFile(t, stateFile)
	//t.Logf("prior state: %v", state)
	fixedState := pth.withUpdatedStackName(stackName, state)
	fixedStateFile := filepath.Join(tempDir, "fixed-state.json")
	//t.Logf("fixed state: %v", fixedState)
	writeFile(t, fixedStateFile, []byte(fixedState))
	return fixedStateFile
}

// Behaves just like pt.TestLifeCycleInitAndDestroy() but with custom inner test logic. This
// function was obtained by inlining TestLifeCycleInitAndDestroy implementation and generalizing it.
func (pth *programTestHelper) lifecycleInitAndDestroy(customTest func() error) error {
	assert.Falsef(pth.t, pth.opts.RunUpdateTest, "RunUpdateTest is not supported")

	err := pth.pt.TestLifeCyclePrepare()
	if err != nil {
		return fmt.Errorf("copying test to temp dir %s: %w", "<tmpdir>", err)
	}

	pth.pt.TestFinished = false
	if pth.opts.DestroyOnCleanup {
		pth.t.Cleanup(pth.pt.TestCleanUp)
	} else {
		defer pth.pt.TestCleanUp()
	}

	err = pth.pt.TestLifeCycleInitialize()
	if err != nil {
		return fmt.Errorf("initializing test project: %w", err)
	}

	destroyStack := func() {
		destroyErr := pth.pt.TestLifeCycleDestroy()
		assert.NoError(pth.t, destroyErr)
	}
	if pth.opts.DestroyOnCleanup {
		// Allow other tests to refer to this stack until the test is complete.
		pth.t.Cleanup(destroyStack)
	} else {
		// Ensure that before we exit, we attempt to destroy and remove the stack.
		defer destroyStack()
	}

	if err = customTest(); err != nil {
		return err
	}

	pth.pt.TestFinished = true
	return nil
}

func (pth *programTestHelper) withUpdatedStackName(newStackName string, state string) string {
	pth.t.Logf("Replacing %q with %q", pth.parseStackName(state), newStackName)
	return strings.ReplaceAll(state, pth.parseStackName(state), newStackName)
}

func (pth *programTestHelper) parseStackName(state string) string {
	t := pth.t
	type model struct {
		Deployment struct {
			Resources []struct {
				URN  string `json:"urn"`
				Type string `json:"type"`
			} `json:"resources"`
		} `json:"deployment"`
	}
	var m model
	err := json.Unmarshal([]byte(state), &m)
	require.NoError(t, err)
	var stackUrn string
	for _, r := range m.Deployment.Resources {
		if r.Type == "pulumi:pulumi:Stack" {
			stackUrn = r.URN
		}
	}
	return strings.Split(stackUrn, ":")[2]
}

func (b *providerUpgradeBuilder) providerUpgradeRecordBaselines(t *testing.T) {
	info := b.newProviderUpgradeInfo(t)
	ambientProvider, _ := exec.LookPath(b.providerBinary())
	require.Emptyf(t, ambientProvider, "please remove the provider from PATH")
	ensureFolderExists(t, info.recordingDir)
	deleteFileIfExists(t, info.stateFile)
	deleteFileIfExists(t, info.grpcFile)
	test := integration.ProgramTestOptions{
		Dir: b.program,
		Env: append(os.Environ(), fmt.Sprintf("PULUMI_DEBUG_GRPC=%s", info.grpcFile)),
		ExportStateValidator: func(t *testing.T, state []byte) {
			writeFile(t, info.stateFile, state)
			t.Logf("wrote %s", info.stateFile)
		},
		Config: b.config,

		// TODO eks.Cluster fails refresh on 5.42.0
		SkipRefresh: true,
	}
	integration.ProgramTest(t, &test)
}

// There are some limitations in factoring out the provider versoin out of the YAML sources.
//
// To compensate, this function tries to extract the version for verification.
//
// See https://github.com/pulumi/pulumi-yaml/issues/508
func (b *providerUpgradeBuilder) verifyVersion() {
	f := filepath.Join(b.program, "Pulumi.yaml")
	actual := b.parseProviderVersion(f)
	expected := b.baselineVersion
	require.Equalf(b.tt, expected, actual,
		"Please check that %q specifies the %q provider version",
		f, b.baselineVersion)
}

func (b *providerUpgradeBuilder) parseProviderVersion(yamlFile string) string {
	type model struct {
		Resources struct {
			Provider struct {
				Options struct {
					Version string `yaml:"version"`
				} `yaml:"options"`
			} `yaml:"provider"`
		} `json:"resources"`
	}
	bytes, err := os.ReadFile(yamlFile)
	require.NoError(b.tt, err)
	var m model
	yaml.Unmarshal(bytes, &m)
	require.NoError(b.tt, err)
	v := m.Resources.Provider.Options.Version
	require.NotEmptyf(b.tt, v, "Failed to parse Pulumi.yaml: "+
		"resources.provider.options.version is empty")
	return v
}
