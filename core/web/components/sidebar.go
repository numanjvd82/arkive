package components

import (
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
)

const SidebarCSS = "/web/components/sidebar.css"

type DashboardSidebarProps struct {
	User      *models.User
	ActiveNav string
}

func DashboardSidebar(props DashboardSidebarProps) g.Node {
	nodeLabel := "Self-hosted core"
	if props.User != nil && strings.TrimSpace(props.User.BrandName) != "" {
		nodeLabel = "Node: " + strings.TrimSpace(props.User.BrandName)
	}

	return g.Group([]g.Node{
		InlineStyle(SidebarCSS),
		h.Aside(
			h.Class("dashboard-sidebar"),
			h.ID("dashboard-sidebar"),
			g.Attr("aria-hidden", "true"),
			h.Div(
				h.Class("sidebar-top"),
				h.Div(
					h.Class("sidebar-brand-block"),
					h.Span(
						h.Class("sidebar-brand-avatar"),
						lucide.CircleUser(
							h.Class("sidebar-lucide sidebar-lucide-avatar"),
							g.Attr("aria-hidden", "true"),
						),
					),
					h.Div(
						h.Class("sidebar-brand-copy"),
						h.A(
							h.Class("sidebar-brand"),
							h.Href("/dashboard"),
							g.Text("Arkive Core"),
						),
						h.Span(h.Class("sidebar-node"), g.Text(nodeLabel)),
					),
				),
				h.Button(
					h.Class("sidebar-close"),
					h.Type("button"),
					g.Attr("aria-label", "Close menu"),
					lucide.X(
						h.Class("sidebar-lucide sidebar-lucide-close"),
						g.Attr("aria-hidden", "true"),
					),
				),
			),
			h.Div(
				h.Class("sidebar-cta"),
				h.A(
					h.Class("button primary sidebar-upload"),
					h.Href("/dashboard#upload-panel"),
					g.Text("Upload File"),
				),
			),
			h.Nav(
				h.Class("sidebar-links"),
				sidebarLink("/dashboard", "dashboard", props.ActiveNav, "Dashboard", lucide.LayoutDashboard),
				sidebarLink("/files", "files", props.ActiveNav, "Files", lucide.FolderOpen),
				sidebarLink("/shares", "shares", props.ActiveNav, "Shares", lucide.Share2),
				sidebarLink("/settings", "settings", props.ActiveNav, "Settings", lucide.Settings),
			),
			h.Div(
				h.Class("sidebar-footer"),
				h.A(
					h.Class("sidebar-footer-link"),
					h.Href("/settings"),
					lucide.HardDrive(
						h.Class("sidebar-lucide sidebar-lucide-nav"),
						g.Attr("aria-hidden", "true"),
					),
					h.Span(g.Text("Storage Status")),
				),
				h.Div(
					h.Class("sidebar-footer-meta"),
					lucide.Shield(
						h.Class("sidebar-lucide sidebar-lucide-nav"),
						g.Attr("aria-hidden", "true"),
					),
					h.Span(g.Text("Single-user core")),
				),
			),
		),
	})
}

func sidebarLink(href, key, active, label string, icon func(...g.Node) g.Node) g.Node {
	className := "sidebar-link"
	if active == key {
		className += " is-active"
	}
	return h.A(
		h.Class(className),
		h.Href(href),
		icon(
			h.Class("sidebar-lucide sidebar-lucide-nav"),
			g.Attr("aria-hidden", "true"),
		),
		h.Span(g.Text(label)),
	)
}
