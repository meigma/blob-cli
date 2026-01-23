package cmd

import (
	"github.com/meigma/blob"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

// newClient creates a new blob client with options from config.
func newClient(cfg *internalcfg.Config, opts ...blob.Option) (*blob.Client, error) {
	allOpts := append(clientOpts(cfg), opts...)
	return blob.NewClient(allOpts...)
}

// clientOpts returns the base client options from config.
// This is useful when passing options to functions that create their own client.
func clientOpts(cfg *internalcfg.Config) []blob.Option {
	opts := []blob.Option{blob.WithDockerConfig()}
	if cfg.PlainHTTP {
		opts = append(opts, blob.WithPlainHTTP(true))
	}
	return opts
}
