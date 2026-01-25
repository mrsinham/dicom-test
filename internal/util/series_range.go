// internal/util/series_range.go
package util

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

// SeriesRange represents a range of series per study (min-max)
type SeriesRange struct {
	Min int
	Max int
}

// ParseSeriesRange parses a series range string like "3", "3-5", or "1-10"
func ParseSeriesRange(s string) (SeriesRange, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return SeriesRange{Min: 1, Max: 1}, nil
	}

	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)
		if len(parts) != 2 {
			return SeriesRange{}, fmt.Errorf("invalid series range format: %s (expected N or N-M)", s)
		}

		min, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return SeriesRange{}, fmt.Errorf("invalid series range min: %s", parts[0])
		}

		max, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return SeriesRange{}, fmt.Errorf("invalid series range max: %s", parts[1])
		}

		if min < 1 {
			return SeriesRange{}, fmt.Errorf("series range min must be >= 1, got %d", min)
		}

		if max < min {
			return SeriesRange{}, fmt.Errorf("series range max (%d) must be >= min (%d)", max, min)
		}

		return SeriesRange{Min: min, Max: max}, nil
	}

	// Single number
	n, err := strconv.Atoi(s)
	if err != nil {
		return SeriesRange{}, fmt.Errorf("invalid series count: %s", s)
	}

	if n < 1 {
		return SeriesRange{}, fmt.Errorf("series count must be >= 1, got %d", n)
	}

	return SeriesRange{Min: n, Max: n}, nil
}

// GetSeriesCount returns a random series count within the range
func (r SeriesRange) GetSeriesCount(rng *rand.Rand) int {
	if r.Min == r.Max {
		return r.Min
	}
	return r.Min + rng.IntN(r.Max-r.Min+1)
}

// IsMultiSeries returns true if the range can produce more than 1 series
func (r SeriesRange) IsMultiSeries() bool {
	return r.Max > 1
}

// String returns the string representation of the range
func (r SeriesRange) String() string {
	if r.Min == r.Max {
		return strconv.Itoa(r.Min)
	}
	return fmt.Sprintf("%d-%d", r.Min, r.Max)
}
