// internal/util/clinical_test.go
package util

import (
	"math/rand/v2"
	"testing"
)

func TestGetBodyPartsForModality_MR(t *testing.T) {
	parts := GetBodyPartsForModality("MR")
	if len(parts) == 0 {
		t.Error("MR should have body parts")
	}

	// Should contain HEAD for MR
	found := false
	for _, p := range parts {
		if p == "HEAD" {
			found = true
			break
		}
	}
	if !found {
		t.Error("MR should include HEAD body part")
	}
}

func TestGetBodyPartsForModality_Unknown(t *testing.T) {
	parts := GetBodyPartsForModality("UNKNOWN")
	if len(parts) == 0 {
		t.Error("Unknown modality should return default body parts")
	}
}

func TestGenerateProtocolName_Format(t *testing.T) {
	protocol := GenerateProtocolName("MR", "HEAD", nil)
	if protocol == "" {
		t.Error("Protocol name should not be empty")
	}
}

func TestGenerateProtocolName_Deterministic(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(42, 42))
	p1 := GenerateProtocolName("MR", "HEAD", rng1)

	rng2 := rand.New(rand.NewPCG(42, 42))
	p2 := GenerateProtocolName("MR", "HEAD", rng2)

	if p1 != p2 {
		t.Errorf("Same seed should produce same protocol: %s != %s", p1, p2)
	}
}

func TestGenerateClinicalIndication_NotEmpty(t *testing.T) {
	indication := GenerateClinicalIndication("MR", "HEAD", nil)
	if indication == "" {
		t.Error("Clinical indication should not be empty")
	}
}
