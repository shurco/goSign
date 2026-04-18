package password

import (
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	t.Run("returns a valid bcrypt hash", func(t *testing.T) {
		hash, err := GeneratePassword("correcthorsebatterystaple")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hash == "" {
			t.Fatal("expected non-empty hash")
		}
		if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
			t.Fatalf("expected bcrypt prefix, got %q", hash)
		}
	})

	t.Run("hash of same password differs due to salt", func(t *testing.T) {
		h1, _ := GeneratePassword("same")
		h2, _ := GeneratePassword("same")
		if h1 == h2 {
			t.Fatal("two hashes of the same password must differ")
		}
	})

	t.Run("rejects passwords longer than 72 bytes", func(t *testing.T) {
		long := strings.Repeat("a", 100)
		if _, err := GeneratePassword(long); err == nil {
			t.Fatal("expected error for password longer than 72 bytes, got nil")
		}
	})
}

func TestComparePasswords(t *testing.T) {
	hash, err := GeneratePassword("s3cret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ComparePasswords(hash, "s3cret") {
		t.Error("expected correct password to match")
	}
	if ComparePasswords(hash, "other") {
		t.Error("expected wrong password to NOT match")
	}
	if ComparePasswords("not-a-hash", "s3cret") {
		t.Error("expected invalid hash to fail comparison")
	}
}

func TestNewToken(t *testing.T) {
	tok, err := NewToken("payload")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tok) != 32 {
		t.Fatalf("expected md5 hex (32 chars), got %d: %q", len(tok), tok)
	}
	// Two invocations must produce different tokens because bcrypt salt changes.
	tok2, _ := NewToken("payload")
	if tok == tok2 {
		t.Fatal("NewToken should not be deterministic")
	}
}

func TestRandomString(t *testing.T) {
	seen := make(map[string]struct{}, 64)
	for i := 0; i < 64; i++ {
		s := RandomString()
		if len(s) != DefaultIdLength {
			t.Fatalf("expected length %d, got %d", DefaultIdLength, len(s))
		}
		for _, c := range s {
			if !strings.ContainsRune(DefaultIdAlphabet, c) {
				t.Fatalf("character %q is not in the allowed alphabet", c)
			}
		}
		if _, dup := seen[s]; dup {
			t.Fatalf("collision after few samples: %q", s)
		}
		seen[s] = struct{}{}
	}
}
