package components

import (
	lucide "github.com/eduardolat/gomponents-lucide"
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

const TooltipCSS = "/web/components/tooltip.css"

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
	} else if props.Text == "" {
		content = lucide.Info(
			h.Class("tooltip-icon-glyph"),
			g.Attr("aria-hidden", "true"),
		)
	}

	return g.If(
		props.Tooltip != "",
		g.Group([]g.Node{
			InlineStyle(TooltipCSS),
			h.Span(
				h.Class(className),
				g.If(props.ID != "", g.Attr("id", props.ID)),
				g.Attr("tabindex", "0"),
				g.Attr("role", "button"),
				g.Attr("aria-haspopup", "true"),
				g.Attr("aria-label", props.Tooltip),
				g.Attr("data-tooltip", props.Tooltip),
				content,
			),
		}),
	)
}
