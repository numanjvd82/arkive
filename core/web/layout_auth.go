package web

import (
	"fmt"
	"strings"
	"time"

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

	return h.Doctype(h.HTML(
		h.Lang("en"),
		h.Head(buildHeadNodes(LayoutData{
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
		})...),
		h.Body(
			components.InlineStyle(components.AuthLayoutCSS),
			h.Div(
				h.Class("app-shell"),
				authHeader(data.User),
				components.DashboardSidebar(),
				h.Div(h.Class("sidebar-scrim"), g.Attr("aria-hidden", "true")),
				h.Div(h.Class("app-content"), g.Group(content)),
				authFooter(),
			),
			components.ToastHost(),
		),
	))
}

func authHeader(user *models.User) g.Node {
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
			h.Div(
				h.Class("app-header-right"),
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
		ID:    "user-menu",
		Align: "right",
		Label: "Open account menu",
		Trigger: components.Avatar(components.AvatarProps{
			Text:       brandName,
			Size:       40,
			Decorative: true,
		}),
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

func authFooter() g.Node {
	return h.Footer(
		h.Class("app-footer"),
		h.Div(
			h.Class("app-footer-inner"),
			h.Span(g.Text("Arkive Workspace. Securely store and share your files.")),
			h.Span(g.Text("Support: support@arkive.sh")),
			h.Span(g.Text(fmt.Sprintf("© %d Arkive. All rights reserved.", time.Now().Year()))),
		),
	)
}
