package redisclient

import (
	"fmt"
	"strings"
)

// FmtCacheKey formats a key-value pairs with a standard format
func FmtCacheKey(header string, key any) string {
	return fmt.Sprintf("%s:%v", header, key)
}

// ExtractKey extracts the original key from the standardized format
func ExtractKey(header, cacheKey string) string {
	return strings.ReplaceAll(cacheKey, fmt.Sprintf("%s:", header), "")
}
