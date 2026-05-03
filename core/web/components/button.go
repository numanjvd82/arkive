package components

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	c "maragu.dev/gomponents/components"
	h "maragu.dev/gomponents/html"
)

type ButtonProps struct {
	Text    string
	Variant string
	Type    string
	Href    string
	Class   string
	ID      string
	Icon    string
}

const ButtonCSS = "/web/components/button.css"

func Button(props ButtonProps) g.Node {
	classes := c.Classes{
		"button": true,
	}

	if props.Variant != "" {
		classes[props.Variant] = true
	}

	if props.Class != "" {
		classes[props.Class] = true
	}

	if props.Href != "" {
		return g.Group([]g.Node{
			InlineStyle(ButtonCSS),
			h.A(
				h.Class(classes.String()),
				h.Href(props.Href),
				g.If(props.ID != "", g.Attr("id", props.ID)),
				g.If(props.Icon != "", h.Span(
					h.Class("button-icon"),
					renderButtonIcon(props.Icon),
				)),
				h.Span(g.Text(props.Text)),
			),
		})
	}

	buttonType := props.Type
	if buttonType == "" {
		buttonType = "button"
	}

	return g.Group([]g.Node{
		InlineStyle(ButtonCSS),
		h.Button(
			h.Class(classes.String()),
			g.Attr("type", buttonType),
			g.If(props.ID != "", g.Attr("id", props.ID)),
			g.If(props.Icon != "", h.Span(
				h.Class("button-icon"),
				renderButtonIcon(props.Icon),
			)),
			h.Span(g.Text(props.Text)),
		),
	})
}

func renderButtonIcon(name string) g.Node {
	switch name {
	case "key":
		return lucide.Key(
			h.Class("button-lucide"),
			g.Attr("aria-hidden", "true"),
		)
	default:
		return Icon(IconProps{
			Name:       name,
			Size:       "18",
			Decorative: true,
		})
	}
}
