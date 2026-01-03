package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type CookiePageProps struct {
	Ctx PageContext
}

func CookiePage(props CookiePageProps) web.Page {
	_ = props
	return web.Page{
		Title: "Arkive · Cookie Policy",
		CSS:   []string{"/web/pages/cookie.css"},
		Body:  cookieBody(),
	}
}

func cookieBody() g.Node {
	return h.Div(
		h.Class("cookie-page"),
		h.Div(
			h.Class("cookie-hero"),
			h.Div(
				h.Class("container"),
				h.H1(g.Text("Cookie Policy")),
				h.P(h.Class("cookie-meta"), g.Text("Last updated January 02, 2026")),
				h.P(
					h.Class("cookie-intro"),
					g.Text("This Cookie Policy explains how Arkive (arkive.sh) uses cookies and similar technologies on our website and web app."),
				),
				h.P(
					h.Class("cookie-intro"),
					g.Text("Related documents: "),
					h.A(h.Href("/terms"), h.Class("text-link"), g.Text("Terms of Service")),
					g.Text(", "),
					h.A(h.Href("/privacy"), h.Class("text-link"), g.Text("Privacy Policy")),
					g.Text(", and "),
					h.A(h.Href("/aup"), h.Class("text-link"), g.Text("Acceptable Use Policy")),
					g.Text(", and "),
					h.A(h.Href("/abuse"), h.Class("text-link"), g.Text("Copyright & Abuse Reporting")),
					g.Text("."),
				),
			),
		),
		h.Div(
			h.Class("cookie-content"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("cookie-card"),
					cookieSection(
						"1. What cookies are",
						g.Group([]g.Node{
							h.P(g.Text("Cookies are small text files stored on your device. We also use similar technologies like pixels, local storage, or SDKs that serve similar purposes.")),
						}),
					),
					cookieSection(
						"2. How we use cookies",
						g.Group([]g.Node{
							h.P(g.Text("We use cookies that are necessary for the service to function, such as keeping you signed in, preventing abuse, and protecting your session.")),
							h.P(g.Text("We also use privacy-respecting analytics via Cloudflare Web Analytics to understand aggregate traffic and performance. This may involve limited device and log data.")),
						}),
					),
					cookieSection(
						"3. Types of cookies we use",
						g.Group([]g.Node{
							h.P(g.Text("Essential cookies: required for login, security, and core functionality. You cannot opt out of these without disabling the service.")),
							h.P(g.Text("Analytics cookies: used to measure usage and performance. We do not use targeted advertising cookies.")),
						}),
					),
					cookieSection(
						"4. Ads and rewarded offers",
						g.Group([]g.Node{
							h.P(g.Text("We show ads on the website through a third-party ad provider. These ads are not targeted based on personal profiles.")),
							h.P(g.Text("On the website and in the mobile app, rewarded ads may be used to grant temporary storage boosts or rate limit increases. Ad partners may receive a user identifier and basic interaction data to grant rewards and prevent fraud.")),
						}),
					),
					cookieSection(
						"5. Your choices",
						g.Group([]g.Node{
							h.P(g.Text("You can manage cookies through your browser settings. If you block or delete essential cookies, parts of the service may not work.")),
							h.P(g.Text("Some jurisdictions allow you to opt out of certain analytics or advertising technologies. Contact us at support@arkive.sh if you need help with a specific request.")),
						}),
					),
					cookieSection(
						"6. Updates",
						g.Group([]g.Node{
							h.P(g.Text("We may update this Cookie Policy from time to time. The date above shows when it was last revised.")),
						}),
					),
					cookieSection(
						"7. Contact",
						g.Group([]g.Node{
							h.P(
								g.Text("Questions? Email "),
								h.A(h.Href("mailto:support@arkive.sh"), h.Class("text-link"), g.Text("support@arkive.sh")),
								g.Text("."),
							),
						}),
					),
				),
			),
		),
	)
}

func cookieSection(title string, content g.Node) g.Node {
	return h.Section(
		h.Class("cookie-section"),
		h.H2(g.Text(title)),
		content,
	)
}
