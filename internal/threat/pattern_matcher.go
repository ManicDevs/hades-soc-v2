package threat

import (
	"regexp"
	"sync"
)

// PatternMatcher manages pattern-based threat detection
type PatternMatcher struct {
	patterns map[string]*regexp.Regexp
	mu       sync.RWMutex
}

// NewPatternMatcher creates a new pattern matcher
func NewPatternMatcher() *PatternMatcher {
	pm := &PatternMatcher{
		patterns: make(map[string]*regexp.Regexp),
	}

	// Initialize common threat patterns
	pm.initializePatterns()
	return pm
}

// initializePatterns initializes common threat detection patterns
func (pm *PatternMatcher) initializePatterns() {
	patterns := map[string]string{
		"sql_injection":     `(?i)(union|select|insert|update|delete|drop|create|alter|exec|script)\s+(all\s+)?\*?\s*(from|into)`,
		"xss":               `(?i)(<script|javascript:|onload=|onerror=|alert\(|document\.)`,
		"path_traversal":    `(?i)(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c)`,
		"command_injection": `(?i)(;|\||&|\$\(|` + "`" + `|wget|curl|nc|netcat|sh|bash|cmd|powershell)`,
		"ldap_injection":    `(?i)(\(\)|\(\|\)|\(!)`,
		"xml_injection":     `(?i)(<!DOCTYPE|<ENTITY|<SYSTEM|<CDATA)`,
		"no_sql_injection":  `(?i)(sleep\(|benchmark\(|pg_sleep\(|waitfor\s+delay)`,
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			pm.patterns[name] = regex
		}
	}
}

// MatchPattern checks if input matches any known threat patterns
func (pm *PatternMatcher) MatchPattern(input string) (string, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for name, pattern := range pm.patterns {
		if pattern.MatchString(input) {
			return name, true
		}
	}

	return "", false
}

// AddPattern adds a new threat pattern
func (pm *PatternMatcher) AddPattern(name, pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.patterns[name] = regex
	return nil
}

// GetPatterns returns all registered patterns
func (pm *PatternMatcher) GetPatterns() map[string]string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	patterns := make(map[string]string)
	for name, regex := range pm.patterns {
		patterns[name] = regex.String()
	}

	return patterns
}
