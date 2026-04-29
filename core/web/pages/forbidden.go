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
	showAds := shouldShowAds(props.Ctx)
	return web.Page{
		Title:  "Arkive · Forbidden",
		Robots: RobotsNoIndex,
		CSS:    []string{"/web/pages/forbidden.css"},
		JS:     buildForbiddenJS(showAds),
		User:   props.Ctx.User,
		Body:   forbiddenBody(showAds),
	}
}

func forbiddenBody(showAds bool) g.Node {
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
						g.Text("You need an account to reach this page. Sign in to continue to your dashboard."),
					),
					h.Div(
						h.Class("forbidden-actions"),
						components.Button(components.ButtonProps{
							Text:    "Go to login",
							Href:    "/login",
							Variant: "primary",
						}),
						components.Button(components.ButtonProps{
							Text:    "View pricing",
							Href:    "/pricing",
							Variant: "secondary",
						}),
					),
					h.Div(
						h.Class("forbidden-hint"),
						h.Span(h.Class("label"), g.Text("Tip")),
						h.P(g.Text("Already signed in? Try refreshing or opening the dashboard again.")),
					),
				),
				g.If(showAds, h.Aside(
					h.Class("forbidden-ads"),
					h.Div(
						h.Class("forbidden-panel ad-slot compact"),
						h.P(h.Class("ad-label"), g.Text("Ad slot")),
						h.Script(
							g.Attr("async", "async"),
							g.Attr("data-cfasync", "false"),
							h.Src("https://pl28425100.effectivegatecpm.com/3e709d756892597be3b0708e86694b25/invoke.js"),
						),
						h.Div(h.ID("container-3e709d756892597be3b0708e86694b25")),
					),
					h.Div(
						h.Class("forbidden-panel"),
						h.P(h.Class("ad-label"), g.Text("Information")),
						h.P(g.Text("Display and vignette ads keep Arkive accessible.")),
					),
				)),
			),
		),
	)
}

func shouldShowAds(ctx PageContext) bool {
	return ctx.User == nil || !ctx.User.IsPremium
}

func buildForbiddenJS(showAds bool) []string {
	if !showAds {
		return nil
	}
	return nil
}
