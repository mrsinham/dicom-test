package util

import (
	"strings"
	"testing"
)

func TestParseTagFlags_Valid(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		expected map[string]string
	}{
		{
			name:  "single tag",
			flags: []string{"InstitutionName=CHU Bordeaux"},
			expected: map[string]string{
				"InstitutionName": "CHU Bordeaux",
			},
		},
		{
			name:  "multiple tags",
			flags: []string{"PatientName=John Doe", "PatientID=12345"},
			expected: map[string]string{
				"PatientName": "John Doe",
				"PatientID":   "12345",
			},
		},
		{
			name:  "value with equals sign",
			flags: []string{"StudyDescription=A=B Test"},
			expected: map[string]string{
				"StudyDescription": "A=B Test",
			},
		},
		{
			name:  "value with special characters",
			flags: []string{"Manufacturer=GE Healthcare (2024)"},
			expected: map[string]string{
				"Manufacturer": "GE Healthcare (2024)",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseTagFlags(tc.flags)
			if err != nil {
				t.Fatalf("ParseTagFlags(%v) returned error: %v", tc.flags, err)
			}

			if len(parsed) != len(tc.expected) {
				t.Errorf("ParseTagFlags(%v) returned %d tags, want %d", tc.flags, len(parsed), len(tc.expected))
			}

			for key, expectedValue := range tc.expected {
				if gotValue, ok := parsed[key]; !ok {
					t.Errorf("ParseTagFlags(%v) missing key %q", tc.flags, key)
				} else if gotValue != expectedValue {
					t.Errorf("ParseTagFlags(%v)[%q] = %q, want %q", tc.flags, key, gotValue, expectedValue)
				}
			}
		})
	}
}

func TestParseTagFlags_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
	}{
		{
			name:  "missing equals sign",
			flags: []string{"InstitutionName CHU Bordeaux"},
		},
		{
			name:  "just tag name",
			flags: []string{"PatientName"},
		},
		{
			name:  "empty string",
			flags: []string{""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseTagFlags(tc.flags)
			if err == nil {
				t.Errorf("ParseTagFlags(%v) should return error for invalid format", tc.flags)
			}
		})
	}
}

func TestParseTagFlags_UnknownTag(t *testing.T) {
	tests := []struct {
		name       string
		flags      []string
		suggestion string // expected suggestion in error message
	}{
		{
			name:       "unknown tag with suggestion",
			flags:      []string{"PatientNam=John Doe"},
			suggestion: "PatientName",
		},
		{
			name:       "completely unknown tag",
			flags:      []string{"CompletelyInvalidTag=Value"},
			suggestion: "", // no suggestion expected
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseTagFlags(tc.flags)
			if err == nil {
				t.Errorf("ParseTagFlags(%v) should return error for unknown tag", tc.flags)
			}
			if tc.suggestion != "" && !strings.Contains(err.Error(), tc.suggestion) {
				t.Errorf("Error for %v should suggest %q, got: %v", tc.flags, tc.suggestion, err)
			}
		})
	}
}

func TestParseTagFlags_EmptyValue(t *testing.T) {
	flags := []string{"PatientName="}
	parsed, err := ParseTagFlags(flags)
	if err != nil {
		t.Fatalf("ParseTagFlags(%v) returned error: %v", flags, err)
	}

	if value := parsed["PatientName"]; value != "" {
		t.Errorf("ParseTagFlags(%v)[PatientName] = %q, want empty string", flags, value)
	}
}

func TestParseTagFlags_CaseNormalization(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		canonicalName string
	}{
		{"lowercase", "patientname=John", "PatientName"},
		{"uppercase", "PATIENTNAME=John", "PatientName"},
		{"mixed case", "pAtIeNtNaMe=John", "PatientName"},
		{"correct case", "PatientName=John", "PatientName"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseTagFlags([]string{tc.input})
			if err != nil {
				t.Fatalf("ParseTagFlags([%q]) returned error: %v", tc.input, err)
			}

			// The key should be the canonical name
			if _, ok := parsed[tc.canonicalName]; !ok {
				t.Errorf("ParseTagFlags([%q]) should use canonical name %q as key, got keys: %v",
					tc.input, tc.canonicalName, parsed.Keys())
			}
		})
	}
}

func TestParsedTags_Has(t *testing.T) {
	parsed := ParsedTags{
		"PatientName": "John Doe",
		"PatientID":   "12345",
	}

	tests := []struct {
		name     string
		expected bool
	}{
		{"PatientName", true},
		{"PatientID", true},
		{"StudyDescription", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := parsed.Has(tc.name); got != tc.expected {
				t.Errorf("ParsedTags.Has(%q) = %v, want %v", tc.name, got, tc.expected)
			}
		})
	}
}

func TestParsedTags_Get(t *testing.T) {
	parsed := ParsedTags{
		"PatientName": "John Doe",
		"PatientID":   "",
	}

	tests := []struct {
		name     string
		expected string
		exists   bool
	}{
		{"PatientName", "John Doe", true},
		{"PatientID", "", true},
		{"StudyDescription", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, exists := parsed.Get(tc.name)
			if got != tc.expected {
				t.Errorf("ParsedTags.Get(%q) value = %q, want %q", tc.name, got, tc.expected)
			}
			if exists != tc.exists {
				t.Errorf("ParsedTags.Get(%q) exists = %v, want %v", tc.name, exists, tc.exists)
			}
		})
	}
}

func TestParsedTags_GetWithScope(t *testing.T) {
	parsed := ParsedTags{
		"PatientName":      "John Doe",
		"PatientID":        "12345",
		"StudyDescription": "CT Scan",
		"InstitutionName":  "Hospital",
		"SeriesDescription": "Axial",
		"WindowCenter":     "40",
	}

	tests := []struct {
		scope    TagScope
		expected map[string]string
	}{
		{
			scope: ScopePatient,
			expected: map[string]string{
				"PatientName": "John Doe",
				"PatientID":   "12345",
			},
		},
		{
			scope: ScopeStudy,
			expected: map[string]string{
				"StudyDescription": "CT Scan",
				"InstitutionName":  "Hospital",
			},
		},
		{
			scope: ScopeSeries,
			expected: map[string]string{
				"SeriesDescription": "Axial",
			},
		},
		{
			scope: ScopeImage,
			expected: map[string]string{
				"WindowCenter": "40",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.scope.String(), func(t *testing.T) {
			got := parsed.GetWithScope(tc.scope)

			if len(got) != len(tc.expected) {
				t.Errorf("GetWithScope(%v) returned %d tags, want %d", tc.scope, len(got), len(tc.expected))
			}

			for key, expectedValue := range tc.expected {
				if gotValue, ok := got[key]; !ok {
					t.Errorf("GetWithScope(%v) missing key %q", tc.scope, key)
				} else if gotValue != expectedValue {
					t.Errorf("GetWithScope(%v)[%q] = %q, want %q", tc.scope, key, gotValue, expectedValue)
				}
			}
		})
	}
}

func TestParsedTags_Keys(t *testing.T) {
	parsed := ParsedTags{
		"PatientName": "John Doe",
		"PatientID":   "12345",
	}

	keys := parsed.Keys()
	if len(keys) != 2 {
		t.Errorf("ParsedTags.Keys() returned %d keys, want 2", len(keys))
	}

	// Check that both keys are present (order doesn't matter)
	hasPatientName := false
	hasPatientID := false
	for _, k := range keys {
		if k == "PatientName" {
			hasPatientName = true
		}
		if k == "PatientID" {
			hasPatientID = true
		}
	}

	if !hasPatientName {
		t.Error("ParsedTags.Keys() missing PatientName")
	}
	if !hasPatientID {
		t.Error("ParsedTags.Keys() missing PatientID")
	}
}

func TestParseTagFlags_EmptySlice(t *testing.T) {
	parsed, err := ParseTagFlags([]string{})
	if err != nil {
		t.Fatalf("ParseTagFlags([]) returned error: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("ParseTagFlags([]) returned %d tags, want 0", len(parsed))
	}
}

func TestParseTagFlags_NilSlice(t *testing.T) {
	parsed, err := ParseTagFlags(nil)
	if err != nil {
		t.Fatalf("ParseTagFlags(nil) returned error: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("ParseTagFlags(nil) returned %d tags, want 0", len(parsed))
	}
}
