package server

import (
	"fmt"
	"strings"
)

type Toolset string
type AccessMode string

const (
	ToolsetDeveloper Toolset = "developer"
	ToolsetBusiness  Toolset = "business"
	ToolsetAll       Toolset = "all"

	AccessReadWrite AccessMode = "read_write"
	AccessReadOnly  AccessMode = "read_only"
)

type Options struct {
	Toolset Toolset
	Profile string
	Access  AccessMode
}

func ParseToolset(value string) (Toolset, error) {
	normalized := Toolset(strings.ToLower(strings.TrimSpace(value)))
	switch normalized {
	case "", ToolsetAll:
		return ToolsetAll, nil
	case ToolsetDeveloper, ToolsetBusiness:
		return normalized, nil
	default:
		return "", fmt.Errorf("unsupported toolset %q (allowed: developer|business|all)", value)
	}
}

func defaultOptions() Options {
	return Options{
		Toolset: ToolsetAll,
		Profile: "generic",
		Access:  AccessReadWrite,
	}
}

func ParseAccessMode(value string) (AccessMode, error) {
	normalized := AccessMode(strings.ToLower(strings.TrimSpace(value)))
	switch normalized {
	case "", AccessReadWrite:
		return AccessReadWrite, nil
	case AccessReadOnly:
		return AccessReadOnly, nil
	default:
		return "", fmt.Errorf("unsupported access mode %q (allowed: read_write|read_only)", value)
	}
}
