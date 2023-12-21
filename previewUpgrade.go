package providertest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/providers"
	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

type UpgradePreviewOpts struct {
	// Where the recording cache is located containing previous deployment state and gRPC logs.
	// The variable {baselineVersion} which will be replaced with the baseline version.
	// The variable {programDir} which will be replaced with the program's directory name.
	// If empty, this will default to "testdata/recorded/TestProviderUpgrade/{programDir}/{baselineVersion}"
	CacheDir string

	BaselineProviderConfig ProviderConfiguration
}

type UpgradePreviewResult struct {
	auto.PreviewResult
	// The full parsed diffs from the preview.
	Diffs Diffs
}

type Snapshot struct {
	StackExport apitype.UntypedDeployment
	GrpcLog     []byte
}

// PreviewUpdate
func PreviewUpgrade(pulumiTest *pulumitest.PulumiTest, baselineProviderConfig ProviderConfiguration, opts UpgradePreviewOpts) auto.PreviewResult {
	pulumiTest.T().Helper()
	snapshot, hasSnapshot := TryReadSnapshot(pulumiTest.T(), opts.CacheDir)
	if !hasSnapshot {
		pulumiTest.T().Logf("no snapshot cache found at %s, creating one", opts.CacheDir)
		snapshot = SnapshotUp(pulumiTest, baselineProviderConfig)
		err := WriteSnapshot(opts.CacheDir, snapshot)
		if err != nil {
			pulumiTest.T().Fatalf("failed to write snapshot to %s: %v", opts.CacheDir, err)
		}
	}
	return PreviewUpdateFromSnapshot(pulumiTest, snapshot)
}

func PreviewUpdateFromSnapshot(pulumiTest *pulumitest.PulumiTest, snapshot *Snapshot) auto.PreviewResult {
	previewCopy := pulumiTest.CopyToTempDir()
	previewCopy.InstallStack("test")
	previewCopy.ImportStack(snapshot.StackExport)
	return previewCopy.Preview()
}

func SnapshotUp(pulumiTest *pulumitest.PulumiTest, provider ProviderConfiguration, extraOpts ...opttest.Option) *Snapshot {
	pulumiTest.T().Helper()
	recording := pulumiTest.CopyToTempDir()
	// Copy options
	opts := append([]opttest.Option{}, extraOpts...)
	// Set up gRPC recording
	grpcLogPath := filepath.Join(recording.Source(), "grpc.log")
	opts = append(opts, opttest.Env("PULUMI_DEBUG_GRPC", grpcLogPath))
	providerOpt, err := provider.BuildOpt()
	if err != nil {
		pulumiTest.T().Fatalf("failed to build provider options: %v", err)
	}
	opts = append(opts, providerOpt)
	recording.WithOptions(opts...)

	recording.InstallStack("test") // TODO: configure
	recording.Up()
	stackExport := recording.ExportStack()
	grpcLogBytes, err := os.ReadFile(grpcLogPath)
	if err != nil && !os.IsNotExist(err) {
		pulumiTest.T().Fatalf("failed to read grpc log for %s: %v", recording.Source(), err)
	}

	return &Snapshot{
		StackExport: stackExport,
		GrpcLog:     grpcLogBytes,
	}
}

func WriteSnapshot(dir string, snapshot *Snapshot) error {
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

// ProviderConfiguration is the name plus a discriminated union of the different way to configure the provider.
// Only one of Attach, LocalBinary, or DownloadBinary should be set.
type ProviderConfiguration struct {
	// The name of the provider without the "pulumi-" prefix.
	Name string
	// The factory to use to start the provider then attach when executing operations.
	Attach providers.ProviderFactory
	// The path to a local binary to use when executing operations.
	LocalBinaryPath string
	// The version of the provider to download and install.
	DownloadBinaryVersion string
	// Disable attaching the provider to the program under test.
	// Just set the path the to provider binary in the program's Pulumi.yaml.
	// This cannot be used with the Attach option.
	DisableAttach bool
}

func (p ProviderConfiguration) BuildOpt() (opttest.Option, error) {
	if p.Name == "" {
		return nil, fmt.Errorf("provider name must be set")
	}
	if p.Attach != nil {
		if p.DisableAttach {
			return nil, fmt.Errorf("cannot Attach when DisableAttach set")
		}
		return opttest.AttachProvider(p.Name, p.Attach), nil
	}
	if p.LocalBinaryPath != "" {
		if p.DisableAttach {
			return opttest.LocalProviderPath(p.Name, p.LocalBinaryPath), nil
		}
		return opttest.AttachProviderBinary(p.Name, p.LocalBinaryPath), nil
	}
	if p.DownloadBinaryVersion != "" {
		if p.DisableAttach {
			path, err := providers.DownloadPluginBinary(p.Name, p.DownloadBinaryVersion)
			if err != nil {
				return nil, err
			}
			return opttest.LocalProviderPath(p.Name, path), nil
		}
		return opttest.AttachDownloadedPlugin(p.Name, p.DownloadBinaryVersion), nil
	}
	return nil, fmt.Errorf("provider configuration must have one of Attach, LocalBinaryPath, or DownloadBinaryVersion set")
}
