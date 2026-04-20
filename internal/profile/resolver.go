package profile

import (
	"context"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/onec"
)

const (
	Auto    = "auto"
	Generic = "generic"
	Buh30   = "buh_3_0"
	Unknown = "unknown"
)

// Normalize validates and normalizes profile value from CLI.
func Normalize(value string) (string, error) {
	profile := strings.ToLower(strings.TrimSpace(value))
	switch profile {
	case "", Auto:
		return Auto, nil
	case Generic, Buh30, Unknown:
		return profile, nil
	default:
		return "", fmt.Errorf("unsupported profile %q (allowed: auto|generic|buh_3_0|unknown)", value)
	}
}

// Resolve returns the effective profile.
// If profileFlag is explicit (not auto), it is returned as-is.
// If profileFlag is auto, the function tries to detect a known profile from /configuration.
func Resolve(ctx context.Context, client *onec.Client, profileFlag string) (string, error) {
	normalized, err := Normalize(profileFlag)
	if err != nil {
		return "", err
	}
	if normalized != Auto {
		return normalized, nil
	}

	var info onec.ConfigurationInfo
	if err := client.Get(ctx, "/configuration", &info); err != nil {
		return Generic, err
	}
	return Detect(info.Name, info.Version), nil
}

// Detect identifies profile based on configuration name/version.
func Detect(configName, configVersion string) string {
	name := strings.ToLower(strings.TrimSpace(configName))
	version := strings.TrimSpace(configVersion)

	if strings.Contains(name, "бухгалтерияпредприятия") && strings.HasPrefix(version, "3.0") {
		return Buh30
	}
	return Generic
}
