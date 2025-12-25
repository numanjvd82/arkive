package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

func ForbiddenPage() web.Page {
	return web.Page{
		Title: "Arkive · Forbidden",
		CSS:   []string{"/web/pages/forbidden.css"},
		Body:  forbiddenBody(),
	}
}

func forbiddenBody() g.Node {
	return h.Div(
		h.Class("forbidden-page"),
		h.Div(
			h.Class("forbidden-container"),
			components.Card(components.CardProps{
				Title:    "403 · Forbidden",
				Subtitle: "You need an account to access this page.",
				Class:    "forbidden-card",
				Body: []g.Node{
					h.P(
						h.Class("forbidden-text"),
						g.Text("Log in to continue to your dashboard."),
					),
					components.Button(components.ButtonProps{
						Text:    "Go to login",
						Href:    "/login",
						Variant: "primary",
						Class:   "forbidden-action",
					}),
				},
			}),
		),
	)
}
