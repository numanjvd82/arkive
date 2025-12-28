package components

import (
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type BrandLogoProps struct {
	Text    string
	Href    string
	Class   string
	Stacked bool
}

const BrandLogoCSS = "/web/components/brand_logo.css"

func BrandLogo(props BrandLogoProps) g.Node {
	text := props.Text
	if text == "" {
		text = "arkive.sh"
	}

	classes := []string{"brand-logo"}
	if strings.TrimSpace(props.Class) != "" {
		classes = append(classes, strings.TrimSpace(props.Class))
	}
	if props.Stacked {
		classes = append(classes, "is-stacked")
	}

	content := []g.Node{
		h.Span(
			h.Class("brand-mark"),
			h.Span(h.Class("brand-core")),
		),
		h.Span(h.Class("logo-text"), g.Text(text)),
	}

	var node g.Node
	if props.Href != "" {
		node = h.A(
			h.Class(strings.Join(classes, " ")),
			h.Href(props.Href),
			g.Group(content),
		)
	} else {
		node = h.Span(
			h.Class(strings.Join(classes, " ")),
			g.Group(content),
		)
	}

	return g.Group([]g.Node{
		InlineStyle(BrandLogoCSS),
		node,
	})
}
