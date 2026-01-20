package util

import (
	"strings"
	"testing"
)

func TestGenerateDeterministicUID_Consistency(t *testing.T) {
	seed := "test_seed_123"
	uid1 := GenerateDeterministicUID(seed)
	uid2 := GenerateDeterministicUID(seed)

	if uid1 != uid2 {
		t.Errorf("Same seed should produce same UID: %s != %s", uid1, uid2)
	}
}

func TestGenerateDeterministicUID_Different(t *testing.T) {
	uid1 := GenerateDeterministicUID("seed1")
	uid2 := GenerateDeterministicUID("seed2")

	if uid1 == uid2 {
		t.Errorf("Different seeds should produce different UIDs")
	}
}

func TestGenerateDeterministicUID_Length(t *testing.T) {
	uid := GenerateDeterministicUID("test_seed")

	if len(uid) > 64 {
		t.Errorf("UID length %d exceeds DICOM maximum of 64 chars", len(uid))
	}
}

func TestGenerateDeterministicUID_NoLeadingZeros(t *testing.T) {
	uid := GenerateDeterministicUID("test_seed")
	segments := strings.Split(uid, ".")

	for i, segment := range segments {
		if len(segment) > 1 && segment[0] == '0' {
			t.Errorf("Segment %d has leading zero: %s", i, segment)
		}
	}
}

func TestGenerateDeterministicUID_Format(t *testing.T) {
	uid := GenerateDeterministicUID("test_seed")

	// Should start with DICOM prefix
	expectedPrefix := "1.2.826.0.1.3680043.8.498."
	if !strings.HasPrefix(uid, expectedPrefix) {
		t.Errorf("UID should start with %s, got %s", expectedPrefix, uid)
	}
}
