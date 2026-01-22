package policy

import (
	"errors"
	"fmt"

	"github.com/meigma/blob/policy"
	"github.com/meigma/blob/policy/opa"
	"github.com/meigma/blob/policy/sigstore"
	"github.com/meigma/blob/policy/slsa"
	"github.com/meigma/blob/registry"

	"github.com/meigma/blob-cli/internal/config"
)

// BuildPolicies constructs registry.Policy instances from config and command flags.
// It combines policies from the config file (unless noDefaultPolicy is true)
// with policies from policy files and OPA rego files.
func BuildPolicies(
	cfg *config.Config,
	ref string,
	policyFiles []string,
	policyRego string,
	noDefaultPolicy bool,
) ([]registry.Policy, error) {
	var policies []registry.Policy

	// 1. Config policies (unless skipped)
	if !noDefaultPolicy && cfg != nil {
		configPolicies := cfg.GetPoliciesForRef(ref)
		for i, cfgPolicy := range configPolicies {
			regPolicy, err := ConvertConfigPolicy(cfgPolicy)
			if err != nil {
				return nil, fmt.Errorf("config policy %d: %w", i, err)
			}
			if regPolicy != nil {
				policies = append(policies, regPolicy)
			}
		}
	}

	// 2. YAML policy files
	for _, path := range policyFiles {
		cfgPolicy, err := LoadFile(path)
		if err != nil {
			return nil, fmt.Errorf("loading policy %s: %w", path, err)
		}
		regPolicy, err := ConvertConfigPolicy(*cfgPolicy)
		if err != nil {
			return nil, fmt.Errorf("policy %s: %w", path, err)
		}
		if regPolicy != nil {
			policies = append(policies, regPolicy)
		}
	}

	// 3. OPA Rego file
	if policyRego != "" {
		p, err := opa.NewPolicy(opa.WithPolicyFile(policyRego))
		if err != nil {
			return nil, fmt.Errorf("loading rego policy %s: %w", policyRego, err)
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// ConvertConfigPolicy converts a config.Policy to a registry.Policy.
func ConvertConfigPolicy(cfgPolicy config.Policy) (registry.Policy, error) {
	var policies []registry.Policy

	// Handle signature policy
	if cfgPolicy.Signature != nil {
		sigPolicy, err := buildSignaturePolicy(cfgPolicy.Signature)
		if err != nil {
			return nil, fmt.Errorf("signature policy: %w", err)
		}
		if sigPolicy != nil {
			policies = append(policies, sigPolicy)
		}
	}

	// Handle provenance policy
	if cfgPolicy.Provenance != nil {
		provPolicy, err := buildProvenancePolicy(cfgPolicy.Provenance)
		if err != nil {
			return nil, fmt.Errorf("provenance policy: %w", err)
		}
		if provPolicy != nil {
			policies = append(policies, provPolicy)
		}
	}

	if len(policies) == 0 {
		return nil, nil //nolint:nilnil // nil policy with no error is valid (no verification required)
	}
	if len(policies) == 1 {
		return policies[0], nil
	}
	return policy.RequireAll(policies...), nil
}

// buildSignaturePolicy creates a sigstore policy from config.
func buildSignaturePolicy(sig *config.SignaturePolicy) (registry.Policy, error) {
	// Error if both keyless and key are specified to avoid ambiguity
	if sig.Keyless != nil && sig.Key != nil {
		return nil, errors.New("signature policy cannot specify both keyless and key")
	}

	if sig.Keyless != nil {
		if sig.Keyless.Issuer == "" {
			return nil, errors.New("keyless issuer is required")
		}
		if sig.Keyless.Identity == "" {
			return nil, errors.New("keyless identity is required")
		}
		return sigstore.NewPolicy(
			sigstore.WithIdentity(sig.Keyless.Issuer, sig.Keyless.Identity),
		)
	}
	if sig.Key != nil {
		if sig.Key.Path != "" {
			return nil, errors.New("key-based signature verification not yet implemented")
		}
		if sig.Key.URL != "" {
			return nil, errors.New("key URL signature verification not yet implemented")
		}
		return nil, errors.New("signature key must specify path or url")
	}
	return nil, errors.New("signature policy must specify keyless or key")
}

// buildProvenancePolicy creates an SLSA policy from config.
func buildProvenancePolicy(prov *config.ProvenancePolicy) (registry.Policy, error) {
	if prov.SLSA == nil {
		return nil, errors.New("provenance policy must specify slsa")
	}

	// Repository is required for GitHubActionsWorkflow
	if prov.SLSA.Repository != "" {
		var opts []any

		if prov.SLSA.Branch != "" {
			opts = append(opts, slsa.WithWorkflowBranches(prov.SLSA.Branch))
		}
		if prov.SLSA.Tag != "" {
			opts = append(opts, slsa.WithWorkflowTags(prov.SLSA.Tag))
		}

		return slsa.GitHubActionsWorkflow(prov.SLSA.Repository, opts...)
	}

	// Fallback to basic builder requirement
	if prov.SLSA.Builder != "" {
		return slsa.RequireBuilder(prov.SLSA.Builder), nil
	}

	return nil, errors.New("slsa policy must specify repository or builder")
}
