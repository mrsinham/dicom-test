// internal/util/series_range_test.go
package util

import (
	"math/rand/v2"
	"testing"
)

func TestParseSeriesRange_SingleNumber(t *testing.T) {
	tests := []struct {
		input   string
		wantMin int
		wantMax int
	}{
		{"1", 1, 1},
		{"3", 3, 3},
		{"10", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r, err := ParseSeriesRange(tt.input)
			if err != nil {
				t.Fatalf("ParseSeriesRange(%q) failed: %v", tt.input, err)
			}
			if r.Min != tt.wantMin || r.Max != tt.wantMax {
				t.Errorf("ParseSeriesRange(%q) = {%d, %d}, want {%d, %d}", tt.input, r.Min, r.Max, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestParseSeriesRange_Range(t *testing.T) {
	tests := []struct {
		input   string
		wantMin int
		wantMax int
	}{
		{"1-3", 1, 3},
		{"3-5", 3, 5},
		{"1-10", 1, 10},
		{"2 - 4", 2, 4}, // with spaces
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r, err := ParseSeriesRange(tt.input)
			if err != nil {
				t.Fatalf("ParseSeriesRange(%q) failed: %v", tt.input, err)
			}
			if r.Min != tt.wantMin || r.Max != tt.wantMax {
				t.Errorf("ParseSeriesRange(%q) = {%d, %d}, want {%d, %d}", tt.input, r.Min, r.Max, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestParseSeriesRange_Invalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"0", "zero value"},
		{"-1", "negative value"},
		{"5-3", "max < min"},
		{"abc", "not a number"},
		{"1-abc", "invalid max"},
		{"abc-3", "invalid min"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := ParseSeriesRange(tt.input)
			if err == nil {
				t.Errorf("ParseSeriesRange(%q) expected error for %s", tt.input, tt.desc)
			}
		})
	}
}

func TestParseSeriesRange_Empty(t *testing.T) {
	r, err := ParseSeriesRange("")
	if err != nil {
		t.Fatalf("ParseSeriesRange(\"\") failed: %v", err)
	}
	if r.Min != 1 || r.Max != 1 {
		t.Errorf("ParseSeriesRange(\"\") = {%d, %d}, want {1, 1}", r.Min, r.Max)
	}
}

func TestSeriesRange_GetSeriesCount(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))

	// Fixed range
	fixed := SeriesRange{Min: 3, Max: 3}
	for i := 0; i < 10; i++ {
		if c := fixed.GetSeriesCount(rng); c != 3 {
			t.Errorf("Fixed range GetSeriesCount() = %d, want 3", c)
		}
	}

	// Variable range - verify all values are in range
	variable := SeriesRange{Min: 2, Max: 5}
	for i := 0; i < 100; i++ {
		c := variable.GetSeriesCount(rng)
		if c < 2 || c > 5 {
			t.Errorf("Variable range GetSeriesCount() = %d, want 2-5", c)
		}
	}
}

func TestSeriesRange_IsMultiSeries(t *testing.T) {
	tests := []struct {
		r    SeriesRange
		want bool
	}{
		{SeriesRange{Min: 1, Max: 1}, false},
		{SeriesRange{Min: 1, Max: 2}, true},
		{SeriesRange{Min: 3, Max: 5}, true},
	}

	for _, tt := range tests {
		if got := tt.r.IsMultiSeries(); got != tt.want {
			t.Errorf("SeriesRange{%d, %d}.IsMultiSeries() = %v, want %v", tt.r.Min, tt.r.Max, got, tt.want)
		}
	}
}

func TestSeriesRange_String(t *testing.T) {
	tests := []struct {
		r    SeriesRange
		want string
	}{
		{SeriesRange{Min: 1, Max: 1}, "1"},
		{SeriesRange{Min: 3, Max: 3}, "3"},
		{SeriesRange{Min: 2, Max: 5}, "2-5"},
	}

	for _, tt := range tests {
		if got := tt.r.String(); got != tt.want {
			t.Errorf("SeriesRange{%d, %d}.String() = %q, want %q", tt.r.Min, tt.r.Max, got, tt.want)
		}
	}
}
