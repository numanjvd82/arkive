package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type PricingPageProps struct {
	Ctx PageContext
}

func PricingPage(props PricingPageProps) web.Page {
	_ = props
	return web.Page{
		Title:         "Arkive · Fair Use",
		Description:   "See Arkive fair use policy for our storage-first file sharing platform.",
		CanonicalPath: "/pricing",
		OGImage:       DefaultOGImage,
		Robots:        RobotsIndex,
		CSS:           []string{"/web/pages/pricing.css"},
		Body:          pricingBody(),
	}
}

func pricingBody() g.Node {
	return h.Div(
		h.Class("pricing-page"),
		h.Div(
			h.Class("pricing-hero"),
			h.Div(
				h.Class("container"),
				h.H1(g.Text("Fair Use Policy")),
				h.P(h.Class("pricing-meta"), g.Text("Last updated April 28, 2026")),
				h.P(
					h.Class("pricing-intro"),
					g.Text("Arkive is a storage-first file sharing platform. Fair use limits help prevent abuse and keep performance consistent for everyone."),
				),
			),
		),
		h.Div(
			h.Class("pricing-content"),
			h.Div(
				h.Class("container"),
				h.Section(
					h.Class("pricing-card fairuse-card"),
					h.Ul(
						h.Class("fairuse-list"),
						h.Li(g.Text("Accounts with no login or file activity for 30 days may be archived.")),
						h.Li(g.Text("Activity means owner logins or authenticated actions; public link views do not count.")),
						h.Li(g.Text("Archived files are deleted after 7 days.")),
						h.Li(g.Text("We email you before archiving or deleting files.")),
						h.Li(g.Text("Abuse protection: no spam, link farms, scraping, hotlinking misuse, or using Arkive as a public mirror.")),
						h.Li(g.Text("Illegal content or repeated abuse can result in suspension.")),
					),
				),
			),
		),
	)
}
