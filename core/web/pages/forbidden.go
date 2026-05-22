package pages

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type ForbiddenPageProps struct {
	Ctx PageContext
}

func ForbiddenPage(props ForbiddenPageProps) web.Page {
	return web.Page{
		Title:   "Arkive · Forbidden",
		Robots:  RobotsNoIndex,
		CSS:     []string{"/web/pages/forbidden.css"},
		HideNav: true,
		User:    props.Ctx.User,
		Body:    forbiddenBody(),
	}
}

func forbiddenBody() g.Node {
	return h.Div(
		h.Class("forbidden-page"),
		h.Div(h.Class("forbidden-bg-texture")),
		h.Div(h.Class("forbidden-scanline")),
		h.Main(
			h.Class("forbidden-shell"),
			h.Div(
				h.Class("forbidden-lock-wrap"),
				h.Div(h.Class("forbidden-lock-glow")),
				h.Div(
					h.Class("forbidden-lock-card"),
					lucide.Lock(
						h.Class("forbidden-lock-icon"),
						g.Attr("aria-hidden", "true"),
					),
					h.Div(h.Class("forbidden-corner forbidden-corner-top")),
					h.Div(h.Class("forbidden-corner forbidden-corner-bottom")),
				),
			),
			h.Section(
				h.Class("forbidden-copy"),
				h.P(h.Class("forbidden-eyebrow"), g.Text("Error Code: 403_FORBIDDEN")),
				h.H1(h.Class("forbidden-title"), g.Text("403: Access Denied")),
				h.Div(h.Class("forbidden-rule")),
				h.P(
					h.Class("forbidden-lead"),
					g.Text("You do not have permission to access this resource."),
				),
				h.Div(
					h.Class("forbidden-actions"),
					components.Button(components.ButtonProps{
						Text:    "Dashboard",
						Href:    "/dashboard",
						Variant: "primary",
						Class:   "forbidden-action-primary",
						Icon:    "shield",
					}),
					components.Button(components.ButtonProps{
						Text:    "Login",
						Href:    "/login",
						Variant: "secondary",
						Class:   "forbidden-action-secondary",
						Icon:    "arrow-left",
					}),
				),
				h.Div(
					h.Class("forbidden-footer"),
					h.P(
						h.Class("forbidden-footer-copy"),
						g.Text("Powered by "),
						h.Span(h.Class("forbidden-footer-strong"), g.Text("Arkive Core")),
					),
					lucide.BadgeCheck(
						h.Class("forbidden-footer-icon"),
						g.Attr("aria-hidden", "true"),
					),
				),
			),
		),
	)
}
