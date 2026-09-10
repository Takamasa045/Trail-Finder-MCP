package config

import (
	"fmt"
	"os"
	"strings"
)

const DefaultVersion = "0.2.0"

const Disclaimer = "参考情報です。現地の状況・装備・公式発表を優先してください。安全は保証しません。 / For reference only. Prioritize local conditions, gear, and official notices. This does not guarantee hiking safety."

func Version() string {
	if v := strings.TrimSpace(os.Getenv("TRAILFINDER_VERSION")); v != "" {
		return v
	}
	return DefaultVersion
}

func UserAgent() string {
	if ua := strings.TrimSpace(os.Getenv("TRAILFINDER_USER_AGENT")); ua != "" {
		return ua
	}
	return fmt.Sprintf("trail-finder-mcp/%s (+https://github.com/Takamasa045/Trail-Finder-MCP)", Version())
}

func Env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func DefaultTZ() string {
	return Env("DEFAULT_TZ", "Asia/Tokyo")
}

func DefaultLang() string {
	return Env("DEFAULT_LANG", "ja")
}
