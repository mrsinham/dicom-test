// internal/util/institutions_test.go
package util

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestGenerateInstitution_ReturnsValidData(t *testing.T) {
	inst := GenerateInstitution(nil)

	if inst.Name == "" {
		t.Error("Institution name should not be empty")
	}
	if inst.Address == "" {
		t.Error("Institution address should not be empty")
	}
	if inst.Department == "" {
		t.Error("Department should not be empty")
	}
}

func TestGenerateInstitution_Deterministic(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(42, 42))
	inst1 := GenerateInstitution(rng1)

	rng2 := rand.New(rand.NewPCG(42, 42))
	inst2 := GenerateInstitution(rng2)

	if inst1.Name != inst2.Name {
		t.Errorf("Same seed should produce same institution: %s != %s", inst1.Name, inst2.Name)
	}
}

func TestGenerateStationName_Format(t *testing.T) {
	station := GenerateStationName("MR", "HEAD", nil)

	if station == "" {
		t.Error("Station name should not be empty")
	}
	if !strings.Contains(station, "_") {
		t.Errorf("Station name should contain underscore: %s", station)
	}
}
