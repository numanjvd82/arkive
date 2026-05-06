package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const LightboxCSS = "/web/components/lightbox.css"

func Lightbox() g.Node {
	return g.Group([]g.Node{
		InlineStyle(LightboxCSS),
		h.Div(
			h.ID("lightbox-backdrop"),
			h.Class("lightbox-backdrop is-hidden"),
			g.Attr("aria-hidden", "true"),
			g.Attr("data-lightbox-state", "closed"),
			h.Div(
				h.Class("lightbox-shell"),
				g.Attr("role", "dialog"),
				g.Attr("aria-modal", "true"),
				h.Div(
					h.Class("lightbox-stage"),
					h.Img(
						h.ID("lightbox-image"),
						h.Class("lightbox-image"),
						h.Alt(""),
					),
				),
				h.Div(
					h.Class("lightbox-controls"),
					h.Div(
						h.Class("lightbox-meta"),
						h.Span(h.Class("lightbox-title"), g.Attr("id", "lightbox-title")),
						h.Span(h.Class("lightbox-hint"), g.Text("Drag to pan · Pinch to zoom · Scroll to zoom")),
					),
					h.Div(
						h.Class("lightbox-actions"),
						h.Button(
							h.Class("lightbox-action"),
							g.Attr("type", "button"),
							g.Attr("data-lightbox-action", "fit"),
							g.Text("Fit"),
						),
						h.Button(
							h.Class("lightbox-action"),
							g.Attr("type", "button"),
							g.Attr("data-lightbox-action", "zoom-out"),
							g.Text("–"),
						),
						h.Button(
							h.Class("lightbox-action"),
							g.Attr("type", "button"),
							g.Attr("data-lightbox-action", "zoom-in"),
							g.Text("+"),
						),
						h.Button(
							h.Class("lightbox-action lightbox-close"),
							g.Attr("type", "button"),
							g.Attr("aria-label", "Close"),
							Icon(IconProps{Name: "x", Size: "16", Decorative: true}),
						),
					),
				),
			),
		),
	})
}
