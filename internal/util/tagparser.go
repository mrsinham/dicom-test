package util

import (
	"fmt"
	"strings"
)

// ParsedTags represents a map of tag names to their values.
// Tag names are stored in their canonical form (as defined in the registry).
type ParsedTags map[string]string

// ParseTagFlags parses a slice of tag flags in the format "TagName=Value".
// It validates each tag name against the registry and returns an error
// if an unknown tag is encountered.
// The returned ParsedTags map uses canonical tag names as keys.
func ParseTagFlags(flags []string) (ParsedTags, error) {
	result := make(ParsedTags)

	for _, flag := range flags {
		// Find the first '=' to split tag name and value
		idx := strings.Index(flag, "=")
		if idx == -1 {
			return nil, fmt.Errorf("invalid tag format %q: missing '=' (expected TagName=Value)", flag)
		}

		tagName := strings.TrimSpace(flag[:idx])
		value := flag[idx+1:] // Value after the '=' (can be empty, preserves spaces)

		if tagName == "" {
			return nil, fmt.Errorf("invalid tag format %q: empty tag name", flag)
		}

		// Validate and get canonical name from registry
		tagInfo, err := GetTagByName(tagName)
		if err != nil {
			return nil, err
		}

		// Store with canonical name
		result[tagInfo.Name] = value
	}

	return result, nil
}

// Has returns true if the tag with the given name exists in the parsed tags.
func (pt ParsedTags) Has(name string) bool {
	_, ok := pt[name]
	return ok
}

// Get returns the value for the given tag name and a boolean indicating
// whether the tag exists. Returns empty string and false if not found.
func (pt ParsedTags) Get(name string) (string, bool) {
	value, ok := pt[name]
	return value, ok
}

// GetWithScope returns a new ParsedTags containing only the tags
// that match the specified scope.
func (pt ParsedTags) GetWithScope(scope TagScope) ParsedTags {
	result := make(ParsedTags)

	for name, value := range pt {
		// Look up the tag info to get its scope
		tagInfo, err := GetTagByName(name)
		if err != nil {
			// Skip tags that aren't in the registry (shouldn't happen if parsed correctly)
			continue
		}

		if tagInfo.Scope == scope {
			result[name] = value
		}
	}

	return result
}

// Keys returns a slice of all tag names in the parsed tags.
// The order is not guaranteed.
func (pt ParsedTags) Keys() []string {
	keys := make([]string, 0, len(pt))
	for k := range pt {
		keys = append(keys, k)
	}
	return keys
}
