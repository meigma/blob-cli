package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/blob-cli/internal/archive"
)

func TestLsCmd_NilConfig(t *testing.T) {
	viper.Reset()

	ctx := context.Background()

	lsCmd.SetContext(ctx)
	err := lsCmd.RunE(lsCmd, []string{"ghcr.io/test:v1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestLsText_Empty(t *testing.T) {
	var entries []*archive.DirEntry
	flags := lsFlags{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := lsText(entries, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	assert.Empty(t, buf.String())
}

func TestLsText_SimpleList(t *testing.T) {
	entries := []*archive.DirEntry{
		{Name: "config", IsDir: true},
		{Name: "README.md", IsDir: false},
	}
	flags := lsFlags{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := lsText(entries, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	assert.Equal(t, "config/\nREADME.md\n", buf.String())
}

func TestLsText_Long(t *testing.T) {
	entries := []*archive.DirEntry{
		{Name: "config", IsDir: true, Mode: fs.ModeDir | 0o755},
		{Name: "file.txt", IsDir: false, Mode: 0o644, Size: 1024},
	}
	flags := lsFlags{long: true}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := lsText(entries, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "drwxr-xr-x")
	assert.Contains(t, output, "-rw-r--r--")
	assert.Contains(t, output, "1024")
	assert.Contains(t, output, "config/")
	assert.Contains(t, output, "file.txt")
}

func TestLsText_LongHuman(t *testing.T) {
	entries := []*archive.DirEntry{
		{Name: "large.bin", IsDir: false, Mode: 0o644, Size: 1536 * 1024}, // 1.5M
	}
	flags := lsFlags{long: true, human: true}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := lsText(entries, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "1.5M")
}

func TestLsText_Digest(t *testing.T) {
	hash := []byte{0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x9a}
	entries := []*archive.DirEntry{
		{Name: "file.txt", IsDir: false, Hash: hash},
	}
	flags := lsFlags{digest: true}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := lsText(entries, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "sha256:abcdef123456")
}

func TestLsJSON(t *testing.T) {
	modTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	hash := []byte{0xab, 0xcd, 0xef, 0x12, 0x34, 0x56}
	entries := []*archive.DirEntry{
		{Name: "config", Path: "config", IsDir: true, Mode: fs.ModeDir | 0o755},
		{Name: "file.txt", Path: "file.txt", IsDir: false, Mode: 0o644, Size: 1024, ModTime: modTime, Hash: hash},
	}
	flags := lsFlags{long: true, human: true, digest: true}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := lsJSON("ghcr.io/test:v1", "/", entries, flags)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)

	var got lsResult
	err = json.Unmarshal(buf.Bytes(), &got)
	require.NoError(t, err)

	assert.Equal(t, "ghcr.io/test:v1", got.Ref)
	assert.Equal(t, "/", got.Path)
	require.Len(t, got.Entries, 2)

	// Check directory entry
	assert.Equal(t, "config", got.Entries[0].Name)
	assert.True(t, got.Entries[0].IsDir)
	assert.Equal(t, "drwxr-xr-x", got.Entries[0].Mode)

	// Check file entry
	assert.Equal(t, "file.txt", got.Entries[1].Name)
	assert.False(t, got.Entries[1].IsDir)
	assert.Equal(t, "-rw-r--r--", got.Entries[1].Mode)
	assert.Equal(t, uint64(1024), got.Entries[1].Size)
	assert.Equal(t, "1.0K", got.Entries[1].SizeHuman)
	assert.Equal(t, "sha256:abcdef123456", got.Entries[1].Digest)
}

func TestFormatEntrySize(t *testing.T) {
	tests := []struct {
		name  string
		size  uint64
		human bool
		want  string
	}{
		{name: "raw", size: 1024, human: false, want: "1024"},
		{name: "human", size: 1024, human: true, want: "1.0K"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatEntrySize(tt.size, tt.human)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatEntryDigest(t *testing.T) {
	tests := []struct {
		name  string
		entry *archive.DirEntry
		want  string
	}{
		{
			name:  "directory",
			entry: &archive.DirEntry{IsDir: true, Hash: []byte{0xab, 0xcd}},
			want:  "",
		},
		{
			name:  "file_no_hash",
			entry: &archive.DirEntry{IsDir: false, Hash: nil},
			want:  "",
		},
		{
			name:  "file_with_hash",
			entry: &archive.DirEntry{IsDir: false, Hash: []byte{0xab, 0xcd, 0xef, 0x12, 0x34, 0x56}},
			want:  "sha256:abcdef123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatEntryDigest(tt.entry)
			assert.Equal(t, tt.want, got)
		})
	}
}
