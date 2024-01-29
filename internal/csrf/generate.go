package csrf

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

// generateToken returns a cryptographically random hex string
func generateToken() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("failed to generate CSRF token: %v", err)
	}
	return hex.EncodeToString(bytes)
}
