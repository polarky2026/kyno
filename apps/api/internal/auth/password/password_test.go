package password

import (
	"errors"
	"strings"
	"testing"
)

func TestHashAndVerify(t *testing.T) {
	encoded, err := Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := Verify("correct horse battery staple", encoded)
	if err != nil || !ok {
		t.Fatalf("Verify with the right password = %v, %v; want true, nil", ok, err)
	}

	ok, err = Verify("wrong password", encoded)
	if err != nil || ok {
		t.Fatalf("Verify with a wrong password = %v, %v; want false, nil", ok, err)
	}
}

func TestHashUsesEncodedFormatWithDefaultParams(t *testing.T) {
	encoded, err := Hash("secret-password")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	if !strings.HasPrefix(encoded, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("unexpected encoded hash: %s", encoded)
	}
}

func TestHashUsesARandomSalt(t *testing.T) {
	first, _ := Hash("same-password")
	second, _ := Hash("same-password")

	if first == second {
		t.Fatal("two hashes of the same password must differ")
	}
}

func TestVerifyRejectsMalformedHashes(t *testing.T) {
	cases := []string{
		"",
		"not-a-hash",
		"$bcrypt$v=19$m=19456,t=2,p=1$c2FsdA$a2V5",
		"$argon2id$v=19$m=x,t=2,p=1$c2FsdA$a2V5",
		"$argon2id$v=19$m=19456,t=2,p=1$!!!$a2V5",
	}

	for _, encoded := range cases {
		if _, err := Verify("password", encoded); !errors.Is(err, ErrInvalidHash) {
			t.Errorf("Verify(%q) error = %v; want ErrInvalidHash", encoded, err)
		}
	}
}

func TestVerifyRejectsUnknownVersion(t *testing.T) {
	_, err := Verify("password", "$argon2id$v=16$m=19456,t=2,p=1$c2FsdA$a2V5")
	if !errors.Is(err, ErrIncompatibleVersion) {
		t.Fatalf("error = %v; want ErrIncompatibleVersion", err)
	}
}
