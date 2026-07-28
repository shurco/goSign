package workerclient

import "testing"

func TestTarget(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"127.0.0.1:8089", "127.0.0.1:8089"},
		{"worker-a:8089,worker-b:8089", "gosign-workers:///worker-a:8089,worker-b:8089"},
		{"dns:///worker:8089", "dns:///worker:8089"},
	}
	for _, tt := range tests {
		if got := target(tt.in); got != tt.want {
			t.Errorf("target(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGeoLocatorNilClient(t *testing.T) {
	g := NewGeoLocator(nil)
	if loc := g.GetLocation("8.8.8.8"); loc != nil {
		t.Errorf("GetLocation with nil client must return nil, got %+v", loc)
	}
}
