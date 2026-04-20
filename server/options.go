package server

import (
	"fmt"
	"strings"
)

type Toolset string

const (
	ToolsetDeveloper Toolset = "developer"
	ToolsetBusiness  Toolset = "business"
	ToolsetAll       Toolset = "all"
)

type Options struct {
	Toolset Toolset
	Profile string
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
	}
}
