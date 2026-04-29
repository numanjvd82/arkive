package pages

import (
	"fmt"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type NotFoundPageProps struct {
	Ctx  PageContext
	Path string
}

func NotFoundPage(props NotFoundPageProps) web.Page {
	title := "Arkive · Page not found"
	if props.Path != "" {
		title = fmt.Sprintf("Arkive · %s not found", props.Path)
	}

	return web.Page{
		Title:  title,
		Robots: RobotsNoIndex,
		CSS:    []string{"/web/pages/not_found.css"},
		User:   props.Ctx.User,
		Body: h.Main(
			h.Class("not-found"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("not-found-grid"),
					h.Section(
						h.Class("not-found-copy"),
						h.P(h.Class("not-found-eyebrow label"), g.Text("404 error")),
						h.H1(
							h.Class("text-display"),
							g.Text("This page drifted off course"),
						),
						h.P(
							h.Class("not-found-lead body-lg"),
							g.Text("We could not find what you were looking for. Try heading back home or jump into your library."),
						),
						h.Div(
							h.Class("not-found-actions"),
							components.Button(components.ButtonProps{
								Text:    "Go to home",
								Href:    "/",
								Variant: "primary",
							}),
							components.Button(components.ButtonProps{
								Text:    "Contact support",
								Href:    "/contact",
								Variant: "secondary",
							}),
						),
						h.Div(
							h.Class("not-found-hint"),
							h.Span(h.Class("label"), g.Text("Tip")),
							h.P(g.Text("Check the URL or use the search from your dashboard.")),
						),
					),
					h.Aside(
						h.Class("not-found-ads"),
						h.Div(
							h.Class("not-found-panel ad-slot compact"),
							h.P(h.Class("ad-label"), g.Text("Ad slot")),
							h.Script(
								g.Attr("async", "async"),
								g.Attr("data-cfasync", "false"),
								h.Src("https://pl28425100.effectivegatecpm.com/3e709d756892597be3b0708e86694b25/invoke.js"),
							),
							h.Div(h.ID("container-3e709d756892597be3b0708e86694b25")),
						),
						h.Div(
							h.Class("not-found-panel"),
							h.P(h.Class("ad-label"), g.Text("Information")),
							h.P(g.Text("Display and vignette ads help keep Arkive running.")),
						),
					),
				),
			),
		),
	}
}
