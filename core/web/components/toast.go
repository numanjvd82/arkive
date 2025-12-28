package components

import (
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
				Icon(IconProps{Name: "check", Size: "20", Decorative: true, Class: "toast-status-icon"}),
			),
			h.Span(
				g.Attr("data-toast-icon", "warning"),
				Icon(IconProps{Name: "warning", Size: "20", Decorative: true, Class: "toast-status-icon"}),
			),
			h.Span(
				g.Attr("data-toast-icon", "error"),
				Icon(IconProps{Name: "error", Size: "20", Decorative: true, Class: "toast-status-icon"}),
			),
			h.Span(
				g.Attr("data-toast-icon", "info"),
				Icon(IconProps{Name: "info", Size: "20", Decorative: true, Class: "toast-status-icon"}),
			),
			h.Span(
				g.Attr("data-toast-icon", "close"),
				Icon(IconProps{Name: "x", Size: "16", Decorative: true, Class: "toast-close-icon"}),
			),
		),
	})
}
