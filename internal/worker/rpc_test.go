package worker

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	workerv1 "github.com/shurco/gosign/internal/rpc/workerv1"
	"github.com/shurco/gosign/pkg/geolocation"
	"github.com/shurco/gosign/pkg/logging"
)

const testToken = "test-worker-token"

// newBufconnClient spins up the worker gRPC service (with auth) on an
// in-memory listener and returns a connected client.
func newBufconnClient(t *testing.T) workerv1.WorkerServiceClient {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() { _ = lis.Close() })

	geo, _ := geolocation.NewService("") // no-op mode: lookups return nil

	srv := grpc.NewServer(grpc.UnaryInterceptor(authUnaryInterceptor(testToken)))
	workerv1.RegisterWorkerServiceServer(srv, newRPCService(nil, geo, logging.Log))
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return workerv1.NewWorkerServiceClient(conn)
}

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// authCtx attaches a valid bearer token to the outgoing call.
func authCtx(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+testToken)
}

func TestAuthRejectsUnauthenticatedCalls(t *testing.T) {
	client := newBufconnClient(t)
	ctx := testCtx(t)

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{"no token", ctx},
		{"wrong token", metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer wrong-token")},
		{"malformed header", metadata.AppendToOutgoingContext(ctx, "authorization", testToken)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Status(tt.ctx, &workerv1.StatusRequest{})
			if status.Code(err) != codes.Unauthenticated {
				t.Fatalf("Status: expected Unauthenticated, got %v", err)
			}
		})
	}
}

func TestStatusWithAuth(t *testing.T) {
	client := newBufconnClient(t)

	resp, err := client.Status(authCtx(testCtx(t)), &workerv1.StatusRequest{})
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if resp.GetStartedAt() == nil {
		t.Error("Status: started_at must be set")
	}
	if resp.GetLastRunAt() != nil {
		t.Error("Status: last_run_at must be empty before the first run")
	}
}

func TestLookupIPWithoutDatabase(t *testing.T) {
	client := newBufconnClient(t)

	resp, err := client.LookupIP(authCtx(testCtx(t)), &workerv1.LookupIPRequest{Ip: "8.8.8.8"})
	if err != nil {
		t.Fatalf("LookupIP: %v", err)
	}
	if resp.GetFound() {
		t.Error("LookupIP: found must be false without a database")
	}
}
