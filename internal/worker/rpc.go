package worker

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	workerv1 "github.com/shurco/gosign/internal/rpc/workerv1"
	"github.com/shurco/gosign/internal/trust"
	"github.com/shurco/gosign/internal/worker/tasks"
	"github.com/shurco/gosign/pkg/appdir"
	"github.com/shurco/gosign/pkg/geolocation"
	"github.com/shurco/gosign/pkg/logging"
)

// maintenanceLockKey is the Postgres advisory lock shared by all worker
// replicas: the replica that holds it runs maintenance, the others skip.
// The lock is session-scoped, so Postgres releases it automatically if the
// holding connection dies mid-update.
const maintenanceLockKey = int64(0x676F5369676E0001) // "goSign" + 0001

// rpcService implements workerv1.WorkerServiceServer.
type rpcService struct {
	workerv1.UnimplementedWorkerServiceServer

	pool *pgxpool.Pool
	log  *logging.Logger
	geo  *geolocation.Service

	startedAt time.Time

	mu        sync.Mutex
	lastRunAt time.Time
}

func newRPCService(pool *pgxpool.Pool, geo *geolocation.Service, log *logging.Logger) *rpcService {
	return &rpcService{pool: pool, geo: geo, log: log, startedAt: time.Now()}
}

func (s *rpcService) markRun() {
	s.mu.Lock()
	s.lastRunAt = time.Now()
	s.mu.Unlock()
}

// withMaintenanceLock runs fn under the cross-replica advisory lock.
// Returns acquired=false when another replica holds the lock.
func (s *rpcService) withMaintenanceLock(ctx context.Context, fn func(context.Context) error) (acquired bool, err error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("acquire db connection: %w", err)
	}
	defer conn.Release()

	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", maintenanceLockKey).Scan(&acquired); err != nil {
		return false, fmt.Errorf("try advisory lock: %w", err)
	}
	if !acquired {
		return false, nil
	}
	defer func() {
		// Unlock must run even when ctx is already cancelled.
		unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if _, unlockErr := conn.Exec(unlockCtx, "SELECT pg_advisory_unlock($1)", maintenanceLockKey); unlockErr != nil {
			s.log.Warn().Err(unlockErr).Msg("Failed to release maintenance lock")
		}
	}()

	return true, fn(ctx)
}

// runScheduled executes all maintenance tasks; only one replica at a time.
func (s *rpcService) runScheduled(ctx context.Context) {
	acquired, err := s.withMaintenanceLock(ctx, func(ctx context.Context) error {
		s.markRun()
		if err := trust.Update(); err != nil {
			s.log.Err(err).Msg("Adobe trust list update failed")
		}
		tasks.SyncGeoLite2(ctx, s.pool, s.log)
		return nil
	})
	if err != nil {
		s.log.Err(err).Msg("Scheduled maintenance failed")
		return
	}
	if !acquired {
		s.log.Info().Msg("Maintenance skipped: another worker replica holds the lock")
		return
	}
	if err := s.geo.ReloadIfChanged(); err != nil {
		s.log.Warn().Err(err).Msg("GeoLite2 reload failed after maintenance")
	}
}

// Status reports worker health and the state of maintenance tasks.
func (s *rpcService) Status(_ context.Context, _ *workerv1.StatusRequest) (*workerv1.StatusResponse, error) {
	resp := &workerv1.StatusResponse{
		StartedAt: timestamppb.New(s.startedAt),
	}

	s.mu.Lock()
	if !s.lastRunAt.IsZero() {
		resp.LastRunAt = timestamppb.New(s.lastRunAt)
	}
	s.mu.Unlock()

	if fi, err := os.Stat(appdir.GeoLite2()); err == nil {
		resp.Geolite2Available = true
		resp.Geolite2ModifiedAt = timestamppb.New(fi.ModTime())
	}
	return resp, nil
}

// LookupIP resolves an IP address using the worker-owned GeoLite2 database.
func (s *rpcService) LookupIP(_ context.Context, req *workerv1.LookupIPRequest) (*workerv1.LookupIPResponse, error) {
	loc := s.geo.GetLocation(req.GetIp())
	if loc == nil {
		return &workerv1.LookupIPResponse{Found: false}, nil
	}
	return &workerv1.LookupIPResponse{
		Found:   true,
		City:    loc.City,
		Country: loc.Country,
		Full:    loc.Full,
	}, nil
}

// SyncGeoLite2 triggers an on-demand GeoLite2 database refresh.
func (s *rpcService) SyncGeoLite2(ctx context.Context, req *workerv1.SyncGeoLite2Request) (*workerv1.SyncGeoLite2Response, error) {
	var outcome tasks.SyncOutcome
	acquired, err := s.withMaintenanceLock(ctx, func(ctx context.Context) error {
		s.markRun()
		var syncErr error
		outcome, syncErr = tasks.SyncGeoLite2OnDemand(ctx, s.pool, s.log, tasks.SyncParams{
			Force:      req.GetForce(),
			Method:     req.GetMethod(),
			URL:        req.GetUrl(),
			LicenseKey: req.GetLicenseKey(),
			AccountID:  req.GetAccountId(),
		})
		return syncErr
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "geolite2 sync: %v", err)
	}
	if !acquired {
		return &workerv1.SyncGeoLite2Response{Status: workerv1.SyncStatus_SYNC_STATUS_BUSY}, nil
	}

	if err := s.geo.ReloadIfChanged(); err != nil {
		s.log.Warn().Err(err).Msg("GeoLite2 reload failed after sync")
	}

	st := workerv1.SyncStatus_SYNC_STATUS_SUCCESS
	if outcome == tasks.SyncOutcomeSkipped {
		st = workerv1.SyncStatus_SYNC_STATUS_SKIPPED
	}
	return &workerv1.SyncGeoLite2Response{Status: st}, nil
}

// UpdateTrustLists triggers an on-demand Adobe trust list refresh.
func (s *rpcService) UpdateTrustLists(ctx context.Context, _ *workerv1.UpdateTrustListsRequest) (*workerv1.UpdateTrustListsResponse, error) {
	acquired, err := s.withMaintenanceLock(ctx, func(context.Context) error {
		s.markRun()
		return trust.Update()
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "trust list update: %v", err)
	}
	if !acquired {
		return &workerv1.UpdateTrustListsResponse{Status: workerv1.SyncStatus_SYNC_STATUS_BUSY}, nil
	}
	return &workerv1.UpdateTrustListsResponse{Status: workerv1.SyncStatus_SYNC_STATUS_SUCCESS}, nil
}

// serveGRPC binds the internal gRPC endpoint and serves in the background.
// All calls must carry the bearer token. The returned server must be
// stopped with GracefulStop on shutdown.
func serveGRPC(addr, token string, svc *rpcService, log *logging.Logger) (*grpc.Server, error) {
	if token == "" {
		return nil, fmt.Errorf("worker grpc: auth token must not be empty")
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("worker grpc: listen on %s: %w", addr, err)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(authUnaryInterceptor(token)))
	workerv1.RegisterWorkerServiceServer(srv, svc)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Err(err).Msg("worker grpc: serve failed")
		}
	}()
	return srv, nil
}

// authUnaryInterceptor rejects calls without a valid bearer token.
func authUnaryInterceptor(token string) grpc.UnaryServerInterceptor {
	want := []byte("Bearer " + token)
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		auth := md.Get("authorization")
		if len(auth) == 0 || subtle.ConstantTimeCompare([]byte(auth[0]), want) != 1 {
			return nil, status.Error(codes.Unauthenticated, "invalid worker token")
		}
		return handler(ctx, req)
	}
}
