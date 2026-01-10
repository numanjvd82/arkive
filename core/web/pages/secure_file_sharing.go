package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type SecureFileSharingPageProps struct {
	Ctx PageContext
}

func SecureFileSharingPage(props SecureFileSharingPageProps) web.Page {
	_ = props
	return web.Page{
		Title:         "Secure File Sharing | Arkive",
		Description:   "Secure file sharing with password protection, privacy-first defaults, and fast delivery. Share with anyone without losing control.",
		CanonicalPath: "/secure-file-sharing",
		OGImage:       DefaultOGImage,
		Robots:        RobotsIndex,
		CSS:           []string{"/web/pages/home.css", "/web/pages/seo.css"},
		HideNav:       true,
		Body: h.Div(
			h.Class("seo-page"),
			marketingHeader(),
			h.Section(
				h.Class("seo-hero"),
				h.Div(
					h.Class("container"),
					h.Span(h.Class("seo-eyebrow"), g.Text("Secure sharing")),
					h.H1(g.Text("Secure file sharing built for speed.")),
					h.P(g.Text("Arkive protects every share link with privacy-first defaults, optional passwords, and encrypted storage. Share instantly and keep control.")),
					h.Div(
						h.Class("seo-actions"),
						h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Start sharing")),
						h.A(h.Class("button secondary"), h.Href("/pricing"), g.Text("See pricing")),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Security features that stay out of your way.")),
					h.P(g.Text("From encrypted storage to optional passwords, Arkive is designed to keep links safe without adding friction for recipients.")),
					h.Div(
						h.Class("seo-grid"),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Password-protected links")),
							h.P(g.Text("Lock any share link with a password in seconds.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Encrypted storage")),
							h.P(g.Text("Files stay protected at rest while you control access.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Share without logins")),
							h.P(g.Text("Recipients can download without creating an account.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Drop Pages keep secure links stable.")),
					h.P(g.Text("Drop Pages (Share Pages) let you publish multiple files on a single page and update them over time without changing the link.")),
					h.Div(
						h.Class("seo-grid"),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("One secure link")),
							h.P(g.Text("Share a single URL that stays valid as you update files.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Optional passwords")),
							h.P(g.Text("Protect entire pages with a password when needed.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("View + download stats")),
							h.P(g.Text("Track views and downloads to stay in control.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("How secure sharing works")),
					h.Div(
						h.Class("seo-steps"),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 1")),
							h.P(g.Text("Upload your file or create a Drop Page.")),
						),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 2")),
							h.P(g.Text("Add a password or keep the link open.")),
						),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 3")),
							h.P(g.Text("Share instantly and let recipients download.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Secure file sharing FAQ")),
					h.Div(
						h.Class("seo-faq"),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Can I share securely without forcing logins?")),
							h.P(g.Text("Yes. Arkive links can be password protected while still allowing no-login downloads.")),
						),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Do secure links expire?")),
							h.P(g.Text("You can set expiry times, or keep links active while your account is active.")),
						),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("What is a Drop Page?")),
							h.P(g.Text("A Drop Page is a multi-file share page you can update without changing the link.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-cta"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Ready to share securely?")),
					h.P(g.Text("Create an account and start sharing secure links in seconds.")),
					h.Div(
						h.Class("seo-actions"),
						h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Create account")),
						h.A(h.Class("button secondary"), h.Href("/contact"), g.Text("Contact sales")),
					),
				),
			),
			marketingFooter(),
		),
	}
}
