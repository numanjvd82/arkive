package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	. "arkive/core/web"
)

func HomePage() g.Node {
	return Layout(LayoutData{
		Title: "Arkive",
		CSS:   "/static/pages/home.css",
	},
		h.Main(
			h.Class("hero"),
			h.Div(
				h.Class("container"),
				h.H1(g.Text("Keep your artifacts organized, searchable, and shareable.")),
				h.P(g.Text("Arkive is your home for project references, notes, and documents. Stay focused while everything important stays within reach.")),
				h.Div(
					h.Class("hero-actions"),
					h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Get Started")),
					h.A(h.Class("button ghost"), h.Href("/login"), g.Text("I already have an account")),
				),
			),
		),
		h.Section(
			h.Class("features"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("feature"),
					h.Img(h.Src("/static/images/feature-search.svg"), h.Alt("Search feature illustration")),
					h.H2(g.Text("Powerful Search")),
					h.P(g.Text("Quickly find what you need with our advanced search capabilities.")),
				),
				h.Div(
					h.Class("feature"),
					h.Img(h.Src("/static/images/feature-organization.svg"), h.Alt("Organization feature illustration")),
					h.H2(g.Text("Organized Storage")),
					h.P(g.Text("Keep your artifacts neatly organized for easy access.")),
				),
				h.Div(
					h.Class("feature"),
					h.Img(h.Src("/static/images/feature-sharing.svg"), h.Alt("Sharing feature illustration")),
					h.H2(g.Text("Easy Sharing")),
					h.P(g.Text("Share your collections with colleagues and collaborators effortlessly.")),
				),
			),
		),
	)
}
