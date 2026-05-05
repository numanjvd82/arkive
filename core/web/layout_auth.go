package web

import (
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web/components"
)

func AuthLayout(data LayoutData, content ...g.Node) g.Node {
	pageTitle := data.Title
	if pageTitle == "" {
		pageTitle = "arkive.sh"
	}

	headNodes := buildHeadNodes(LayoutData{
		Title:         pageTitle,
		Description:   data.Description,
		CanonicalURL:  data.CanonicalURL,
		Robots:        data.Robots,
		OGTitle:       data.OGTitle,
		OGDescription: data.OGDescription,
		OGImage:       data.OGImage,
		OGType:        data.OGType,
		TwitterCard:   data.TwitterCard,
		JSONLD:        data.JSONLD,
		CSS:           data.CSS,
		JS:            data.JS,
	})
	return h.Doctype(h.HTML(
		h.Lang("en"),
		h.Head(headNodes...),
		h.Body(
			components.InlineStyle(components.AuthLayoutCSS),
			h.Div(
				h.Class("app-shell"),
				authHeader(data.User, data.SearchPlaceholder),
				components.DashboardSidebar(components.DashboardSidebarProps{
					User:      data.User,
					ActiveNav: data.ActiveNav,
				}),
				h.Div(h.Class("sidebar-scrim"), g.Attr("aria-hidden", "true")),
				h.Div(h.Class("app-content"), g.Group(content)),
			),
			components.SearchPalette(data.SearchPlaceholder),
			components.ToastHost(),
		),
	))
}

func authHeader(user *models.User, placeholder string) g.Node {
	if strings.TrimSpace(placeholder) == "" {
		placeholder = "Search system..."
	}
	return h.Header(
		h.Class("app-header"),
		h.Div(
			h.Class("app-header-inner"),
			h.Div(
				h.Class("app-header-left"),
				h.Button(
					h.Class("button secondary icon-button sidebar-toggle"),
					h.Type("button"),
					h.ID("sidebar-toggle"),
					g.Attr("aria-controls", "dashboard-sidebar"),
					g.Attr("aria-expanded", "false"),
					g.Attr("aria-label", "Open menu"),
					lucide.Menu(
						h.Class("app-header-lucide"),
						g.Attr("aria-hidden", "true"),
					),
				),
				h.Div(
					h.Class("app-search"),
					lucide.Search(
						h.Class("app-header-lucide app-search-icon"),
						g.Attr("aria-hidden", "true"),
					),
					h.Button(
						h.Class("app-search-trigger"),
						h.ID("app-search-trigger"),
						h.Type("button"),
						g.Attr("aria-label", placeholder),
						h.Span(h.Class("app-search-trigger-text"), g.Text(placeholder)),
						h.Kbd(h.Class("app-search-trigger-kbd"), g.Text("/")),
					),
				),
			),
			h.Div(
				h.Class("app-header-right"),
				h.A(
					h.Class("app-header-action"),
					h.Href("/dashboard#recent-activity"),
					g.Attr("aria-label", "Jump to recent activity"),
					lucide.Bell(
						h.Class("app-header-lucide"),
						g.Attr("aria-hidden", "true"),
					),
				),
				renderUserMenu(user),
			),
		),
	)
}

func renderUserMenu(user *models.User) g.Node {
	if user == nil {
		return h.Span()
	}

	brandName := strings.TrimSpace(user.BrandName)
	if brandName == "" {
		brandName = "Account"
	}

	return components.Dropdown(components.DropdownProps{
		Align: "right",
		Label: "Open account menu",
		Class: "app-user-dropdown",
		Trigger: lucide.CircleUser(
			h.Class("app-header-lucide app-user-icon"),
			g.Attr("aria-hidden", "true"),
		),
		Menu: h.Div(
			h.Class("dropdown-content"),
			h.Div(
				h.Class("dropdown-meta"),
				h.Span(h.Class("dropdown-label"), g.Text(brandName)),
				h.Span(h.Class("dropdown-email"), g.Text(user.Email)),
			),
			h.Div(h.Class("dropdown-divider")),
			h.A(h.Class("dropdown-item"), h.Href("/settings"), g.Attr("role", "menuitem"), g.Text("Settings")),
			h.Form(
				h.Method("post"),
				h.Action("/logout"),
				h.Button(
					h.Class("dropdown-item is-danger"),
					h.Type("submit"),
					g.Attr("role", "menuitem"),
					g.Text("Logout"),
				),
			),
		),
	})
}
