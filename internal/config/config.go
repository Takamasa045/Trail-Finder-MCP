package config

import (
	"fmt"
	"os"
)

const DefaultVersion = "0.1.0"

func Version() string {
	if v := os.Getenv("TRAILFINDER_VERSION"); v != "" {
		return v
	}
	return DefaultVersion
}

func UserAgent() string {
	if ua := os.Getenv("TRAILFINDER_USER_AGENT"); ua != "" {
		return ua
	}
	return fmt.Sprintf("trail-finder-mcp/%s", Version())
}
