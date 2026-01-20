package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// GenerateDeterministicUID generates a deterministic DICOM UID from a seed string.
//
// The UID is generated using SHA256 hash of the seed, ensuring the same seed
// always produces the same UID. The result is a valid DICOM UID (max 64 chars,
// no leading zeros in components).
func GenerateDeterministicUID(seed string) string {
	// DICOM UID prefix for compatibility
	prefix := "1.2.826.0.1.3680043.8.498"

	// Generate SHA256 hash of seed
	hash := sha256.Sum256([]byte(seed))
	hashHex := hex.EncodeToString(hash[:])

	// Convert first 30 hex chars to numeric string
	hashBytes := hashHex[:30]
	numericValue := new(big.Int)
	numericValue.SetString(hashBytes, 16)
	numericSuffix := numericValue.String()

	// Create segments, ensuring no segment starts with 0
	var segments []string
	for i := 0; i < len(numericSuffix) && len(segments) < 3; i += 10 {
		end := i + 10
		if end > len(numericSuffix) {
			end = len(numericSuffix)
		}
		segment := numericSuffix[i:end]

		// Remove leading zeros (unless segment is just "0")
		if segment != "0" && len(segment) > 0 && segment[0] == '0' {
			segment = strings.TrimLeft(segment, "0")
			if segment == "" {
				segment = "1"
			}
		}

		if segment != "" {
			segments = append(segments, segment)
		}
	}

	suffix := strings.Join(segments, ".")
	uid := fmt.Sprintf("%s.%s", prefix, suffix)

	// Ensure UID is not too long (max 64 chars)
	if len(uid) > 64 {
		uid = uid[:63]
		uid = strings.TrimSuffix(uid, ".")
	}

	return uid
}
