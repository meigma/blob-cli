package detect

import (
	"testing"
)

func TestIsBinary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content []byte
		wantBin bool
	}{
		{
			name:    "empty content",
			content: []byte{},
			wantBin: false,
		},
		{
			name:    "plain text",
			content: []byte("Hello, World!\n"),
			wantBin: false,
		},
		{
			name:    "text with tabs and newlines",
			content: []byte("Line 1\tColumn\nLine 2\tColumn\r\n"),
			wantBin: false,
		},
		{
			name:    "JSON content",
			content: []byte(`{"key": "value", "number": 42}`),
			wantBin: false,
		},
		{
			name:    "null byte makes it binary",
			content: []byte("Hello\x00World"),
			wantBin: true,
		},
		{
			name:    "binary with null at start",
			content: []byte{0x00, 0x01, 0x02, 0x03},
			wantBin: true,
		},
		{
			name:    "PNG header",
			content: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantBin: true,
		},
		{
			name:    "invalid UTF-8",
			content: []byte{0xFF, 0xFE, 0x00, 0x01},
			wantBin: true,
		},
		{
			name:    "many control characters",
			content: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0E, 0x0F},
			wantBin: true,
		},
		{
			name:    "valid UTF-8 with unicode",
			content: []byte("Hello 世界 🌍"),
			wantBin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsBinary(tt.content)
			if got != tt.wantBin {
				t.Errorf("IsBinary() = %v, want %v", got, tt.wantBin)
			}
		})
	}
}

func TestIsBinary_LargeContent(t *testing.T) {
	t.Parallel()

	// Create content larger than maxScanSize (8KB)
	largeText := make([]byte, 16*1024)
	for i := range largeText {
		largeText[i] = 'a'
	}

	// Should not be binary (all 'a' characters)
	if IsBinary(largeText) {
		t.Error("large text content should not be detected as binary")
	}

	// Put a null byte after the scan window - should NOT be detected
	largeText[10*1024] = 0x00
	if IsBinary(largeText) {
		t.Error("null byte after scan window should not trigger binary detection")
	}

	// Put a null byte within the scan window - should be detected
	largeText[4*1024] = 0x00
	if !IsBinary(largeText) {
		t.Error("null byte within scan window should trigger binary detection")
	}
}
