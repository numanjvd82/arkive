package components

import (
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type CopyButtonProps struct {
	Text      string
	TargetID  string
	Value     string
	Class     string
	Variant   string
	AriaLabel string
}

const CopyButtonJS = "/web/components/copy_button.js"

func CopyButton(props CopyButtonProps) g.Node {
	classes := []string{"button"}
	if strings.TrimSpace(props.Variant) != "" {
		classes = append(classes, strings.TrimSpace(props.Variant))
	}
	if strings.TrimSpace(props.Class) != "" {
		classes = append(classes, strings.TrimSpace(props.Class))
	}
	label := props.Text
	if strings.TrimSpace(label) == "" {
		label = "Copy"
	}

	attrs := []g.Node{
		h.Class(strings.Join(classes, " ")),
		h.Type("button"),
		g.Attr("data-copy-button", "true"),
	}
	if props.AriaLabel != "" {
		attrs = append(attrs, g.Attr("aria-label", props.AriaLabel))
	}
	if props.Value != "" {
		attrs = append(attrs, g.Attr("data-copy-value", props.Value))
	} else if props.TargetID != "" {
		attrs = append(attrs, g.Attr("data-copy-target", props.TargetID))
	}

	return g.Group([]g.Node{
		InlineScript(CopyButtonJS),
		h.Button(
			g.Group(attrs),
			g.Text(label),
		),
	})
}
