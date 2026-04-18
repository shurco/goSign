package config

import (
	"reflect"
	"sort"
	"testing"
)

func TestSplitCommaNonEmpty(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{",,,", nil},
		{"a", []string{"a"}},
		{"a, b ,c", []string{"a", "b", "c"}},
		{"  ,x,  y,", []string{"x", "y"}},
	}
	for _, tt := range tests {
		got := splitCommaNonEmpty(tt.in)
		if len(got) == 0 && len(tt.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("splitCommaNonEmpty(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestGetenvBool(t *testing.T) {
	cases := map[string]bool{
		"":      false,
		"true":  true,
		"True":  true,
		"1":     true,
		"yes":   true,
		"false": false,
		"0":     false,
		"no":    false,
	}
	for raw, want := range cases {
		t.Setenv(envPrefix+"TEST_FLAG", raw)
		def := !want // choose default opposite of want so we verify parsing, not default
		if raw == "" {
			def = want
		}
		if got := getenvBool("TEST_FLAG", def); got != want {
			t.Errorf("getenvBool(%q, default=%v) = %v, want %v", raw, def, got, want)
		}
	}
}

func TestLoadRequiresJWTSecret(t *testing.T) {
	t.Setenv(envPrefix+"JWT_SECRET", "")
	if err := Load(); err == nil {
		t.Fatal("Load() must fail when JWT_SECRET is empty")
	}
}

func TestLoadAppliesDevCORSOriginsWhenDevMode(t *testing.T) {
	t.Setenv(envPrefix+"JWT_SECRET", "test-secret")
	t.Setenv(envPrefix+"DEV_MODE", "true")
	t.Setenv(envPrefix+"CORS_ALLOWED_ORIGINS", "")

	if err := Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := append([]string(nil), Data().CORSAllowedOrigins...)
	sort.Strings(got)
	want := append([]string(nil), defaultDevCORSOrigins...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dev-mode CORS origins mismatch:\n got:  %v\n want: %v", got, want)
	}
}

func TestLoadRespectsExplicitCORSOrigins(t *testing.T) {
	t.Setenv(envPrefix+"JWT_SECRET", "test-secret")
	t.Setenv(envPrefix+"DEV_MODE", "true")
	t.Setenv(envPrefix+"CORS_ALLOWED_ORIGINS", "https://example.com, https://foo.bar")

	if err := Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"https://example.com", "https://foo.bar"}
	if !reflect.DeepEqual(Data().CORSAllowedOrigins, want) {
		t.Errorf("CORSAllowedOrigins = %v, want %v", Data().CORSAllowedOrigins, want)
	}
}
