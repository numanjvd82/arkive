package pages

import (
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
		Title:  "Arkive · Forbidden",
		Robots: RobotsNoIndex,
		CSS:    []string{"/web/pages/forbidden.css"},
		User:   props.Ctx.User,
		Body:   forbiddenBody(),
	}
}

func forbiddenBody() g.Node {
	return h.Div(
		h.Class("forbidden-page"),
		h.Div(
			h.Class("container"),
			h.Div(
				h.Class("forbidden-grid"),
				h.Section(
					h.Class("forbidden-copy"),
					h.P(h.Class("forbidden-eyebrow label"), g.Text("403 error")),
					h.H1(
						h.Class("text-display"),
						g.Text("Access denied"),
					),
					h.P(
						h.Class("forbidden-lead body-lg"),
						g.Text("You must sign in as instance administrator to reach this page."),
					),
					h.Div(
						h.Class("forbidden-actions"),
						components.Button(components.ButtonProps{
							Text:    "Go to login",
							Href:    "/login",
							Variant: "primary",
						}),
						components.Button(components.ButtonProps{
							Text:    "Go to dashboard",
							Href:    "/dashboard",
							Variant: "secondary",
						}),
					),
					h.Div(
						h.Class("forbidden-hint"),
						h.Span(h.Class("label"), g.Text("Tip")),
						h.P(g.Text("Already signed in? Try refreshing or opening the dashboard again.")),
					),
				),
			),
		),
	)
}
