package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type TooltipProps struct {
	ID       string
	Class    string
	Text     string
	Tooltip  string
	IconName string
	IconSize string
}

func Tooltip(props TooltipProps) g.Node {
	className := "tooltip-icon"
	if props.Class != "" {
		className += " " + props.Class
	}

	content := g.Node(g.Text(props.Text))
	if props.IconName != "" {
		content = Icon(IconProps{
			Name:       props.IconName,
			Size:       props.IconSize,
			Decorative: true,
		})
	}

	return g.If(
		props.Tooltip != "",
		h.Span(
			h.Class(className),
			g.If(props.ID != "", g.Attr("id", props.ID)),
			g.If(props.Tooltip != "", g.Attr("data-tooltip", props.Tooltip)),
			content,
		),
	)
}
