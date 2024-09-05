package providers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pulumi/providertest/grpclog"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

// ReplayInvokes wraps a provider factory, intercepting all invokes and replaying them from a gRPC log.
// Example:
// providerFactory := providers.ResourceProviderFactory(providerServer)
// cacheDir := providertest.GetUpgradeCacheDir(filepath.Base(dir), "5.60.0")
// factoryWithReplay := providerFactory.ReplayInvokes(filepath.Join(cacheDir, "grpc.json"), true)
func (pf ProviderFactory) ReplayInvokes(grpcLogPath string, allowLiveFallback bool) ProviderFactory {
	interceptors := ProviderInterceptors{
		Invoke: func(ctx context.Context, in *pulumirpc.InvokeRequest, client pulumirpc.ResourceProviderClient) (*pulumirpc.InvokeResponse, error) {
			log, err := grpclog.LoadLog(grpcLogPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load gRPC log: %w", err)
			}
			invokes, err := log.Invokes()
			if err != nil {
				return nil, fmt.Errorf("failed to get invokes from log: %w", err)
			}
			requestedToken := in.GetTok()
			// Avoid using range due to invokes containing sync locks.
			for i := 0; i < len(invokes); i++ {
				if invokes[i].Request.Tok == requestedToken {
					if reflect.DeepEqual(in.Args.AsMap(), invokes[i].Request.Args.AsMap()) {
						return &invokes[i].Response, nil
					}
				}
			}
			if allowLiveFallback {
				return client.Invoke(ctx, in)
			} else {
				return nil, fmt.Errorf("failed to find invoke %s in gRPC log", requestedToken)
			}
		},
	}
	return func(ctx context.Context, pt PulumiTest) (Port, error) {
		port, err := pf(ctx, pt)
		if err != nil {
			return -1, err
		}
		interceptResourceProviderServer, err := NewProviderInterceptProxy(ctx, port, interceptors)
		if err != nil {
			return -1, err
		}
		return startResourceProviderServer(ctx, pt, func(pt PulumiTest) (pulumirpc.ResourceProviderServer, error) {
			return interceptResourceProviderServer, nil
		})
	}
}
