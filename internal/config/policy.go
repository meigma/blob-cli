package config

import (
	"regexp"
	"sync"
)

// patternCache caches compiled regex patterns for performance.
var (
	patternCache   = make(map[string]*regexp.Regexp)
	patternCacheMu sync.RWMutex
)

// GetPoliciesForRef returns all policies that match the given reference.
// The reference should be fully expanded (after alias resolution).
// Returns nil if no policies match.
//
// Multiple matching policies are returned in order; callers typically
// combine them with AND logic (all policies must pass).
func (c *Config) GetPoliciesForRef(ref string) []Policy {
	if len(c.Policies) == 0 {
		return nil
	}

	var matched []Policy
	for _, rule := range c.Policies {
		re, err := getPattern(rule.Match)
		if err != nil {
			// Invalid pattern - skip (should have been caught by validation)
			continue
		}
		if re.MatchString(ref) {
			matched = append(matched, rule.Policy)
		}
	}

	return matched
}

// MatchedPolicyRule contains a matched policy with its original pattern.
type MatchedPolicyRule struct {
	// Pattern is the regex pattern that matched.
	Pattern string

	// Policy is the policy configuration.
	Policy Policy
}

// MatchedPolicyRules returns the policy rules that match the reference,
// including the original match pattern for debugging/display.
func (c *Config) MatchedPolicyRules(ref string) []MatchedPolicyRule {
	if len(c.Policies) == 0 {
		return nil
	}

	var matched []MatchedPolicyRule
	for _, rule := range c.Policies {
		re, err := getPattern(rule.Match)
		if err != nil {
			// Invalid pattern - skip
			continue
		}
		if re.MatchString(ref) {
			matched = append(matched, MatchedPolicyRule{
				Pattern: rule.Match,
				Policy:  rule.Policy,
			})
		}
	}

	return matched
}

// getPattern returns a compiled regex for the given pattern, using cache.
func getPattern(pattern string) (*regexp.Regexp, error) {
	// Try read lock first
	patternCacheMu.RLock()
	re, ok := patternCache[pattern]
	patternCacheMu.RUnlock()

	if ok {
		return re, nil
	}

	// Compile and cache
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	patternCacheMu.Lock()
	patternCache[pattern] = re
	patternCacheMu.Unlock()

	return re, nil
}
