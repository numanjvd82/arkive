package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func Nav() g.Node {
	return h.Header(
		h.Class("site-nav"),
		h.Div(
			h.Class("container nav-inner"),
			h.A(
				h.Class("nav-brand"),
				h.Href("/"),
				g.Text("arkive.sh"),
			),
			h.Button(
				h.Class("button secondary theme-toggle"),
				h.Type("button"),
				h.ID("theme-toggle"),
				g.Attr("aria-label", "Theme: system"),
				h.Span(h.Class("theme-label"), g.Text("system")),
			),
		),
	)
}
