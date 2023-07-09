package utils

import "strings"

func GetArguments(message string) []string {
	fields := strings.Fields(message)
	if len(fields) <= 1 {
		return []string{""}
	}
	return fields[1:]
}
