package preview

import (
	"strings"
	"testing"
)

func TestFormatHex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  []byte
		contains []string
		notEmpty bool
	}{
		{
			name:     "empty content",
			content:  []byte{},
			contains: []string{"(empty)"},
		},
		{
			name:    "hello world",
			content: []byte("Hello World!"),
			contains: []string{
				"00000000",          // offset
				"48 65 6c 6c 6f 20", // "Hello " in hex
				"|Hello World!",     // ASCII representation
			},
			notEmpty: true,
		},
		{
			name:    "binary content with nulls",
			content: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			contains: []string{
				"00000000",
				"00 01 02 03 04 05",
				"|......", // non-printable shown as dots (with padding after)
			},
			notEmpty: true,
		},
		{
			name:    "printable and non-printable mix",
			content: []byte("AB\x00CD\x7FEF"),
			contains: []string{
				"41 42 00 43 44 7f 45 46", // hex bytes
				"|AB.CD.EF",               // ASCII with dots for non-printable
			},
			notEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatHex(tt.content)

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("FormatHex() missing %q in output:\n%s", want, got)
				}
			}

			if tt.notEmpty && got == "" {
				t.Error("FormatHex() returned empty string for non-empty content")
			}
		})
	}
}

func TestFormatHex_MultiLine(t *testing.T) {
	t.Parallel()

	// 32 bytes = 2 lines of output
	content := make([]byte, 32)
	for i := range content {
		content[i] = byte(i)
	}

	got := FormatHex(content)
	lines := strings.Split(strings.TrimSpace(got), "\n")

	if len(lines) != 2 {
		t.Errorf("FormatHex() got %d lines, want 2", len(lines))
	}

	// First line should start with offset 00000000
	if !strings.HasPrefix(lines[0], "00000000") {
		t.Errorf("first line should start with 00000000, got: %s", lines[0])
	}

	// Second line should start with offset 00000010 (16 in hex)
	if !strings.HasPrefix(lines[1], "00000010") {
		t.Errorf("second line should start with 00000010, got: %s", lines[1])
	}
}

func TestFormatHex_PartialLine(t *testing.T) {
	t.Parallel()

	// 20 bytes = 1 full line + 4 bytes partial
	content := make([]byte, 20)
	for i := range content {
		content[i] = byte('A' + i)
	}

	got := FormatHex(content)
	lines := strings.Split(strings.TrimSpace(got), "\n")

	if len(lines) != 2 {
		t.Errorf("FormatHex() got %d lines, want 2", len(lines))
	}

	// Second line should have padding for missing bytes
	// The ASCII section should still have the closing pipe aligned
	if !strings.Contains(lines[1], "|") {
		t.Error("partial line should contain ASCII section with pipes")
	}
}
