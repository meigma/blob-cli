// Package policy provides utilities for loading and building verification policies.
package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/meigma/blob-cli/internal/config"
)

// File represents a YAML policy file structure.
// This matches the format described in DESIGN.md.
type File struct {
	Signature  *SignatureFile  `yaml:"signature"`
	Provenance *ProvenanceFile `yaml:"provenance"`
}

// SignatureFile defines signature verification in a policy file.
type SignatureFile struct {
	Keyless *KeylessFile `yaml:"keyless"`
	Key     *KeyFile     `yaml:"key"`
}

// KeylessFile defines Sigstore keyless verification.
type KeylessFile struct {
	Issuer   string `yaml:"issuer"`
	Identity string `yaml:"identity"`
}

// KeyFile defines key-based signature verification.
type KeyFile struct {
	Path string `yaml:"path"`
	URL  string `yaml:"url"`
}

// ProvenanceFile defines provenance verification in a policy file.
type ProvenanceFile struct {
	SLSA *SLSAFile `yaml:"slsa"`
}

// SLSAFile defines SLSA provenance requirements.
type SLSAFile struct {
	Builder    string `yaml:"builder"`
	Repository string `yaml:"repository"`
	Branch     string `yaml:"branch"`
	Tag        string `yaml:"tag"`
}

// LoadFile loads and parses a YAML policy file.
func LoadFile(path string) (*config.Policy, error) {
	//nolint:gosec // path is intentionally user-provided for policy loading
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading policy file: %w", err)
	}

	var pf File
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("parsing policy file: %w", err)
	}

	return convertFileToConfig(&pf), nil
}

// convertFileToConfig converts a policy file to config.Policy.
func convertFileToConfig(pf *File) *config.Policy {
	if pf == nil {
		return nil
	}

	p := &config.Policy{}

	if pf.Signature != nil {
		p.Signature = &config.SignaturePolicy{}
		if pf.Signature.Keyless != nil {
			p.Signature.Keyless = &config.KeylessConfig{
				Issuer:   pf.Signature.Keyless.Issuer,
				Identity: pf.Signature.Keyless.Identity,
			}
		}
		if pf.Signature.Key != nil {
			p.Signature.Key = &config.KeyConfig{
				Path: pf.Signature.Key.Path,
				URL:  pf.Signature.Key.URL,
			}
		}
	}

	if pf.Provenance != nil {
		p.Provenance = &config.ProvenancePolicy{}
		if pf.Provenance.SLSA != nil {
			p.Provenance.SLSA = &config.SLSAConfig{
				Builder:    pf.Provenance.SLSA.Builder,
				Repository: pf.Provenance.SLSA.Repository,
				Branch:     pf.Provenance.SLSA.Branch,
				Tag:        pf.Provenance.SLSA.Tag,
			}
		}
	}

	return p
}
