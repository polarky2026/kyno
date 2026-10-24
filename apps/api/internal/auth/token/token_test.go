package token

import (
	"bytes"
	"testing"
)

func TestNewReturnsURLSafeTokenAndItsHash(t *testing.T) {
	plain, hash, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if len(plain) != 43 {
		t.Fatalf("token length = %d, want 43", len(plain))
	}
	if !bytes.Equal(hash, Hash(plain)) {
		t.Fatal("returned hash does not match Hash(token)")
	}
	if len(hash) != 32 {
		t.Fatalf("hash length = %d, want 32", len(hash))
	}
}

func TestNewReturnsDifferentTokens(t *testing.T) {
	first, _, _ := New()
	second, _, _ := New()

	if first == second {
		t.Fatal("two tokens must differ")
	}
}

func TestHashIsDeterministic(t *testing.T) {
	if !bytes.Equal(Hash("abc"), Hash("abc")) {
		t.Fatal("Hash must return the same value for the same input")
	}
	if bytes.Equal(Hash("abc"), Hash("abd")) {
		t.Fatal("different tokens must have different hashes")
	}
}
