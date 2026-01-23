package cmd

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatCmd_NilConfig(t *testing.T) {
	viper.Reset()

	ctx := context.Background()

	catCmd.SetContext(ctx)
	err := catCmd.RunE(catCmd, []string{"ghcr.io/test:v1", "config.json"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestCatCmd_MinimumArgs(t *testing.T) {
	// Verify command requires at least 2 args (ref + file)
	assert.Equal(t, "cat <ref> <file>...", catCmd.Use)

	// Cobra's MinimumNArgs(2) is set
	err := catCmd.Args(catCmd, []string{"only-one-arg"})
	require.Error(t, err)

	err = catCmd.Args(catCmd, []string{"ref", "file"})
	require.NoError(t, err)

	err = catCmd.Args(catCmd, []string{"ref", "file1", "file2"})
	require.NoError(t, err)
}
