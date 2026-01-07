package web

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web/components"
)

func AuthLayout(data LayoutData, content ...g.Node) g.Node {
	pageTitle := data.Title
	if pageTitle == "" {
		pageTitle = "arkive.sh"
	}

	return h.Doctype(h.HTML(
		h.Lang("en"),
		h.Head(buildHeadNodes(LayoutData{Title: pageTitle, CSS: data.CSS, JS: data.JS})...),
		h.Body(
			components.InlineStyle(components.AuthLayoutCSS),
			h.Div(
				h.Class("app-shell"),
				authHeader(),
				components.DashboardSidebar(),
				h.Div(h.Class("sidebar-scrim"), g.Attr("aria-hidden", "true")),
				h.Div(h.Class("app-content"), g.Group(content)),
				authFooter(),
			),
			components.ToastHost(),
		),
	))
}

func authHeader() g.Node {
	return h.Header(
		h.Class("app-header"),
		h.Div(
			h.Class("app-header-inner"),
			h.Div(
				h.Class("app-header-left"),
				h.Button(
					h.Class("button secondary icon-button"),
					h.Type("button"),
					h.ID("sidebar-toggle"),
					g.Attr("aria-controls", "dashboard-sidebar"),
					g.Attr("aria-expanded", "false"),
					g.Attr("aria-label", "Open menu"),
					h.Span(
						h.Class("app-header-icon"),
						components.Icon(components.IconProps{
							Name:       "menu",
							Size:       "20",
							Decorative: true,
						}),
					),
				),
			),
		),
	)
}

func authFooter() g.Node {
	return h.Footer(
		h.Class("app-footer"),
		h.Div(
			h.Class("app-footer-inner"),
			h.Span(g.Text("Arkive workspace")),
			h.Span(g.Text("Support: support@arkive.sh")),
		),
	)
}
