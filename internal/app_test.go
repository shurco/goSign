package app

import (
	"testing"

	"github.com/shurco/gosign/internal/config"
)

func TestModeHandling(t *testing.T) {
	// Test dev mode configuration
	cfg := &config.Config{
		DevMode:  true,
		HTTPAddr: "0.0.0.0:8088",
	}
	
	if !cfg.DevMode {
		t.Errorf("Expected DevMode to be true, got %v", cfg.DevMode)
	}
	
	// Test production mode configuration
	cfg = &config.Config{
		DevMode:  false,
		HTTPAddr: "0.0.0.0:8088",
	}
	
	if cfg.DevMode {
		t.Errorf("Expected DevMode to be false, got %v", cfg.DevMode)
	}
}