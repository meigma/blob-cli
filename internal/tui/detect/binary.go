// Package detect provides utilities for detecting file content types.
package detect

import (
	"slices"
	"unicode/utf8"
)

// maxScanSize is the maximum number of bytes to scan for binary detection.
const maxScanSize = 8 * 1024 // 8KB

// IsBinary detects if content is binary based on several heuristics:
// - Presence of null bytes
// - Invalid UTF-8 sequences
// - High ratio of control characters
//
// Returns true if the content appears to be binary.
func IsBinary(content []byte) bool {
	if len(content) == 0 {
		return false
	}

	// Only scan the first maxScanSize bytes
	sample := content[:min(len(content), maxScanSize)]

	// Check for null bytes - strong indicator of binary
	if slices.Contains(sample, 0) {
		return true
	}

	// Check if content is valid UTF-8
	if !utf8.Valid(sample) {
		return true
	}

	// Count control characters (except common ones like tab, newline, carriage return)
	controlCount := 0
	for _, b := range sample {
		if isControlChar(b) {
			controlCount++
		}
	}

	// If more than 10% of characters are control characters, treat as binary
	threshold := len(sample) / 10
	return controlCount > threshold
}

// isControlChar returns true if the byte is a control character
// that typically indicates binary content.
// Excludes common text control characters: tab (9), newline (10), carriage return (13).
func isControlChar(b byte) bool {
	// Control characters are 0x00-0x1F and 0x7F
	if b < 32 && b != 9 && b != 10 && b != 13 {
		return true
	}
	if b == 0x7F {
		return true
	}
	return false
}
