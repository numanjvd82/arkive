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
				h.A(h.Href("/secure-file-sharing"), g.Text("Secure sharing")),
				h.A(h.Href("/share-large-files"), g.Text("Large files")),
				h.A(h.Href("/file-sharing-without-login"), g.Text("No-login sharing")),
				h.A(h.Href("/drop-pages"), g.Text("Drop Pages")),
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
				h.A(h.Href("/privacy"), h.Class("text-link"), g.Text("Privacy Policy")),
				h.A(h.Href("/cookies"), h.Class("text-link"), g.Text("Cookie Policy")),
				h.A(h.Href("/terms"), h.Class("text-link"), g.Text("Terms of Service")),
				h.A(h.Href("/aup"), h.Class("text-link"), g.Text("Acceptable Use")),
				h.A(h.Href("/abuse"), h.Class("text-link"), g.Text("Copyright & Abuse")),
			),
		),
	)
}
