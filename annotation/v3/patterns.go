package annotation

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// 命名 pattern 是预编译、锚定的 regex,通过短名(email、url、...)引用。
// 通过 pattern=<re> 传入的用户 pattern 只编译一次,并在进程生命周期内缓存。
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
		// 兼容 v2 的别名,使现有 tag 继续可用。
		"abc":    regexp.MustCompile(`^[a-zA-Z]+$`),
		"abc123": regexp.MustCompile(`^[a-zA-Z0-9]+$`),
		"123":    regexp.MustCompile(`^\d+$`),
		"digits": regexp.MustCompile(`^\d+$`),
	}
)

// compilePattern 返回给定 pattern 的缓存 *regexp.Regexp。
// 前缀 "#" 选择一个命名 pattern(例如 "#email");否则原样编译该 pattern。
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
