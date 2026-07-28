// Package workerclient is the server-side gRPC client to worker replicas:
// authenticated calls, client-side load balancing, and an adapter exposing
// worker-backed geolocation lookups to HTTP handlers.
package workerclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"

	workerv1 "github.com/shurco/gosign/internal/rpc/workerv1"
)

// staticScheme resolves comma-separated address lists for round robin.
const staticScheme = "gosign-workers"

func init() {
	resolver.Register(staticResolverBuilder{})
}

// New dials worker replicas. addr may be a single "host:port", a
// comma-separated list of addresses (client-side round robin), or a full
// gRPC target such as "dns:///worker:8089" (round robin over DNS records,
// e.g. docker compose service replicas).
func New(addr, token string) (*grpc.ClientConn, workerv1.WorkerServiceClient, error) {
	conn, err := grpc.NewClient(target(addr),
		// The channel stays on localhost / the private docker network;
		// calls are authenticated with a bearer token per RPC.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(tokenCreds(token)),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		// Reconnect backoff is capped so worker recovery is noticed quickly.
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  time.Second,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   15 * time.Second,
			},
			MinConnectTimeout: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("worker client: %w", err)
	}
	return conn, workerv1.NewWorkerServiceClient(conn), nil
}

// target normalizes the configured address into a gRPC target.
func target(addr string) string {
	if strings.Contains(addr, "://") {
		return addr
	}
	if strings.Contains(addr, ",") {
		return staticScheme + ":///" + addr
	}
	return addr
}

// tokenCreds attaches the worker bearer token to every RPC.
type tokenCreds string

func (t tokenCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + string(t)}, nil
}

// RequireTransportSecurity is false: transport protection comes from network
// isolation (localhost / private network), authentication from the token.
func (t tokenCreds) RequireTransportSecurity() bool { return false }

// staticResolverBuilder resolves "gosign-workers:///a:1,b:2" to a fixed
// address set for client-side round robin.
type staticResolverBuilder struct{}

func (staticResolverBuilder) Scheme() string { return staticScheme }

func (staticResolverBuilder) Build(t resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	var addrs []resolver.Address
	for _, a := range strings.Split(t.Endpoint(), ",") {
		if a = strings.TrimSpace(a); a != "" {
			addrs = append(addrs, resolver.Address{Addr: a})
		}
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("worker client: empty address list in target %q", t)
	}
	if err := cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		return nil, err
	}
	return nopResolver{}, nil
}

type nopResolver struct{}

func (nopResolver) ResolveNow(resolver.ResolveNowOptions) {}
func (nopResolver) Close()                                {}
