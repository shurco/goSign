package geolocation

import (
	"fmt"
	"net/netip"
	"strings"
	"sync"

	"github.com/oschwald/geoip2-golang/v2"
)

// Location represents geolocation data for an IP address
type Location struct {
	City    string `json:"city"`
	Country string `json:"country"`
	Full    string `json:"full"` // "City, Country" format
}

// Service provides geolocation lookup functionality using GeoLite2 database
type Service struct {
	db   *geoip2.Reader
	mu   sync.RWMutex
	path string
}

// NewService creates a new geolocation service
// dbPath is the path to GeoLite2-City.mmdb file
// If dbPath is empty, the service will work in no-op mode (returns empty location)
func NewService(dbPath string) (*Service, error) {
	s := &Service{
		path: dbPath,
	}

	if dbPath == "" {
		return s, nil
	}

	db, err := geoip2.Open(dbPath)
	if err != nil {
		// Keep service non-nil so callers can still use Reload() once the DB appears.
		// GetLocation will return empty results while db is nil.
		return s, fmt.Errorf("failed to open GeoLite2 database: %w", err)
	}

	s.db = db
	return s, nil
}

// GetLocation returns location data for the given IP address
// Returns nil if location cannot be determined
func (s *Service) GetLocation(ipStr string) *Location {
	if s.db == nil || ipStr == "" {
		return nil
	}

	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	record, err := s.db.City(addr)
	if err != nil {
		return nil
	}

	loc := &Location{}

	// Get city name (prefer English, fallback to any available)
	if city := record.City.Names.English; city != "" {
		loc.City = city
	} else {
		names := record.City.Names
		if names.Spanish != "" {
			loc.City = names.Spanish
		} else if names.French != "" {
			loc.City = names.French
		} else if names.German != "" {
			loc.City = names.German
		} else if names.Russian != "" {
			loc.City = names.Russian
		} else if names.Japanese != "" {
			loc.City = names.Japanese
		} else if names.SimplifiedChinese != "" {
			loc.City = names.SimplifiedChinese
		} else if names.BrazilianPortuguese != "" {
			loc.City = names.BrazilianPortuguese
		}
	}

	// Get country name (prefer English, fallback to any available)
	if country := record.Country.Names.English; country != "" {
		loc.Country = country
	} else {
		names := record.Country.Names
		if names.Spanish != "" {
			loc.Country = names.Spanish
		} else if names.French != "" {
			loc.Country = names.French
		} else if names.German != "" {
			loc.Country = names.German
		} else if names.Russian != "" {
			loc.Country = names.Russian
		} else if names.Japanese != "" {
			loc.Country = names.Japanese
		} else if names.SimplifiedChinese != "" {
			loc.Country = names.SimplifiedChinese
		} else if names.BrazilianPortuguese != "" {
			loc.Country = names.BrazilianPortuguese
		}
	}

	// Build full location string
	var parts []string
	if loc.City != "" {
		parts = append(parts, loc.City)
	}
	if loc.Country != "" {
		parts = append(parts, loc.Country)
	}
	if len(parts) > 0 {
		loc.Full = strings.Join(parts, ", ")
	}

	if loc.Full == "" {
		return nil
	}

	return loc
}

// GetLocationString returns location string in format "City, Country" for the given IP address
// Returns empty string if location cannot be determined
// This is a convenience method for backward compatibility
func (s *Service) GetLocationString(ipStr string) string {
	loc := s.GetLocation(ipStr)
	if loc == nil {
		return ""
	}
	return loc.Full
}

// Close closes the GeoLite2 database
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Reload reloads the GeoLite2 database from the configured path
func (s *Service) Reload() error {
	if s.path == "" {
		return nil
	}

	// Open new reader first so we don't drop the current DB on failure.
	db, err := geoip2.Open(s.path)
	if err != nil {
		return fmt.Errorf("failed to reload GeoLite2 database: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.db
	s.db = db

	if old != nil {
		_ = old.Close()
	}
	return nil
}
