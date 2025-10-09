package model_utils

import (
	"one-api/common/config"
	"strings"
)

// HasPrefixCaseInsensitive 大小写不敏感的前缀匹配
func HasPrefixCaseInsensitive(s, prefix string) bool {
	if config.ModelNameCaseInsensitiveEnabled {
		return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
	}
	return strings.HasPrefix(s, prefix)
}

// ContainsCaseInsensitive 大小写不敏感的包含匹配
func ContainsCaseInsensitive(s, substr string) bool {
	if config.ModelNameCaseInsensitiveEnabled {
		return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
	}
	return strings.Contains(s, substr)
}
