package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const NavCSS = "/web/components/nav.css"

func Nav() g.Node {
	return g.Group([]g.Node{
		InlineStyle(NavCSS),
		h.Header(
			h.Class("site-nav"),
			h.Div(
				h.Class("container nav-inner"),
				BrandLogo(BrandLogoProps{
					Href:  "/",
					Class: "nav-brand",
				}),
				h.Button(
					h.Class("button secondary theme-toggle"),
					h.Type("button"),
					h.ID("theme-toggle"),
					g.Attr("aria-label", "Theme: system"),
					h.Span(h.Class("theme-label"), g.Text("system")),
				),
			),
		),
	})
}
