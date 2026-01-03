package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type AbusePageProps struct {
	Ctx PageContext
}

func AbusePage(props AbusePageProps) web.Page {
	_ = props
	return web.Page{
		Title: "Arkive · Copyright & Abuse Reporting",
		CSS:   []string{"/web/pages/abuse.css"},
		Body:  abuseBody(),
	}
}

func abuseBody() g.Node {
	return h.Div(
		h.Class("abuse-page"),
		h.Div(
			h.Class("abuse-hero"),
			h.Div(
				h.Class("container"),
				h.H1(g.Text("Copyright & Abuse Reporting")),
				h.P(h.Class("abuse-meta"), g.Text("Last updated January 02, 2026")),
				h.P(
					h.Class("abuse-intro"),
					g.Text("This page explains how to report copyright infringement and other abuse on Arkive (arkive.sh). We currently accept reports by email only and do not yet have a registered DMCA agent or mailing address."),
				),
				h.P(
					h.Class("abuse-intro"),
					g.Text("Related documents: "),
					h.A(h.Href("/terms"), h.Class("text-link"), g.Text("Terms of Service")),
					g.Text(", "),
					h.A(h.Href("/privacy"), h.Class("text-link"), g.Text("Privacy Policy")),
					g.Text(", "),
					h.A(h.Href("/cookies"), h.Class("text-link"), g.Text("Cookie Policy")),
					g.Text(", and "),
					h.A(h.Href("/aup"), h.Class("text-link"), g.Text("Acceptable Use Policy")),
					g.Text("."),
				),
			),
		),
		h.Div(
			h.Class("abuse-content"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("abuse-card"),
					abuseSection(
						"1. When to use this page",
						g.Group([]g.Node{
							h.P(g.Text("Use this process to report copyright infringement or abuse such as malware, phishing, child sexual abuse material, non-consensual intimate content, privacy violations, or other unlawful activity.")),
						}),
					),
					abuseSection(
						"2. Copyright notices (DMCA)",
						g.Group([]g.Node{
							h.P(g.Text("Include all of the following: (a) a physical or electronic signature; (b) identification of the copyrighted work you claim is infringed; (c) identification of the material and its location on Arkive (URL or share link); (d) your name, address (if available), telephone number, and email; (e) a statement that you have a good-faith belief the use is not authorized by the rights holder, its agent, or the law; and (f) a statement under penalty of perjury that the information is accurate and you are authorized to act on behalf of the rights holder.")),
							h.P(g.Text("Submitting false or misleading notices may result in liability.")),
						}),
					),
					abuseSection(
						"3. Counter-notices",
						g.Group([]g.Node{
							h.P(g.Text("If your content was removed due to a copyright notice and you believe this was a mistake or misidentification, you may submit a counter-notice that includes: (a) your physical or electronic signature; (b) identification of the removed material and its prior location; (c) a statement under penalty of perjury that you have a good-faith belief the material was removed by mistake or misidentification; and (d) your name, address (if available), telephone number, and email, plus a statement that you consent to the jurisdiction of the federal court in your district (if in the U.S.) or where Arkive may be found (if outside the U.S.), and that you will accept service of process from the original complainant.")),
						}),
					),
					abuseSection(
						"4. Repeat infringer policy",
						g.Group([]g.Node{
							h.P(g.Text("We may terminate accounts of repeat infringers in appropriate circumstances and may restrict sharing or access to enforce this policy.")),
						}),
					),
					abuseSection(
						"5. Abuse reports (non-copyright)",
						g.Group([]g.Node{
							h.P(g.Text("Include as much detail as possible: a description of the issue, URLs or share links, any evidence that supports your report, and your contact information for follow-up questions.")),
						}),
					),
					abuseSection(
						"6. How we respond",
						g.Group([]g.Node{
							h.P(g.Text("We review reports and may remove or disable access to content, restrict sharing, suspend accounts, or request additional information. We may also share information with rights holders or law enforcement when required by law.")),
						}),
					),
					abuseSection(
						"7. Contact",
						g.Group([]g.Node{
							h.P(
								g.Text("Send reports to "),
								h.A(h.Href("mailto:support@arkive.sh"), h.Class("text-link"), g.Text("support@arkive.sh")),
								g.Text(". Please include a clear subject line such as “Copyright Report” or “Abuse Report.”"),
							),
						}),
					),
				),
			),
		),
	)
}

func abuseSection(title string, content g.Node) g.Node {
	return h.Section(
		h.Class("abuse-section"),
		h.H2(g.Text(title)),
		content,
	)
}
