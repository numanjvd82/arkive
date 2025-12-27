package header

import "strings"

func BuildContentDisposition(filename string, disposition string) string {
	name := strings.TrimSpace(filename)
	name = strings.ReplaceAll(name, "\"", "'")
	if name == "" {
		return ""
	}
	kind := strings.ToLower(strings.TrimSpace(disposition))
	if kind != "inline" && kind != "attachment" {
		kind = "attachment"
	}
	return kind + "; filename=\"" + name + "\""
}
