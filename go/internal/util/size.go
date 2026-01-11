package util

import (
	"fmt"
	"regexp"
	"strconv"
)

// ParseSize parses a size string (e.g., "4.5GB", "100MB") into bytes.
//
// Supported units: KB, MB, GB
// Returns the size in bytes or an error if the format is invalid.
func ParseSize(sizeStr string) (int64, error) {
	pattern := `^(\d+(?:\.\d+)?)(KB|MB|GB)$`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(sizeStr)

	if matches == nil {
		return 0, fmt.Errorf("invalid format: '%s'. Use format like '100MB', '4.5GB'", sizeStr)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %v", err)
	}

	unit := matches[2]
	multipliers := map[string]int64{
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	multiplier, ok := multipliers[unit]
	if !ok {
		return 0, fmt.Errorf("unsupported unit: '%s'. Use KB, MB, or GB", unit)
	}

	return int64(value * float64(multiplier)), nil
}
