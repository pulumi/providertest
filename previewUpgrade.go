package providertest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// type UpgradePreviewOpts struct {
// 	// Where the recording cache is located containing previous deployment state (stack.json) and gRPC log (grpc.log).
// 	// If empty, this will not cache the snapshot. Specify "." to use the current working directory.
// 	CacheDir string

// 	// The provider configuration to use when executing the initial deployment.
// 	BaselineProviderConfig providertest.ProviderConfiguration
// }

type UpgradePreviewResult struct {
	auto.PreviewResult
	// The full parsed diffs from the preview.
	Diffs Diffs
}

type Snapshot struct {
	StackExport apitype.UntypedDeployment
	GrpcLog     []byte
}

// // PreviewUpgrade
// func PreviewUpgrade(pulumiTest *pulumitest.PulumiTest, baselineProviderConfig ProviderConfiguration, opts UpgradePreviewOpts) auto.PreviewResult {
// 	pulumiTest.T().Helper()
// 	var snapshot *Snapshot
// 	if opts.CacheDir == "" {
// 		snapshot = SnapshotUp(pulumiTest, baselineProviderConfig)
// 	} else {
// 		var hasSnapshot bool
// 		snapshot, hasSnapshot = TryReadSnapshot(pulumiTest.T(), opts.CacheDir)
// 		if !hasSnapshot {
// 			pulumiTest.T().Logf("no snapshot cache found at %s, creating one", opts.CacheDir)
// 			snapshot = SnapshotUp(pulumiTest, baselineProviderConfig)
// 			err := WriteSnapshot(opts.CacheDir, snapshot)
// 			if err != nil {
// 				pulumiTest.T().Fatalf("failed to write snapshot to %s: %v", opts.CacheDir, err)
// 			}
// 		}
// 	}
// 	return PreviewUpdateFromSnapshot(pulumiTest, snapshot)
// }

// Setup is run in a separate temp folder. Uses the same stack name as the original test, if set.
func CachedSnapshot(pulumiTest *pulumitest.PulumiTest, cacheDir string, setup func(test *pulumitest.PulumiTest), setupOpts ...opttest.Option) *Snapshot {
	pulumiTest.T().Helper()
	var snapshot *Snapshot
	if cacheDir == "" {
		pulumiTest.T().Fatal("CacheDir is required. Use \".\" to use the current working directory.")
	}
	// Maintain current stack name
	currentStack := pulumiTest.CurrentStack()
	stackName := ""
	if currentStack != nil {
		stackName = currentStack.Name()
	} else {
		pulumiTest.T().Logf("warning: no current stack found, stack name will be randomised every run. Initialise a stack before calling CachedSnapshot to avoid this.")
	}

	var hasSnapshot bool
	snapshot, hasSnapshot = TryReadSnapshot(pulumiTest.T(), cacheDir)
	if hasSnapshot {
		return snapshot
	}
	pulumiTest.T().Logf("no snapshot cache found at %s, creating one", cacheDir)
	setupCopy := pulumiTest.CopyToTempDir(setupOpts...)
	setupCopy.InstallStack(stackName)
	snapshot = CaptureSnapshot(setupCopy, setup)
	err := WriteSnapshot(cacheDir, snapshot)
	if err != nil {
		pulumiTest.T().Fatalf("failed to write snapshot to %s: %v", cacheDir, err)
	}
	return snapshot
}

// PreviewUpdateFromSnapshot runs `pulumi preview` with a specific provider config and captures the resulting snapshot.
func PreviewUpdateFromSnapshot(pulumiTest *pulumitest.PulumiTest, snapshot *Snapshot) auto.PreviewResult {
	pulumiTest.T().Helper()
	previewCopy := pulumiTest.CopyToTempDir()
	previewCopy.InstallStack("test")
	previewCopy.ImportStack(snapshot.StackExport)
	return previewCopy.Preview()
}

// // SnapshotUp runs `pulumi up` with a specific provider config and captures the resulting snapshot.
// // Deprecated: use CachedSnapshot instead.
// func SnapshotUp(pulumiTest *pulumitest.PulumiTest, opts ...opttest.Option) *Snapshot {
// 	pulumiTest.T().Helper()

// 	return CaptureSnapshot(pulumiTest, func(test *pulumitest.PulumiTest) {
// 		test.Up()
// 	}, opts...)
// }

// CaptureSnapshot runs the given steps and captures the resulting stack state and gRPC log.
func CaptureSnapshot(source *pulumitest.PulumiTest, executeSteps func(test *pulumitest.PulumiTest), extraOpts ...opttest.Option) *Snapshot {
	source.T().Helper()
	recordingDir := source.T().TempDir()
	// Copy options
	opts := append([]opttest.Option{}, extraOpts...)
	// Set up gRPC recording
	grpcLogPath := filepath.Join(recordingDir, "grpc.log")
	opts = append(opts, opttest.Env("PULUMI_DEBUG_GRPC", grpcLogPath))
	recording := source.CopyTo(recordingDir, opts...)

	recording.InstallStack("test")
	executeSteps(recording)
	stackExport := recording.ExportStack()
	grpcLogBytes, err := os.ReadFile(grpcLogPath)
	if err != nil && !os.IsNotExist(err) {
		source.T().Fatalf("failed to read grpc log for %s: %v", recording.Source(), err)
	}

	return &Snapshot{
		StackExport: stackExport,
		GrpcLog:     grpcLogBytes,
	}
}

func WriteSnapshot(dir string, snapshot *Snapshot) error {
	if snapshot == nil {
		return fmt.Errorf("snapshot must not be nil")
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	stackBytes, err := json.Marshal(snapshot.StackExport)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(dir, "stack.json"), stackBytes, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "grpc.log"), snapshot.GrpcLog, 0644)
	return err
}

// TryReadSnapshot reads a snapshot from the given directory.
// If the snapshot does not exist, returns nil, nil.
func TryReadSnapshot(t *testing.T, path string) (*Snapshot, bool) {
	t.Helper()
	stackBytes, err := os.ReadFile(filepath.Join(path, "stack.json"))
	if err != nil {
		if os.IsNotExist(err) {
			t.Logf("stack.json export snapshot not found at %s", path)
			return nil, false
		}
		t.Fatalf("failed to read stack.json export snapshot at %s: %v", path, err)
		return nil, false
	}
	var stackExport apitype.UntypedDeployment
	err = json.Unmarshal(stackBytes, &stackExport)
	if err != nil {
		t.Fatalf("failed to unmarshal stack.json export snapshot at %s: %v", path, err)
		return nil, false
	}
	grpcLogBytes, err := os.ReadFile(filepath.Join(path, "grpc.log"))
	if err != nil {
		if os.IsNotExist(err) {
			t.Logf("warning: grpc.log not found at %s, continuing without log. Delete stack.json to force re-capture.", path)
		} else {
			t.Fatalf("failed to read grpc.log export snapshot at %s: %v", path, err)
			return nil, false
		}
	}
	return &Snapshot{
		StackExport: stackExport,
		GrpcLog:     grpcLogBytes,
	}, true
}
