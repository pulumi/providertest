package pulumitest

import (
	"os"

	"github.com/pulumi/providertest/grpclog"
)

// GrpcLog reads the gRPC log for the current stack based on the PULUMI_DEBUG_GRPC env var.
func (pt *PulumiTest) GrpcLog() *grpclog.GrpcLog {
	pt.t.Helper()

	if pt.currentStack == nil {
		pt.t.Log("can't read gRPC log: no current stack")
		return nil
	}

	env := pt.CurrentStack().Workspace().GetEnvVars()
	if env == nil || env["PULUMI_DEBUG_GRPC"] == "" {
		pt.t.Log("can't read gRPC log: PULUMI_DEBUG_GRPC env var not set")
		return nil
	}

	log, err := grpclog.LoadLog(env["PULUMI_DEBUG_GRPC"])
	if err != nil {
		pt.t.Fatalf("failed to load grpc log: %s", err)
	}
	return log
}

// ClearGrpcLog clears the gRPC log for the current stack based on the PULUMI_DEBUG_GRPC env var.
func (pt *PulumiTest) ClearGrpcLog() {
	env := pt.CurrentStack().Workspace().GetEnvVars()
	if env == nil || env["PULUMI_DEBUG_GRPC"] == "" {
		pt.t.Log("can't clear gRPC log: PULUMI_DEBUG_GRPC env var not set")
		return
	}
	if err := os.RemoveAll(env["PULUMI_DEBUG_GRPC"]); err != nil {
		pt.t.Fatalf("failed to clear gRPC log: %s", err)
	}
}
