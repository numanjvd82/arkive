package components

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const ToastCSS = "/web/components/toast.css"
const ToastJS = "/web/components/toast.js"

func ToastHost() g.Node {
	return g.Group([]g.Node{
		h.Div(
			h.ID("toast-host"),
			h.Class("toast-host"),
			g.Attr("aria-live", "polite"),
			g.Attr("aria-atomic", "true"),
		),
		h.Div(
			h.Class("toast-icon-bank"),
			g.Attr("aria-hidden", "true"),
			h.Span(
				g.Attr("data-toast-icon", "success"),
				lucide.CircleCheckBig(
					h.Class("toast-lucide"),
					g.Attr("aria-hidden", "true"),
				),
			),
			h.Span(
				g.Attr("data-toast-icon", "warning"),
				lucide.TriangleAlert(
					h.Class("toast-lucide"),
					g.Attr("aria-hidden", "true"),
				),
			),
			h.Span(
				g.Attr("data-toast-icon", "error"),
				lucide.CircleAlert(
					h.Class("toast-lucide"),
					g.Attr("aria-hidden", "true"),
				),
			),
			h.Span(
				g.Attr("data-toast-icon", "info"),
				lucide.Info(
					h.Class("toast-lucide"),
					g.Attr("aria-hidden", "true"),
				),
			),
			h.Span(
				g.Attr("data-toast-icon", "processing"),
				lucide.LoaderCircle(
					h.Class("toast-lucide"),
					g.Attr("aria-hidden", "true"),
				),
			),
			h.Span(
				g.Attr("data-toast-icon", "close"),
				lucide.X(
					h.Class("toast-lucide toast-lucide-close"),
					g.Attr("aria-hidden", "true"),
				),
			),
		),
	})
}
