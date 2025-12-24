package components

import (
	"html"
	"os"
	"path/filepath"
	"strings"

	g "maragu.dev/gomponents"
	c "maragu.dev/gomponents/components"
	h "maragu.dev/gomponents/html"
)

type IconProps struct {
	Name       string
	Size       string
	Color      string
	Class      string
	Title      string
	AriaLabel  string
	Decorative bool
}

var iconSizes = map[string]string{
	"sm": "16",
	"md": "24",
	"lg": "32",
	"xl": "48",
}

func Icon(props IconProps) g.Node {
	classes := c.Classes{
		"icon": true,
	}

	if props.Class != "" {
		classes[props.Class] = true
	}

	sizeAttr := ""
	if size, ok := iconSizes[props.Size]; ok {
		sizeAttr = size
	} else if props.Size != "" {
		sizeAttr = props.Size
	} else {
		sizeAttr = iconSizes["md"]
	}

	iconPath := filepath.Join("core", "web", "static", "icons", props.Name+".svg")
	raw, err := os.ReadFile(iconPath)
	if err != nil {
		return h.Span(h.Class(classes.String()))
	}

	attrParts := []string{
		`width="` + html.EscapeString(sizeAttr) + `"`,
		`height="` + html.EscapeString(sizeAttr) + `"`,
		`class="` + html.EscapeString(classes.String()) + `"`,
	}
	if props.Color != "" {
		attrParts = append(attrParts, `style="color: `+html.EscapeString(props.Color)+`"`)
	}

	if props.Decorative || (props.Title == "" && props.AriaLabel == "") {
		attrParts = append(attrParts, `aria-hidden="true"`, `role="presentation"`)
	} else {
		label := props.AriaLabel
		if label == "" {
			label = props.Title
		}
		attrParts = append(attrParts, `role="img"`, `aria-label="`+html.EscapeString(label)+`"`)
	}

	icon := string(raw)
	openIdx := strings.Index(icon, "<svg")
	if openIdx == -1 {
		return h.Span(h.Class(classes.String()))
	}
	closeIdx := strings.Index(icon[openIdx:], ">")
	if closeIdx == -1 {
		return h.Span(h.Class(classes.String()))
	}
	closeIdx += openIdx

	openTag := icon[openIdx : closeIdx+1]
	openTag = strings.TrimSuffix(openTag, ">")
	openTag = openTag + " " + strings.Join(attrParts, " ") + ">"

	icon = icon[:openIdx] + openTag + icon[closeIdx+1:]
	if props.Title != "" && !props.Decorative {
		titleTag := "<title>" + html.EscapeString(props.Title) + "</title>"
		icon = icon[:closeIdx+1] + titleTag + icon[closeIdx+1:]
	}

	return g.Raw(icon)
}
