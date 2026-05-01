package pages

import (
	"fmt"
	"time"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web/components"
)

func marketingHeader() g.Node {
	return h.Header(
		h.Class("site-header"),
		h.Div(
			h.Class("container nav"),
			components.BrandLogo(components.BrandLogoProps{
				Href:  "/",
				Class: "nav-brand",
			}),
			h.Nav(
				h.Class("nav-links"),
				h.A(h.Href("/pricing"), g.Text("Pricing")),
				h.A(h.Href("/contact"), g.Text("Contact")),
				h.A(h.Href("/#features"), g.Text("Features")),
				h.A(h.Href("/#sharing"), g.Text("Sharing")),
				h.A(h.Href("/#security"), g.Text("Security")),
				h.A(h.Href("/#roadmap"), g.Text("Roadmap")),
			),
			h.Div(
				h.Class("nav-actions"),
				h.A(h.Class("button secondary"), h.Href("/login"), g.Text("Log in")),
				h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Create account")),
			),
		),
	)
}

func marketingFooter() g.Node {
	return h.Footer(
		h.Class("site-footer"),
		h.Div(
			h.Class("container footer-grid"),
			h.Div(
				h.Class("footer-brand"),
				h.H3(g.Text("Arkive")),
				h.P(g.Text("Share files with freedom, speed, and security.")),
				h.P(
					h.Class("footer-legal"),
					g.Text(fmt.Sprintf("© %d Arkive. All rights reserved.", time.Now().Year())),
				),
			),
			h.Div(
				h.Class("footer-links"),
				h.A(h.Href("/pricing"), g.Text("Pricing")),
				h.A(h.Href("/#features"), g.Text("Features")),
				h.A(h.Href("/#sharing"), g.Text("Sharing")),
				h.A(h.Href("/#security"), g.Text("Security")),
				h.A(h.Href("/#roadmap"), g.Text("Roadmap")),
			),
			h.Div(
				h.Class("footer-links"),
				h.A(h.Href("/login"), g.Text("Login")),
				h.A(h.Href("/signup"), g.Text("Create account")),
				h.A(h.Href("/contact"), g.Text("Contact")),
	
			),
		),
	)
}
