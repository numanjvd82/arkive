package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const SidebarCSS = "/web/components/sidebar.css"
const SidebarJS = "/web/components/sidebar.js"

func DashboardSidebar() g.Node {
	return g.Group([]g.Node{
		InlineStyle(SidebarCSS),
		InlineScript(SidebarJS),
		h.Aside(
			h.Class("dashboard-sidebar"),
			h.ID("dashboard-sidebar"),
			g.Attr("aria-hidden", "true"),
			h.Div(
				h.Class("sidebar-top"),
				BrandLogo(BrandLogoProps{
					Href:  "/dashboard",
					Class: "sidebar-brand",
				}),
				h.Button(
					h.Class("sidebar-close"),
					h.Type("button"),
					g.Attr("aria-label", "Close menu"),
					Icon(IconProps{
						Name:       "x",
						Size:       "16",
						Decorative: true,
					}),
				),
			),
			h.Nav(
				h.Class("sidebar-links"),
				h.A(
					h.Class("sidebar-link is-active"),
					h.Href("/dashboard"),
					g.Text("Dashboard"),
				),
				h.A(
					h.Class("sidebar-link"),
					h.Href("/files"),
					g.Text("Files"),
				),
				h.A(
					h.Class("sidebar-link"),
					h.Href("#upload-panel"),
					g.Text("Uploads"),
				),
			),
			h.Div(
				h.Class("sidebar-footer"),
				h.Button(
					h.Class("sidebar-theme theme-toggle"),
					h.Type("button"),
					h.ID("theme-toggle"),
					g.Attr("aria-label", "Theme: system"),
					h.Span(h.Class("theme-label"), g.Text("system")),
				),
				h.Form(
					h.Method("post"),
					h.Action("/logout"),
					h.Button(h.Class("sidebar-logout"), h.Type("submit"), g.Text("Logout")),
				),
			),
		),
	})
}
