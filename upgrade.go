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
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/pulumi/providertest/flags"
	"github.com/pulumi/providertest/pulumitest/sanitize"
	"github.com/pulumi/providertest/replay"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// Verifies that upgrading the provider does not generate any unexpected replacements.
//
// Specifically check that for a given Pulumi program located in dir, users can run pulumi up on a
// baseline provider version, then upgrade the provider to the new version under test, run pulumi up
// again and observe an empty diff.
func (pt *ProviderTest) VerifyUpgrade(t *testing.T, mode UpgradeTestMode) {
	pt.newProviderUpgradeBuilder(t).run(t, mode)
}

func (pt *ProviderTest) VerifyUpgradeSnapshot(t *testing.T) {
	pt.newProviderUpgradeBuilder(t).providerUpgradeRecordBaselines(t)
}

func (pt *ProviderTest) newProviderUpgradeBuilder(t *testing.T) *providerUpgradeBuilder {
	require.NotEmptyf(t, pt.dir, "dir is required")
	return &providerUpgradeBuilder{
		tt:                  t,
		program:             pt.dir,
		config:              pt.config,
		providerUpgradeOpts: pt.upgradeOpts,
	}
}

// Tracks resource coverage through upgrade tests.
type upgradeCoverage struct {
	resources map[string]struct{}
}

func (u *upgradeCoverage) checkStateFile(t *testing.T, stateFile string) {
	type stack struct {
		Deployment struct {
			Resources []struct {
				Type string `json:"type"`
			} `json:"resources"`
		} `json:"deployment"`
	}
	b, err := os.ReadFile(stateFile)
	if err != nil {
		return // perhaps it did not exist, no matter
	}

	var st stack
	require.NoError(t, json.Unmarshal(b, &st))

	if u.resources == nil {
		u.resources = map[string]struct{}{}
	}

	for _, r := range st.Deployment.Resources {
		if strings.Contains(r.Type, "providers") {
			continue
		}
		if strings.Contains(r.Type, "Stack") {
			continue
		}
		u.resources[r.Type] = struct{}{}
	}
}

// This is a temporary helper method to assess upgrade resource coverage until better methods for
// tracking coverage are built. Run with -test.v to see the data logged. This finds all recorded
// GRPC states and traverses them to find the union of all resources used. It does not take into
// account if the corresponding tests are skipped or passing.
func ReportUpgradeCoverage(t *testing.T) {
	t.Helper()
	u := &upgradeCoverage{}
	dir := filepath.Join("testdata", "recorded", "TestProviderUpgrade")

	states := findFiles(t, dir, func(fn string) bool {
		return filepath.Base(fn) == "state.json"
	})

	for _, s := range states {
		u.checkStateFile(t, s)
	}

	covered := u.resources
	t.Logf("Resources covered: %d", len(covered))

	sorted := []string{}
	for k := range covered {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	for _, s := range sorted {
		t.Logf("- %s", s)
	}
}

// Enumerates various available modes to test provider upgrades. The modes differ in speed vs
// precision tradeoffs.
type UpgradeTestMode int

const (
	// The least precise but fastest mode. Tests are performed in-memory using pre-recorded
	// snapshots of baseline provider behavior. No cloud credentials are required, no
	// subprocesses are launched, fully debuggable.
	UpgradeTestMode_Quick UpgradeTestMode = iota

	// The medium precision/speed mode. Imports Pulumi statefile recorded on a baseline version,
	// and performs pulumi preview, asserting that the preview results in an empty diff. Cloud
	// credentials are required, but no infra is actually provisioned, making it quicker to
	// verify slow-to-provision resources such as databases.
	UpgradeTestMode_PreviewOnly

	// Full fidelity, slow mode. No pre-recorded snapshots are required. Do a complete pulumi up
	// on the baseline version, followed by a complete pulumi up on the version under test, and
	// assert that there are no observable updates or replacements.
	UpgradeTestMode_Full
)

func (m UpgradeTestMode) String() string {
	switch m {
	case UpgradeTestMode_PreviewOnly:
		return "preview-only"
	case UpgradeTestMode_Quick:
		return "quick"
	case UpgradeTestMode_Full:
		return "full"
	}
	return "<unknown>"
}

func UpgradeTestModes() []UpgradeTestMode {
	return []UpgradeTestMode{
		UpgradeTestMode_PreviewOnly,
		UpgradeTestMode_Quick,
		UpgradeTestMode_Full,
	}
}

func WithSkippedUpgradeTestMode(m UpgradeTestMode, reason string) Option {
	contract.Assertf(reason != "", "reason cannot be empty")
	contract.Assertf(m.String() != "<unknown>", "unknown UpgradeTestMode")
	return func(b *ProviderTest) {
		if b.upgradeOpts.modes == nil {
			b.upgradeOpts.modes = map[UpgradeTestMode]string{}
		}
		b.upgradeOpts.modes[m] = reason
	}
}

func WithBaselineVersion(v string) Option {
	contract.Assertf(v != "", "BaselineVersion cannot be empty")
	return func(b *ProviderTest) { b.upgradeOpts.baselineVersion = v }
}

// When testing upgrades, this option specifies additional baseline dependency versions. For
// example, when testing pulumi-eks, WithBaselineVersion("1.0.4") will define the baseline version
// of eks provider itself, where WithExtraBaselineDependencies(map[string]string{"aws": "5.42.0"})
// will pin the aws dependency.
func WithExtraBaselineDependencies(deps map[string]string) Option {
	return func(b *ProviderTest) { b.upgradeOpts.extraBaselineDeps = deps }
}

func WithProviderName(name string) Option {
	contract.Assertf(name != "", "ProviderName cannot be empty, "+
		"expecting a provider name like `gcp` or `aws`")
	return func(b *ProviderTest) { b.upgradeOpts.providerName = name }
}

// TODO[pulumi/providertest#9] make this redundant.
func WithResourceProviderServer(s pulumirpc.ResourceProviderServer) Option {
	contract.Assertf(s != nil, "ResourceProviderServer cannot be nil")
	return func(b *ProviderTest) { b.upgradeOpts.resourceProviderServer = s }
}

// The structure is mapped directly from Pulumi gRPC DiffResponse structure in the provider protocol
// and is currently unstable / subject to change.
//
// https://github.com/pulumi/pulumi/blob/master/proto/pulumi/provider.proto#L225
//
// Need to verify if this is representative of the actual Pulumi plans, since it only considers the
// decisions made by the provider, not the engine. For example, unclear if replaceOnChanges option
// https://www.pulumi.com/docs/concepts/options/replaceonchanges/ would surface here.
//
// Even with the above caveats, it is reasonable to rely on this for the purposes of testing the
// provider itself.
type Diff struct {
	URN        resource.URN
	HasChanges bool

	// Non-empty Replaces indicates that the plan is a resource replacement and not a simple
	// in-place update.
	Replaces []string

	Diffs               []string
	DeleteBeforeReplace bool

	// May only be populated if there's a change.
	Olds map[string]any
	// May only be populated if there's a change.
	News map[string]any
}

type Diffs []Diff

type DiffValidation = func(*testing.T, Diffs)

func WithDiffValidation(valid DiffValidation) Option {
	return func(b *ProviderTest) { b.upgradeOpts.diffValidation = valid }
}

func NoChanges() DiffValidation {
	return func(t *testing.T, diffs Diffs) {
		for _, d := range diffs {
			assert.Falsef(t, d.HasChanges, "Expected no changes for %v", d)
		}
	}
}

func NoReplacements() DiffValidation {
	return func(t *testing.T, diffs Diffs) {
		for _, d := range diffs {
			if d.HasChanges {
				assert.Emptyf(t, d.Replaces, "Unexpected replacement plan for %v", d)
			}
		}
	}
}

type providerUpgradeOpts struct {
	baselineVersion        string
	modes                  map[UpgradeTestMode]string // skip reason by mode
	providerName           string
	resourceProviderServer pulumirpc.ResourceProviderServer
	extraBaselineDeps      map[string]string
	diffValidation         DiffValidation
}

type providerUpgradeBuilder struct {
	tt      *testing.T
	program string
	config  map[string]string

	providerUpgradeOpts
}

func (b *providerUpgradeBuilder) run(t *testing.T, mode UpgradeTestMode) {
	if flags.Snapshot.IsSet() {
		t.Skipf("skipping because snapshot recording is in progress because %s",
			flags.Snapshot.WhySet())
	}

	switch mode {
	case UpgradeTestMode_Quick:
		if skip, ok := b.modes[UpgradeTestMode_Quick]; ok && skip != "" {
			t.Skip(skip)
		}
		if b.resourceProviderServer == nil {
			t.Skip("WithResourceProviderServer is required for quick mode")
		}
		b.checkProviderUpgradeQuick(t)
	case UpgradeTestMode_PreviewOnly:
		if skip, ok := b.modes[UpgradeTestMode_PreviewOnly]; ok && skip != "" {
			t.Skip(skip)
		}
		if testing.Short() {
			t.Skipf("Skipping in -short mode")
			return
		}
		b.checkProviderUpgradePreviewOnly(t)
	case UpgradeTestMode_Full:
		if skip, ok := b.modes[UpgradeTestMode_Full]; ok && skip != "" {
			t.Skip(skip)
		}
		t.Skip("Full mode is not supported yet")
	}
}

func (b *providerUpgradeBuilder) checkProviderUpgradeQuick(t *testing.T) {
	require.NotNilf(b.tt, b.resourceProviderServer, "WithResourceProviderServer is required")
	info := b.newProviderUpgradeInfo(t)

	bytes, err := os.ReadFile(info.grpcFile)
	require.NoErrorf(t, err,
		"No pre-recorded snapshots found; not recording because %s", flags.Snapshot.WhyNotSet())

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

// Verifies provider upgrades by replaying Diff calls. This is slightly involved. The available
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
		e.recordCheck(req)
	case "/pulumirpc.ResourceProvider/Diff":
		req := unmarshalProto(t, entry.Request, new(pulumirpc.DiffRequest))
		e.fixupDiff(t, req)
		entry.Request = marshalProto(t, req)
		b, err := json.Marshal(entry)
		require.NoError(t, err)
		replay.Replay(t, e.provider, string(b))
		e.verifiedDiffResourceCounter++
		t.Logf("Replayed Diff on %v", req.Urn)
	}
}

func (e *mockPulumiEngine) recordCheck(checkReq *pulumirpc.CheckRequest) {
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
	program := filepath.Base(b.program)
	info.recordingDir = filepath.Join("testdata", "recorded", "TestProviderUpgrade",
		program, b.baselineVersion)
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

	// Skip if state not yet created
	if _, err := os.Stat(info.stateFile); os.IsNotExist(err) {
		t.Logf("No pre-recorded state found for %s, recording baseline behavior.", b.baselineVersion)
		b.providerUpgradeRecordBaselines(t)
	}

	previewLogs := os.Getenv("PULUMI_DEBUG_GRPC")
	if previewLogs == "" {
		previewLogs = filepath.Join(t.TempDir(), "preview-grpc-logs.json")
	}
	t.Logf("Recording preview gRPC logs to %s", previewLogs)

	opts := integration.ProgramTestOptions{
		Dir:    b.program,
		Config: b.config,

		// Skips are required by programTestHelper.previewOnlyUpgradeTest
		SkipUpdate:       true,
		SkipRefresh:      true,
		SkipExportImport: true,

		Env: []string{fmt.Sprintf("PULUMI_DEBUG_GRPC=%s", previewLogs)},

		SkipEmptyPreviewUpdate: true,
	}

	opts = opts.With(b.optionsForPreviewOnly(t))

	ambientProvider, _ := exec.LookPath(b.providerBinary())
	require.NotEmptyf(t, ambientProvider, "expected to find a release candidate provider "+
		"binary in PATH, try to call `make provider` and `export PATH=$PWD/bin:$PATH`")

	pth := newProgramTestHelper(t, opts)
	err := pth.previewOnlyUpgradeTest(info.stateFile)
	require.NoError(t, err)

	diffV := NoChanges()
	if b.diffValidation != nil {
		diffV = b.diffValidation
	}
	verifyChanges(t, previewLogs, diffV)
}

func (b *providerUpgradeBuilder) optionsForPreviewOnly(t *testing.T) integration.ProgramTestOptions {
	projInfo, err := getProjInfo(b.program)
	require.NoError(t, err)
	switch rt := projInfo.Proj.Runtime.Name(); rt {
	case integration.YAMLRuntime:
		return integration.ProgramTestOptions{}
	case integration.NodeJSRuntime:
		return integration.ProgramTestOptions{
			// This will make ProgramTest issue `yarn link @pulumi/eks` or similar,
			// which will start testing the locally built Node SDK *if* it was installed
			// earlier with `yarn install`. Error paths might need some work here, that
			// is what happens if it is not installed yet.
			Dependencies: []string{fmt.Sprintf("@pulumi/%s", b.providerName)},
		}
	case integration.PythonRuntime:
		require.NoError(t, fmt.Errorf("python runtime does not yet support upgrade tests"))
	case integration.DotNetRuntime:
		require.NoError(t, fmt.Errorf("dotnet runtime does not yet support upgrade tests"))
	case integration.GoRuntime:
		require.NoError(t, fmt.Errorf("go runtime does not yet support upgrade tests"))
	case integration.JavaRuntime:
		require.NoError(t, fmt.Errorf("java Runtime does not yet support upgrade tests"))
	default:
		require.NoError(t, fmt.Errorf("unrecognized project runtime: %s", projInfo.Proj.Runtime.Name()))
	}
	return integration.ProgramTestOptions{}
}

func (b *providerUpgradeBuilder) providerBinary() string {
	return fmt.Sprintf("pulumi-resource-%s", b.providerName)
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

func (pth *programTestHelper) previewOnlyUpgradeTest(stateFile string) (finalError error) {
	t := pth.t
	pt := pth.pt
	opts := pth.opts
	return (&programTestWrapper{pth.pt}).lifecycleInitAndDestroy(t, opts, func() error {
		t.Logf("Backing up current stateFile")
		backupStateFile := filepath.Join(t.TempDir(), "backup-state.json")
		if err := pt.RunPulumiCommand("stack", "export", "--file",
			backupStateFile); err != nil {
			return err
		}

		defer func() {
			t.Logf("Restoring original stateFile")
			if err := pt.RunPulumiCommand("stack", "import", "--file",
				backupStateFile); err != nil {
				if finalError != nil {
					finalError = err
				}
			}
		}()

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
		require.Truef(t, opts.SkipEmptyPreviewUpdate,
			"expecting SkipEmptyPreviewUpdate: true")
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
	ensureFolderExists(t, info.recordingDir)
	deleteFileIfExists(t, info.stateFile)
	deleteFileIfExists(t, info.grpcFile)

	test := integration.ProgramTestOptions{
		Dir: b.program,
		Env: append(os.Environ(),
			// Record gRPC logs.
			fmt.Sprintf("PULUMI_DEBUG_GRPC=%s", info.grpcFile),
		),
		ExportStateValidator: func(t *testing.T, state []byte) {
			var stack apitype.UntypedDeployment
			err := json.Unmarshal(state, &stack)
			require.NoError(t, err)

			newStack, err := sanitize.SanitizeSecretsInStackState(&stack)
			require.NoError(t, err)

			newState, err := json.MarshalIndent(newStack, "", "  ")
			require.NoError(t, err)
			writeFile(t, info.stateFile, newState)
			t.Logf("wrote %s", info.stateFile)
		},
		Config: b.config,
		// We could record Refresh for posterity but it is not strictly needed for upgrade
		// tests only. It would be needed for tests that try to use snapshots to inform
		// import or refresh testing.
		SkipRefresh: true,
	}
	test = test.With(b.optionsForRecording(t))
	integration.ProgramTest(t, &test)
}

func (b *providerUpgradeBuilder) optionsForRecording(t *testing.T) integration.ProgramTestOptions {
	projInfo, err := getProjInfo(b.program)
	require.NoError(t, err)
	switch rt := projInfo.Proj.Runtime.Name(); rt {
	case integration.YAMLRuntime:
		return b.optionsForRecordingYAML(t)
	case integration.NodeJSRuntime:
		return b.optionsForRecordingNode()
	case integration.PythonRuntime:
		require.NoError(t, fmt.Errorf("python runtime does not yet support upgrade tests"))
	case integration.DotNetRuntime:
		require.NoError(t, fmt.Errorf("dotnet runtime does not yet support upgrade tests"))
	case integration.GoRuntime:
		require.NoError(t, fmt.Errorf("go runtime does not yet support upgrade tests"))
	case integration.JavaRuntime:
		require.NoError(t, fmt.Errorf("java runtime does not yet support upgrade tests"))
	default:
		require.NoError(t, fmt.Errorf("unrecognized project runtime: %s", projInfo.Proj.Runtime.Name()))
	}
	return integration.ProgramTestOptions{}
}

func (b *providerUpgradeBuilder) optionsForRecordingYAML(t *testing.T) integration.ProgramTestOptions {
	// There should be an elegant way to do this, but for the moment the code brute-forces the
	// issue and installs the baseline versions of necessary plugins in PATH as ambient plugins.
	ambients := []ambientPlugin{}
	ambients = append(ambients, ambientPlugin{
		Provider: b.providerName,
		Version:  b.baselineVersion,
	})
	for p, v := range b.extraBaselineDeps {
		ambients = append(ambients, ambientPlugin{
			Provider: p,
			Version:  v,
		})
	}

	path, err := pathWithAmbientPlugins(t, os.Getenv("PATH"), ambients...)
	require.NoError(t, err)
	// Cannot set PULUMI_IGNORE_AMBIENT_PLUGINS=true here because ambient plugins is how this
	// code installs the baseline dependencies.
	return integration.ProgramTestOptions{Env: []string{fmt.Sprintf("PATH=%s", path)}}
}

func (b *providerUpgradeBuilder) optionsForRecordingNode() integration.ProgramTestOptions {
	// Overrides will make ProgramTest install specific baseline versions of Node SDKs and that
	// in turn will make Pulumi CLI auto-install matching provider binaries.
	overrides := map[string]string{
		fmt.Sprintf("@pulumi/%s", b.providerName): b.baselineVersion,
	}
	for k, v := range b.extraBaselineDeps {
		overrides[fmt.Sprintf("@pulumi/%s", k)] = v
	}
	return integration.ProgramTestOptions{
		Overrides: overrides,

		// Make sure that local provider builds in PATH do not interfere with recording
		// baseline versions.
		Env: []string{"PULUMI_IGNORE_AMBIENT_PLUGINS=true"},
	}
}

func verifyChanges(t *testing.T, grpcLogsFile string, diffV DiffValidation) {
	bytes, err := os.ReadFile(grpcLogsFile)
	require.NoError(t, err)

	type req struct {
		URN  string         `json:"urn"`
		Olds map[string]any `json:"olds"`
		News map[string]any `json:"news"`
	}

	type resp struct {
		Changes             string   `json:"changes"`
		Diffs               []string `json:"diffs"`
		Replaces            []string `json:"replaces"`
		DeleteBeforeReplace bool     `json:"deleteBeforeReplace"`
	}

	type log struct {
		Method   string `json:"method"`
		Request  req    `json:"request"`
		Response resp   `json:"response"`
	}

	var diffs Diffs

	for _, s := range strings.Split(string(bytes), "\n") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		var entry log
		err = json.Unmarshal([]byte(s), &entry)
		require.NoError(t, err)

		if entry.Method == "/pulumirpc.ResourceProvider/Diff" || entry.Method == "/pulumirpc.ResourceProvider/DiffConfig" {
			urn, err := resource.ParseURN(entry.Request.URN)
			require.NoError(t, err)

			d := Diff{
				URN:                 urn,
				HasChanges:          entry.Response.Changes != "" && entry.Response.Changes != "DIFF_NONE",
				Diffs:               entry.Response.Diffs,
				Replaces:            entry.Response.Replaces,
				DeleteBeforeReplace: entry.Response.DeleteBeforeReplace,
			}

			if d.HasChanges {
				d.Olds = map[string]any{}
				d.News = map[string]any{}
				for _, rep := range d.Replaces {
					d.Olds[rep] = entry.Request.Olds[rep]
					d.News[rep] = entry.Request.News[rep]
				}
				for _, diffProp := range d.Diffs {
					d.Olds[diffProp] = entry.Request.Olds[diffProp]
					d.News[diffProp] = entry.Request.News[diffProp]
				}
			}

			diffs = append(diffs, d)
		}
	}

	diffV(t, diffs)
}
