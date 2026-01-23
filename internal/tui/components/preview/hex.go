package preview

import (
	"fmt"
	"strings"
)

// bytesPerLine is the number of bytes displayed per line in hex dump.
const bytesPerLine = 16

// FormatHex formats binary content as a hex dump.
// Format: offset  hex-bytes  |ascii|
// Example: 00000000  48 65 6c 6c 6f 20 57 6f 72 6c 64 21 0a 00 00 00  |Hello World!....|
func FormatHex(content []byte) string {
	if len(content) == 0 {
		return "(empty)"
	}

	var sb strings.Builder
	for offset := 0; offset < len(content); offset += bytesPerLine {
		end := min(offset+bytesPerLine, len(content))
		line := content[offset:end]
		sb.WriteString(formatHexLine(offset, line))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// formatHexLine formats a single line of hex dump.
func formatHexLine(offset int, line []byte) string {
	var sb strings.Builder

	// Offset (8 hex digits)
	sb.WriteString(fmt.Sprintf("%08x  ", offset))

	// Hex bytes with spacing
	for i := range bytesPerLine {
		if i < len(line) {
			sb.WriteString(fmt.Sprintf("%02x ", line[i]))
		} else {
			sb.WriteString("   ")
		}
		// Extra space after 8 bytes for readability
		if i == 7 {
			sb.WriteByte(' ')
		}
	}

	// ASCII representation
	sb.WriteString(" |")
	for _, b := range line {
		if b >= 32 && b < 127 {
			sb.WriteByte(b)
		} else {
			sb.WriteByte('.')
		}
	}
	// Pad ASCII section to align closing pipe
	for range bytesPerLine - len(line) {
		sb.WriteByte(' ')
	}
	sb.WriteByte('|')

	return sb.String()
}
