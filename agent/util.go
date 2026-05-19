package agent

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"strings"
	"time"
)

const (
	dateSearchFormat = "1/2/2006,03:04:05 PM"
	dateFormat       = "1/2/2006 15:04:05"
)

// FDIDToORI converts an FDID to a CAD system internal ORI used for searching.
// It requires an ORIObj array.
func FDIDToORI(orimap []ORIObj, fdid string) string {
	for _, ori := range orimap {
		if ori.FDID == fdid {
			return ori.ORI
		}
	}
	return ""
}

func parseDate(dt string) time.Time {
	t, err := time.Parse(dateFormat, dt)
	if err != nil {
		log.Printf("parseDate: %s could not be parsed, using now()", dt)
		return time.Now()
	}
	return t
}

// generateCodeVerifier produces a base64url-encoded random 32-byte value
// suitable for PKCE code_verifier.
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge computes the S256 PKCE challenge from a verifier:
// base64url(sha256(ascii(verifier)))
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// generateState produces a random hex string for OIDC state CSRF protection.
func generateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateNonce produces a random hex string for OIDC nonce replay protection.
func generateNonce() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// unwantedTraffic determines if a URL should be stored in memory or not
func unwantedTraffic(url string) bool {
	return !strings.HasPrefix(url, "http") ||
		strings.HasSuffix(url, ".css") ||
		strings.HasSuffix(url, ".js") ||
		strings.HasSuffix(url, ".svg") ||
		strings.HasSuffix(url, ".woff2")
}
