package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shurco/gosign/internal/testutil"
	"github.com/shurco/gosign/pkg/geolocation"
	"github.com/shurco/gosign/pkg/logging"
)

// TestMaintenanceLockExcludesConcurrentRuns proves that two callers (e.g.
// two worker replicas) never run maintenance simultaneously: the second one
// observes acquired=false while the first holds the advisory lock, and the
// lock becomes available again after release.
func TestMaintenanceLockExcludesConcurrentRuns(t *testing.T) {
	pool := testutil.NewTestDB(t)

	geo, _ := geolocation.NewService("")
	svc := newRPCService(pool, geo, logging.Log)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	holderInside := make(chan struct{})
	release := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		acquired, err := svc.withMaintenanceLock(ctx, func(context.Context) error {
			close(holderInside)
			<-release
			return nil
		})
		if err != nil {
			t.Errorf("holder: %v", err)
		}
		if !acquired {
			t.Error("holder: expected to acquire the lock")
		}
	}()

	<-holderInside

	// Second caller must be turned away while the lock is held.
	acquired, err := svc.withMaintenanceLock(ctx, func(context.Context) error {
		t.Error("contender must not run while the lock is held")
		return nil
	})
	if err != nil {
		t.Fatalf("contender: %v", err)
	}
	if acquired {
		t.Fatal("contender: expected acquired=false while the lock is held")
	}

	close(release)
	wg.Wait()

	// After release the lock must be obtainable again.
	acquired, err = svc.withMaintenanceLock(ctx, func(context.Context) error { return nil })
	if err != nil {
		t.Fatalf("after release: %v", err)
	}
	if !acquired {
		t.Fatal("after release: expected to acquire the lock")
	}
}
