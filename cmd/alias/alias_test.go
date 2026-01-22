package alias

import (
	"bytes"
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

func TestListCmd_Empty(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		quiet      bool
		wantStdout string
	}{
		{
			name:       "text output",
			output:     "text",
			wantStdout: "No aliases configured.\n",
		},
		{
			name:   "json output",
			output: "json",
			wantStdout: `{
  "aliases": {}
}
`,
		},
		{
			name:       "text output quiet",
			output:     "text",
			quiet:      true,
			wantStdout: "",
		},
		{
			name:       "json output quiet",
			output:     "json",
			quiet:      true,
			wantStdout: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for each test
			viper.Reset()
			viper.Set("output", tt.output)

			cfg := &internalcfg.Config{
				Aliases: map[string]string{},
				Quiet:   tt.quiet,
			}

			ctx := internalcfg.WithConfig(context.Background(), cfg)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			listCmd.SetContext(ctx)
			err := listCmd.RunE(listCmd, []string{})

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStdout, buf.String())
		})
	}
}

func TestListCmd_WithAliases(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		quiet      bool
		aliases    map[string]string
		wantStdout string
	}{
		{
			name:   "text output single alias",
			output: "text",
			aliases: map[string]string{
				"foo": "ghcr.io/acme/repo/foo",
			},
			wantStdout: `Aliases
--------------------------------------------------
foo  -> ghcr.io/acme/repo/foo
`,
		},
		{
			name:   "text output multiple aliases sorted",
			output: "text",
			aliases: map[string]string{
				"zebra": "ghcr.io/acme/repo/zebra",
				"alpha": "ghcr.io/acme/repo/alpha",
				"beta":  "ghcr.io/acme/repo/beta:v1",
			},
			wantStdout: `Aliases
--------------------------------------------------
alpha  -> ghcr.io/acme/repo/alpha
beta   -> ghcr.io/acme/repo/beta:v1
zebra  -> ghcr.io/acme/repo/zebra
`,
		},
		{
			name:   "json output",
			output: "json",
			aliases: map[string]string{
				"foo": "ghcr.io/acme/repo/foo",
			},
		},
		{
			name:   "quiet suppresses output",
			output: "text",
			quiet:  true,
			aliases: map[string]string{
				"foo": "ghcr.io/acme/repo/foo",
			},
			wantStdout: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("output", tt.output)

			cfg := &internalcfg.Config{
				Aliases: tt.aliases,
				Quiet:   tt.quiet,
			}

			ctx := internalcfg.WithConfig(context.Background(), cfg)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			listCmd.SetContext(ctx)
			err := listCmd.RunE(listCmd, []string{})

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)

			switch {
			case tt.quiet:
				assert.Empty(t, buf.String())
			case tt.output == "json":
				// Parse JSON output and verify structure
				var result map[string]map[string]string
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, tt.aliases, result["aliases"])
			default:
				assert.Equal(t, tt.wantStdout, buf.String())
			}
		})
	}
}

func TestSetCmd(t *testing.T) {
	tests := []struct {
		name           string
		existingAlias  map[string]string
		setName        string
		setRef         string
		output         string
		quiet          bool
		wantAction     string
		wantStdout     string
		wantAliasValue string
	}{
		{
			name:           "create new alias text output",
			existingAlias:  map[string]string{},
			setName:        "foo",
			setRef:         "ghcr.io/acme/repo/foo",
			output:         "text",
			wantAction:     "created",
			wantStdout:     "Created alias \"foo\" -> ghcr.io/acme/repo/foo\n",
			wantAliasValue: "ghcr.io/acme/repo/foo",
		},
		{
			name: "update existing alias text output",
			existingAlias: map[string]string{
				"foo": "ghcr.io/acme/repo/foo:v1",
			},
			setName:        "foo",
			setRef:         "ghcr.io/acme/repo/foo:v2",
			output:         "text",
			wantAction:     "updated",
			wantStdout:     "Updated alias \"foo\" -> ghcr.io/acme/repo/foo:v2\n",
			wantAliasValue: "ghcr.io/acme/repo/foo:v2",
		},
		{
			name:           "create new alias quiet",
			existingAlias:  map[string]string{},
			setName:        "foo",
			setRef:         "ghcr.io/acme/repo/foo",
			output:         "text",
			quiet:          true,
			wantAction:     "created",
			wantStdout:     "",
			wantAliasValue: "ghcr.io/acme/repo/foo",
		},
		{
			name:           "create new alias json output",
			existingAlias:  map[string]string{},
			setName:        "bar",
			setRef:         "ghcr.io/acme/repo/bar",
			output:         "json",
			wantAction:     "created",
			wantAliasValue: "ghcr.io/acme/repo/bar",
		},
		{
			name:           "json output respects quiet",
			existingAlias:  map[string]string{},
			setName:        "baz",
			setRef:         "ghcr.io/acme/repo/baz",
			output:         "json",
			quiet:          true,
			wantAction:     "",
			wantStdout:     "",
			wantAliasValue: "ghcr.io/acme/repo/baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir for config
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			viper.Reset()
			viper.Set("output", tt.output)
			viper.Set("internal.config_path", configPath)

			cfg := &internalcfg.Config{
				Output:      "text",
				Compression: "zstd",
				Aliases:     tt.existingAlias,
				Quiet:       tt.quiet,
			}

			ctx := internalcfg.WithConfig(context.Background(), cfg)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			setCmd.SetContext(ctx)
			err := setCmd.RunE(setCmd, []string{tt.setName, tt.setRef})

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)

			// Verify config file was written
			_, err = os.Stat(configPath)
			require.NoError(t, err, "config file should exist")

			// Verify alias was saved
			savedViper := viper.New()
			savedViper.SetConfigFile(configPath)
			require.NoError(t, savedViper.ReadInConfig())

			savedAliases := savedViper.GetStringMapString("aliases")
			assert.Equal(t, tt.wantAliasValue, savedAliases[tt.setName])

			// Verify output
			switch {
			case tt.quiet:
				assert.Empty(t, buf.String())
			case tt.output == "json":
				var result map[string]string
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, tt.wantAction, result["action"])
				assert.Equal(t, tt.setName, result["name"])
				assert.Equal(t, tt.setRef, result["ref"])
			default:
				assert.Equal(t, tt.wantStdout, buf.String())
			}
		})
	}
}

func TestRemoveCmd(t *testing.T) {
	tests := []struct {
		name          string
		existingAlias map[string]string
		removeName    string
		output        string
		quiet         bool
		wantErr       string
		wantStdout    string
	}{
		{
			name: "remove existing alias text output",
			existingAlias: map[string]string{
				"foo": "ghcr.io/acme/repo/foo",
				"bar": "ghcr.io/acme/repo/bar",
			},
			removeName: "foo",
			output:     "text",
			wantStdout: "Removed alias \"foo\"\n",
		},
		{
			name: "remove existing alias quiet",
			existingAlias: map[string]string{
				"foo": "ghcr.io/acme/repo/foo",
			},
			removeName: "foo",
			output:     "text",
			quiet:      true,
			wantStdout: "",
		},
		{
			name: "remove existing alias json output",
			existingAlias: map[string]string{
				"foo": "ghcr.io/acme/repo/foo",
			},
			removeName: "foo",
			output:     "json",
		},
		{
			name:          "remove non-existent alias",
			existingAlias: map[string]string{},
			removeName:    "nonexistent",
			output:        "text",
			wantErr:       `alias "nonexistent" not found`,
		},
		{
			name: "json output respects quiet",
			existingAlias: map[string]string{
				"foo": "ghcr.io/acme/repo/foo",
			},
			removeName: "foo",
			output:     "json",
			quiet:      true,
			wantStdout: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir for config
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			viper.Reset()
			viper.Set("output", tt.output)
			viper.Set("internal.config_path", configPath)

			cfg := &internalcfg.Config{
				Output:      "text",
				Compression: "zstd",
				Aliases:     tt.existingAlias,
				Quiet:       tt.quiet,
			}

			ctx := internalcfg.WithConfig(context.Background(), cfg)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			removeCmd.SetContext(ctx)
			err := removeCmd.RunE(removeCmd, []string{tt.removeName})

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)

			// Verify config file was written
			_, err = os.Stat(configPath)
			require.NoError(t, err, "config file should exist")

			// Verify alias was removed
			savedViper := viper.New()
			savedViper.SetConfigFile(configPath)
			require.NoError(t, savedViper.ReadInConfig())

			savedAliases := savedViper.GetStringMapString("aliases")
			_, exists := savedAliases[tt.removeName]
			assert.False(t, exists, "alias should have been removed")

			// Verify output
			switch {
			case tt.quiet:
				assert.Empty(t, buf.String())
			case tt.output == "json":
				var result map[string]string
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "removed", result["action"])
				assert.Equal(t, tt.removeName, result["name"])
			default:
				assert.Equal(t, tt.wantStdout, buf.String())
			}
		})
	}
}

func TestListCmd_NilConfig(t *testing.T) {
	viper.Reset()
	viper.Set("output", "text")

	// Don't set config in context
	ctx := context.Background()

	listCmd.SetContext(ctx)
	err := listCmd.RunE(listCmd, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestSetCmd_NilConfig(t *testing.T) {
	viper.Reset()
	viper.Set("output", "text")

	ctx := context.Background()

	setCmd.SetContext(ctx)
	err := setCmd.RunE(setCmd, []string{"foo", "bar"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestRemoveCmd_NilConfig(t *testing.T) {
	viper.Reset()
	viper.Set("output", "text")

	ctx := context.Background()

	removeCmd.SetContext(ctx)
	err := removeCmd.RunE(removeCmd, []string{"foo"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}
