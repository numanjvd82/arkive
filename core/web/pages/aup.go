package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type AUPPageProps struct {
	Ctx PageContext
}

func AUPPage(props AUPPageProps) web.Page {
	_ = props
	return web.Page{
		Title: "Arkive · Acceptable Use Policy",
		CSS:   []string{"/web/pages/aup.css"},
		Body:  aupBody(),
	}
}

func aupBody() g.Node {
	return h.Div(
		h.Class("aup-page"),
		h.Div(
			h.Class("aup-hero"),
			h.Div(
				h.Class("container"),
				h.H1(g.Text("Acceptable Use Policy")),
				h.P(h.Class("aup-meta"), g.Text("Last updated January 02, 2026")),
				h.P(
					h.Class("aup-intro"),
					g.Text("This Acceptable Use Policy (AUP) describes prohibited activities when using Arkive. Violations may result in suspension or termination."),
				),
				h.P(
					h.Class("aup-intro"),
					g.Text("Related documents: "),
					h.A(h.Href("/terms"), h.Class("text-link"), g.Text("Terms of Service")),
					g.Text(", "),
					h.A(h.Href("/privacy"), h.Class("text-link"), g.Text("Privacy Policy")),
					g.Text(", and "),
					h.A(h.Href("/cookies"), h.Class("text-link"), g.Text("Cookie Policy")),
					g.Text(", and "),
					h.A(h.Href("/abuse"), h.Class("text-link"), g.Text("Copyright & Abuse Reporting")),
					g.Text("."),
				),
			),
		),
		h.Div(
			h.Class("aup-content"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("aup-card"),
					aupSection(
						"1. Prohibited content",
						g.Group([]g.Node{
							h.P(g.Text("Do not upload, share, or distribute content that is illegal, harmful, or infringes the rights of others.")),
							h.P(g.Text("This includes child sexual abuse material, non-consensual intimate content, malware, phishing content, copyright-infringing files, and content that violates privacy or data protection laws.")),
						}),
					),
					aupSection(
						"2. Prohibited behavior",
						g.Group([]g.Node{
							h.P(g.Text("Do not attempt to gain unauthorized access to accounts, systems, or data.")),
							h.P(g.Text("Do not disrupt the service, probe or scan for vulnerabilities, or interfere with rate limits or security measures.")),
							h.P(g.Text("Do not use Arkive to send spam, run automated scraping, or abuse shared links.")),
						}),
					),
					aupSection(
						"3. Abuse reporting",
						g.Group([]g.Node{
							h.P(
								g.Text("If you believe content violates this policy, report it to "),
								h.A(h.Href("mailto:support@arkive.sh"), h.Class("text-link"), g.Text("support@arkive.sh")),
								g.Text(" or use "),
								h.A(h.Href("/abuse"), h.Class("text-link"), g.Text("Copyright & Abuse Reporting")),
								g.Text(" for required details."),
							),
						}),
					),
					aupSection(
						"4. Enforcement",
						g.Group([]g.Node{
							h.P(g.Text("We may remove content, restrict sharing, suspend rewards, or terminate accounts when we detect or receive reports of abuse.")),
						}),
					),
					aupSection(
						"5. Changes",
						g.Group([]g.Node{
							h.P(g.Text("We may update this AUP from time to time. The date above shows when it was last revised.")),
						}),
					),
				),
			),
		),
	)
}

func aupSection(title string, content g.Node) g.Node {
	return h.Section(
		h.Class("aup-section"),
		h.H2(g.Text(title)),
		content,
	)
}
