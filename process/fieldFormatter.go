package process

import (
	"regexp"
)

var formatters map[string]func(string) string

func init() {
	formatters = map[string]func(string) string{
		"Rich Text": formatText,
	}
}

func formatField(fieldType string, value string) string {
	if _, ok := formatters[fieldType]; ok {
		return formatters[fieldType](value)
	}
	return value
}

func formatText(val string) string {
	find := `(?m)<p>\s+</p>\n`
	reg := regexp.MustCompile(find)

	val = reg.ReplaceAllString(val, "")

	return val
}
