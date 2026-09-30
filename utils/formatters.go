package utils

import "strings"


func FormatListToString(list []string) string {
	return "[ " + strings.Join(list, ", ") + " ]" + "\n"
}

func FormatMapToString(data map[string]string) string {
	var (
		builder strings.Builder
		length  int = len(data) - 1
		counter int = 0
	)
	builder.WriteString("{ ")
	for k, v := range data {
		counter++
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(v)
		if counter <= length {
			builder.WriteString(", ")
		}
	}
	builder.WriteString(" }\n")
	return builder.String()
}