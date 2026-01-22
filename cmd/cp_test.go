package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

func TestCpCmd_NilConfig(t *testing.T) {
	viper.Reset()

	ctx := context.Background()

	cpCmd.SetContext(ctx)
	err := cpCmd.RunE(cpCmd, []string{"ghcr.io/test:v1:/config.json", "./dest"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestCpCmd_MinimumArgs(t *testing.T) {
	// Verify command requires at least 2 args (source + dest)
	assert.Equal(t, "cp <ref>:<path>... <dest>", cpCmd.Use)

	// Cobra's MinimumNArgs(2) is set
	err := cpCmd.Args(cpCmd, []string{"only-one-arg"})
	require.Error(t, err)

	err = cpCmd.Args(cpCmd, []string{"source", "dest"})
	require.NoError(t, err)

	err = cpCmd.Args(cpCmd, []string{"src1", "src2", "dest"})
	require.NoError(t, err)
}

func TestParseSourceArg(t *testing.T) {
	cfg := &internalcfg.Config{
		Aliases: map[string]string{
			"myalias": "ghcr.io/acme/repo",
		},
	}

	tests := []struct {
		name         string
		arg          string
		wantRef      string
		wantPath     string
		wantInputRef string
		wantErr      string
	}{
		{
			name:         "standard ref with tag",
			arg:          "ghcr.io/acme/repo:v1.0.0:/config.json",
			wantRef:      "ghcr.io/acme/repo:v1.0.0",
			wantPath:     "/config.json",
			wantInputRef: "ghcr.io/acme/repo:v1.0.0",
		},
		{
			name:         "ref with registry port",
			arg:          "registry:5000/repo:v1:/path/to/file",
			wantRef:      "registry:5000/repo:v1",
			wantPath:     "/path/to/file",
			wantInputRef: "registry:5000/repo:v1",
		},
		{
			name:         "alias resolution",
			arg:          "myalias:/config.json",
			wantRef:      "ghcr.io/acme/repo:latest",
			wantPath:     "/config.json",
			wantInputRef: "myalias",
		},
		{
			name:         "alias with tag override",
			arg:          "myalias:v2:/config.json",
			wantRef:      "ghcr.io/acme/repo:v2",
			wantPath:     "/config.json",
			wantInputRef: "myalias:v2",
		},
		{
			name:         "directory path with trailing slash",
			arg:          "ghcr.io/acme/repo:v1:/etc/nginx/",
			wantRef:      "ghcr.io/acme/repo:v1",
			wantPath:     "/etc/nginx/",
			wantInputRef: "ghcr.io/acme/repo:v1",
		},
		{
			name:         "root path",
			arg:          "ghcr.io/acme/repo:v1:/",
			wantRef:      "ghcr.io/acme/repo:v1",
			wantPath:     "/",
			wantInputRef: "ghcr.io/acme/repo:v1",
		},
		{
			name:         "directory path without trailing slash",
			arg:          "ghcr.io/acme/repo:v1:/etc/nginx",
			wantRef:      "ghcr.io/acme/repo:v1",
			wantPath:     "/etc/nginx",
			wantInputRef: "ghcr.io/acme/repo:v1",
		},
		{
			name:    "missing path separator",
			arg:     "ghcr.io/acme/repo:v1",
			wantErr: "invalid source format",
		},
		{
			name:    "empty reference",
			arg:     ":/config.json",
			wantErr: "reference cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := parseSourceArg(tt.arg, cfg)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantRef, src.ref)
			assert.Equal(t, tt.wantPath, src.path)
			assert.Equal(t, tt.wantInputRef, src.inputRef)
		})
	}
}

func TestGetDestInfo(t *testing.T) {
	tmpDir := t.TempDir()
	existingDir := filepath.Join(tmpDir, "existing-dir")
	require.NoError(t, os.MkdirAll(existingDir, 0o755))
	existingFile := filepath.Join(tmpDir, "existing-file")
	require.NoError(t, os.WriteFile(existingFile, []byte("test"), 0o644))

	tests := []struct {
		name          string
		dest          string
		wantExists    bool
		wantIsDir     bool
		wantEndsSlash bool
	}{
		{
			name:       "existing directory",
			dest:       existingDir,
			wantExists: true,
			wantIsDir:  true,
		},
		{
			name:       "existing file",
			dest:       existingFile,
			wantExists: true,
			wantIsDir:  false,
		},
		{
			name:       "non-existent path",
			dest:       filepath.Join(tmpDir, "new-file"),
			wantExists: false,
			wantIsDir:  false,
		},
		{
			name:          "path ending with slash",
			dest:          filepath.Join(tmpDir, "newdir") + "/",
			wantExists:    false,
			wantEndsSlash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			di, err := getDestInfo(tt.dest)
			require.NoError(t, err)
			assert.Equal(t, tt.wantExists, di.exists)
			assert.Equal(t, tt.wantIsDir, di.isDir)
			assert.Equal(t, tt.wantEndsSlash, di.endsWithSlash)
		})
	}
}

func TestValidateAndPrepareDestination(t *testing.T) {
	tmpDir := t.TempDir()
	existingDir := filepath.Join(tmpDir, "existing-dir")
	require.NoError(t, os.MkdirAll(existingDir, 0o755))
	existingFile := filepath.Join(tmpDir, "existing-file")
	require.NoError(t, os.WriteFile(existingFile, []byte("test"), 0o644))

	tests := []struct {
		name    string
		sources []cpResolvedSource
		dest    string
		flags   cpFlags
		wantErr string
	}{
		{
			name:    "single file to non-existent path",
			sources: []cpResolvedSource{{isDir: false, cpSource: cpSource{path: "/config.json"}}},
			dest:    filepath.Join(tmpDir, "new-file.json"),
			flags:   cpFlags{recursive: true},
		},
		{
			name:    "single file to existing directory",
			sources: []cpResolvedSource{{isDir: false, cpSource: cpSource{path: "/config.json"}}},
			dest:    existingDir,
			flags:   cpFlags{recursive: true},
		},
		{
			name:    "single file to path ending with slash",
			sources: []cpResolvedSource{{isDir: false, cpSource: cpSource{path: "/config.json"}}},
			dest:    filepath.Join(tmpDir, "newdir") + "/",
			flags:   cpFlags{recursive: true},
		},
		{
			name:    "directory source with recursive flag",
			sources: []cpResolvedSource{{isDir: true, cpSource: cpSource{path: "/etc/nginx"}}},
			dest:    filepath.Join(tmpDir, "nginx-config"),
			flags:   cpFlags{recursive: true},
		},
		{
			name:    "directory source without recursive flag",
			sources: []cpResolvedSource{{isDir: true, cpSource: cpSource{path: "/etc/nginx"}}},
			dest:    filepath.Join(tmpDir, "nginx-config"),
			flags:   cpFlags{recursive: false},
			wantErr: "without -r flag",
		},
		{
			name: "multiple sources to directory",
			sources: []cpResolvedSource{
				{isDir: false, cpSource: cpSource{path: "/a.json"}},
				{isDir: false, cpSource: cpSource{path: "/b.json"}},
			},
			dest:  existingDir,
			flags: cpFlags{recursive: true},
		},
		{
			name: "multiple sources to non-existent directory",
			sources: []cpResolvedSource{
				{isDir: false, cpSource: cpSource{path: "/a.json"}},
				{isDir: false, cpSource: cpSource{path: "/b.json"}},
			},
			dest:  filepath.Join(tmpDir, "new-dir"),
			flags: cpFlags{recursive: true},
		},
		{
			name: "multiple sources to existing file",
			sources: []cpResolvedSource{
				{isDir: false, cpSource: cpSource{path: "/a.json"}},
				{isDir: false, cpSource: cpSource{path: "/b.json"}},
			},
			dest:    existingFile,
			flags:   cpFlags{recursive: true},
			wantErr: "destination must be a directory",
		},
		{
			name: "mixed file and directory sources",
			sources: []cpResolvedSource{
				{isDir: false, cpSource: cpSource{path: "/config.json"}},
				{isDir: true, cpSource: cpSource{path: "/etc/nginx"}},
			},
			dest:  filepath.Join(tmpDir, "mixed-dest"),
			flags: cpFlags{recursive: true},
		},
		{
			name:    "directory source to existing file",
			sources: []cpResolvedSource{{isDir: true, cpSource: cpSource{path: "/etc/nginx"}}},
			dest:    existingFile,
			flags:   cpFlags{recursive: true},
			wantErr: "cannot copy directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := validateAndPrepareDestination(tt.sources, tt.dest, tt.flags)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, path)
		})
	}
}

func TestCpFlags(t *testing.T) {
	// Reset flags for testing
	cpCmd.Flags().Set("recursive", "false")
	cpCmd.Flags().Set("preserve", "true")
	cpCmd.Flags().Set("force", "true")

	flags, err := parseCpFlags(cpCmd)
	require.NoError(t, err)
	assert.False(t, flags.recursive)
	assert.True(t, flags.preserve)
	assert.True(t, flags.force)

	// Reset to defaults
	cpCmd.Flags().Set("recursive", "true")
	cpCmd.Flags().Set("preserve", "false")
	cpCmd.Flags().Set("force", "false")

	flags, err = parseCpFlags(cpCmd)
	require.NoError(t, err)
	assert.True(t, flags.recursive)
	assert.False(t, flags.preserve)
	assert.False(t, flags.force)
}

func TestCpJSON(t *testing.T) {
	result := &cpResult{
		Sources: []cpSourceResult{
			{Ref: "ghcr.io/test:v1", Path: "/config.json"},
		},
		Destination: "/tmp/dest",
		FileCount:   1,
		TotalSize:   1024,
		SizeHuman:   "1.0K",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cpJSON(result)

	w.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)

	var got cpResult
	err = json.NewDecoder(r).Decode(&got)
	require.NoError(t, err)

	assert.Equal(t, result.Destination, got.Destination)
	assert.Equal(t, result.FileCount, got.FileCount)
	assert.Equal(t, result.TotalSize, got.TotalSize)
	require.Len(t, got.Sources, 1)
	assert.Equal(t, "/config.json", got.Sources[0].Path)
}

func TestBuildCopyOpts(t *testing.T) {
	// Without preserve, without force
	flags := cpFlags{recursive: true, preserve: false, force: false}
	opts := buildCopyOpts(flags)
	assert.Len(t, opts, 1) // Only overwrite option (set to false)

	// With preserve
	flags = cpFlags{recursive: true, preserve: true, force: false}
	opts = buildCopyOpts(flags)
	assert.Len(t, opts, 3) // overwrite + mode + times

	// With force
	flags = cpFlags{recursive: true, preserve: false, force: true}
	opts = buildCopyOpts(flags)
	assert.Len(t, opts, 1) // overwrite option (set to true)
}
