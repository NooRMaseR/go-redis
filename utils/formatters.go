package utils

import (
	"slices"
	"strings"
)

func FormatListToString(list []string) string {
	if len(list) == 0 {
		return "[]\n"
	}
	return "[ " + strings.Join(list, ", ") + " ]" + "\n"
}

func FormatMapToStringSorted(data map[string]string) string {
	length := len(data)
	if length == 0 {
		return "{}\n"
	}

	keys := make([]string, 0, length)
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var builder strings.Builder
	builder.WriteString("{ ")
	for i, k := range keys {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(data[k])
	}

	builder.WriteString(" }\n")
	return builder.String()
}

func FormatMapToString(data map[string]string) string {
	length := len(data)
	if length == 0 {
		return "{}\n"
	}

	var (
		builder strings.Builder
		first   bool = true
	)

	builder.WriteString("{ ")
	for k, v := range data {
		if !first {
			builder.WriteString(", ")
		}
		first = false
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(v)
	}

	builder.WriteString(" }\n")
	return builder.String()
}
