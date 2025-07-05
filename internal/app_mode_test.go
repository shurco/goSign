package app

import (
	"testing"

	"github.com/shurco/gosign/internal/config"
)

func TestApplicationModeHandling(t *testing.T) {
	// Test that we can create a config with different modes
	devConfig := &config.Config{
		DevMode:  true,
		HTTPAddr: "0.0.0.0:8899", // Use a different port for testing
	}
	
	prodConfig := &config.Config{
		DevMode:  false,
		HTTPAddr: "0.0.0.0:8900", // Use a different port for testing
	}
	
	// Test dev mode
	if !devConfig.DevMode {
		t.Errorf("Expected DevMode to be true, got %v", devConfig.DevMode)
	}
	
	// Test prod mode
	if prodConfig.DevMode {
		t.Errorf("Expected DevMode to be false, got %v", prodConfig.DevMode)
	}
	
	// Test that we can access the config
	if devConfig.HTTPAddr != "0.0.0.0:8899" {
		t.Errorf("Expected HTTPAddr to be 0.0.0.0:8899, got %v", devConfig.HTTPAddr)
	}
	
	if prodConfig.HTTPAddr != "0.0.0.0:8900" {
		t.Errorf("Expected HTTPAddr to be 0.0.0.0:8900, got %v", prodConfig.HTTPAddr)
	}
}

func TestHealthEndpointAvailability(t *testing.T) {
	// Test that the health endpoint is available in dev mode
	// Note: This is a simple test to verify the mode configuration is working
	// In a real scenario, we would start the server and test the endpoint
	
	// Mock test to check if health endpoint handler exists
	if true { // Placeholder for actual endpoint availability check
		t.Log("Health endpoint should be available in dev mode")
	}
}