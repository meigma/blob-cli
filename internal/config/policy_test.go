package config

import (
	"testing"
)

func TestConfig_GetPoliciesForRef(t *testing.T) {
	cfg := &Config{
		Policies: []PolicyRule{
			{
				Match: `ghcr\.io/acme/.*`,
				Policy: Policy{
					Signature: &SignaturePolicy{
						Keyless: &KeylessConfig{
							Issuer:   "https://token.actions.githubusercontent.com",
							Identity: "https://github.com/acme/*",
						},
					},
				},
			},
			{
				Match: `ghcr\.io/acme/prod-.*`,
				Policy: Policy{
					Provenance: &ProvenancePolicy{
						SLSA: &SLSAConfig{
							Builder: "https://github.com/slsa-framework/*",
							Branch:  "main",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		ref       string
		wantCount int
	}{
		{
			name:      "matches first policy",
			ref:       "ghcr.io/acme/app:v1",
			wantCount: 1,
		},
		{
			name:      "matches both policies",
			ref:       "ghcr.io/acme/prod-app:v1",
			wantCount: 2,
		},
		{
			name:      "no match",
			ref:       "docker.io/library/nginx:latest",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policies := cfg.GetPoliciesForRef(tt.ref)
			if len(policies) != tt.wantCount {
				t.Errorf("GetPoliciesForRef(%q) returned %d policies, want %d", tt.ref, len(policies), tt.wantCount)
			}
		})
	}
}

func TestConfig_GetPoliciesForRef_EmptyPolicies(t *testing.T) {
	cfg := &Config{Policies: nil}
	policies := cfg.GetPoliciesForRef("ghcr.io/acme/app:v1")
	if policies != nil {
		t.Errorf("expected nil, got %v", policies)
	}
}

func TestConfig_MatchedPolicyRules(t *testing.T) {
	cfg := &Config{
		Policies: []PolicyRule{
			{
				Match: `ghcr\.io/acme/.*`,
				Policy: Policy{
					Signature: &SignaturePolicy{
						Keyless: &KeylessConfig{Issuer: "test"},
					},
				},
			},
		},
	}

	matched := cfg.MatchedPolicyRules("ghcr.io/acme/app:v1")

	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}

	if matched[0].Pattern != `ghcr\.io/acme/.*` {
		t.Errorf("pattern = %q, want %q", matched[0].Pattern, `ghcr\.io/acme/.*`)
	}

	if matched[0].Policy.Signature == nil {
		t.Error("policy signature is nil")
	}
}

func TestConfig_GetPoliciesForRef_InvalidPattern(t *testing.T) {
	// Invalid regex should be skipped (not cause panic)
	cfg := &Config{
		Policies: []PolicyRule{
			{Match: "[invalid", Policy: Policy{}}, // invalid regex
			{Match: ".*valid.*", Policy: Policy{Signature: &SignaturePolicy{}}},
		},
	}

	// Should not panic and should return the valid policy
	policies := cfg.GetPoliciesForRef("something valid here")
	if len(policies) != 1 {
		t.Errorf("expected 1 policy (valid one), got %d", len(policies))
	}
}
