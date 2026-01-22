package archive

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"strconv"
)

const (
	_         = iota
	kb uint64 = 1 << (10 * iota)
	mb
	gb
	tb
)

// FormatSize returns a human-readable size string.
// Examples: "0", "512", "1.2K", "3.4M", "5.6G", "1.2T"
func FormatSize(bytes uint64) string {
	switch {
	case bytes >= tb:
		return fmt.Sprintf("%.1fT", float64(bytes)/float64(tb))
	case bytes >= gb:
		return fmt.Sprintf("%.1fG", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1fM", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1fK", float64(bytes)/float64(kb))
	default:
		return strconv.FormatUint(bytes, 10)
	}
}

// FormatDigest returns a truncated SHA256 digest string.
// Returns empty string if hash is nil or empty.
// Example: "sha256:abc123def456"
func FormatDigest(hash []byte) string {
	if len(hash) == 0 {
		return ""
	}
	hexStr := hex.EncodeToString(hash)
	if len(hexStr) > 12 {
		hexStr = hexStr[:12]
	}
	return "sha256:" + hexStr
}

// FormatMode returns a Unix-style file mode string.
// Examples: "-rw-r--r--", "drwxr-xr-x"
func FormatMode(mode fs.FileMode, isDir bool) string {
	var buf [10]byte

	// File type indicator
	if isDir {
		buf[0] = 'd'
	} else {
		buf[0] = '-'
	}

	// Owner permissions
	buf[1] = permChar(mode, 0o400, 'r')
	buf[2] = permChar(mode, 0o200, 'w')
	buf[3] = permChar(mode, 0o100, 'x')

	// Group permissions
	buf[4] = permChar(mode, 0o040, 'r')
	buf[5] = permChar(mode, 0o020, 'w')
	buf[6] = permChar(mode, 0o010, 'x')

	// Other permissions
	buf[7] = permChar(mode, 0o004, 'r')
	buf[8] = permChar(mode, 0o002, 'w')
	buf[9] = permChar(mode, 0o001, 'x')

	return string(buf[:])
}

func permChar(mode, bit fs.FileMode, c byte) byte {
	if mode&bit != 0 {
		return c
	}
	return '-'
}
