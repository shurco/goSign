package workerclient

import (
	"context"
	"time"

	workerv1 "github.com/shurco/gosign/internal/rpc/workerv1"
	"github.com/shurco/gosign/pkg/geolocation"
)

// lookupTimeout keeps request-path lookups fast even when workers are down.
const lookupTimeout = 800 * time.Millisecond

// GeoLocator adapts the worker LookupIP RPC to the geolocation API used by
// HTTP handlers. Lookups are best-effort: any failure yields nil.
type GeoLocator struct {
	client workerv1.WorkerServiceClient
}

// NewGeoLocator creates the adapter; client may be nil (lookups return nil).
func NewGeoLocator(client workerv1.WorkerServiceClient) *GeoLocator {
	return &GeoLocator{client: client}
}

// GetLocation resolves an IP via the worker; nil when unknown or unavailable.
func (g *GeoLocator) GetLocation(ip string) *geolocation.Location {
	if g.client == nil || ip == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), lookupTimeout)
	defer cancel()

	resp, err := g.client.LookupIP(ctx, &workerv1.LookupIPRequest{Ip: ip})
	if err != nil || !resp.GetFound() {
		return nil
	}
	return &geolocation.Location{
		City:    resp.GetCity(),
		Country: resp.GetCountry(),
		Full:    resp.GetFull(),
	}
}
