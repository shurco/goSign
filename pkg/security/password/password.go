package password

import (
	"crypto/md5"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the bcrypt cost used for hashing user passwords.
// Kept configurable (not hardcoded) to simplify tuning.
const DefaultCost = bcrypt.MinCost

// GeneratePassword returns a bcrypt hash of the plain-text password.
// It returns an empty string and a non-nil error on failure (e.g. input longer
// than 72 bytes) so that callers cannot mistake an error message for a hash.
func GeneratePassword(p string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePasswords reports whether inputPwd matches hashedPwd.
func ComparePasswords(hashedPwd, inputPwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(inputPwd)) == nil
}

// NewToken returns a short, hex-encoded MD5 digest of a bcrypt hash of text.
// It is used as a one-shot opaque token (NOT for authenticating passwords).
func NewToken(text string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	digest := md5.Sum(hash)
	return hex.EncodeToString(digest[:]), nil
}
