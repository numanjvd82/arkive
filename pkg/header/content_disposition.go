package header

import "strings"

func BuildContentDisposition(filename string, disposition string) string {
	name := strings.TrimSpace(filename)
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, name)
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
