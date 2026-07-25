package annotation

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// Named patterns are pre-compiled, anchored regexes referenced by short names
// (email, url, ...). User-supplied patterns via pattern=<re> are compiled once
// and cached for the lifetime of the process.
var (
	patternCache sync.Map // string -> *regexp.Regexp

	namedPatterns = map[string]*regexp.Regexp{
		"email":    regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`),
		"url":      regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`),
		"uuid":     regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
		"ipv4":     regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`),
		"ipv6":     regexp.MustCompile(`^[0-9a-fA-F:]+$`),
		"alpha":    regexp.MustCompile(`^[a-zA-Z]+$`),
		"alphanum": regexp.MustCompile(`^[a-zA-Z0-9]+$`),
		"numeric":  regexp.MustCompile(`^\d+$`),
		"hex":      regexp.MustCompile(`^[0-9a-fA-F]+$`),
		"date":     regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
		"base64":   regexp.MustCompile(`^[a-zA-Z0-9+/]*={0,2}$`),
		// v2-compatible aliases so existing tags keep working.
		"abc":    regexp.MustCompile(`^[a-zA-Z]+$`),
		"abc123": regexp.MustCompile(`^[a-zA-Z0-9]+$`),
		"123":    regexp.MustCompile(`^\d+$`),
		"digits": regexp.MustCompile(`^\d+$`),
	}
)

// compilePattern returns a cached *regexp.Regexp for the given pattern.
// A leading "#" selects a named pattern (e.g. "#email"); otherwise the
// pattern is compiled verbatim.
func compilePattern(p string) (*regexp.Regexp, error) {
	if v, ok := patternCache.Load(p); ok {
		return v.(*regexp.Regexp), nil
	}
	re, err := compilePatternOnce(p)
	if err != nil {
		return nil, err
	}
	actual, _ := patternCache.LoadOrStore(p, re)
	return actual.(*regexp.Regexp), nil
}

func compilePatternOnce(p string) (*regexp.Regexp, error) {
	if strings.HasPrefix(p, "#") {
		name := strings.TrimPrefix(p, "#")
		if re, ok := namedPatterns[name]; ok {
			return re, nil
		}
		return nil, fmt.Errorf("unknown named pattern: %s", p)
	}
	return regexp.Compile(p)
}
