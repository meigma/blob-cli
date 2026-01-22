package archive

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		bytes uint64
		want  string
	}{
		{name: "zero", bytes: 0, want: "0"},
		{name: "small", bytes: 512, want: "512"},
		{name: "one_kb", bytes: 1024, want: "1.0K"},
		{name: "kb_with_fraction", bytes: 1536, want: "1.5K"},
		{name: "one_mb", bytes: 1024 * 1024, want: "1.0M"},
		{name: "mb_with_fraction", bytes: 1536 * 1024, want: "1.5M"},
		{name: "one_gb", bytes: 1024 * 1024 * 1024, want: "1.0G"},
		{name: "gb_with_fraction", bytes: 1536 * 1024 * 1024, want: "1.5G"},
		{name: "one_tb", bytes: 1024 * 1024 * 1024 * 1024, want: "1.0T"},
		{name: "tb_with_fraction", bytes: 1536 * 1024 * 1024 * 1024, want: "1.5T"},
		{name: "just_under_kb", bytes: 1023, want: "1023"},
		{name: "just_under_mb", bytes: 1024*1024 - 1, want: "1024.0K"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatSize(tt.bytes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatDigest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		hash []byte
		want string
	}{
		{name: "nil", hash: nil, want: ""},
		{name: "empty", hash: []byte{}, want: ""},
		{
			name: "valid_sha256",
			hash: []byte{0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78, 0x9a},
			want: "sha256:abcdef123456",
		},
		{
			name: "short_hash",
			hash: []byte{0xab, 0xcd},
			want: "sha256:abcd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatDigest(tt.hash)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		mode  fs.FileMode
		isDir bool
		want  string
	}{
		{name: "regular_file_644", mode: 0o644, isDir: false, want: "-rw-r--r--"},
		{name: "regular_file_755", mode: 0o755, isDir: false, want: "-rwxr-xr-x"},
		{name: "regular_file_600", mode: 0o600, isDir: false, want: "-rw-------"},
		{name: "directory_755", mode: 0o755, isDir: true, want: "drwxr-xr-x"},
		{name: "directory_700", mode: 0o700, isDir: true, want: "drwx------"},
		{name: "no_permissions", mode: 0o000, isDir: false, want: "----------"},
		{name: "all_permissions", mode: 0o777, isDir: false, want: "-rwxrwxrwx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatMode(tt.mode, tt.isDir)
			assert.Equal(t, tt.want, got)
		})
	}
}
