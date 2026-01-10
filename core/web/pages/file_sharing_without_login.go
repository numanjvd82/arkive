package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type FileSharingWithoutLoginPageProps struct {
	Ctx PageContext
}

func FileSharingWithoutLoginPage(props FileSharingWithoutLoginPageProps) web.Page {
	_ = props
	return web.Page{
		Title:         "File Sharing Without Login | Arkive",
		Description:   "Share files without forcing recipients to log in. Send a link, add a password if needed, and let anyone download instantly.",
		CanonicalPath: "/file-sharing-without-login",
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
					h.Span(h.Class("seo-eyebrow"), g.Text("No-login sharing")),
					h.H1(g.Text("Share files with anyone, no login required.")),
					h.P(g.Text("Arkive lets you send secure links so recipients can download instantly without creating an account.")),
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
					h.H2(g.Text("No-login downloads with full control.")),
					h.P(g.Text("Recipients do not need an account, but you can still protect links with passwords and manage access.")),
					h.Div(
						h.Class("seo-grid"),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Instant access")),
							h.P(g.Text("Send a link and the file is ready to download.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Password optional")),
							h.P(g.Text("Require a password when you want extra privacy.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("No accounts for viewers")),
							h.P(g.Text("Keep sharing friction-free for clients or teams.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Drop Pages keep everything in one place.")),
					h.P(g.Text("Drop Pages (Share Pages) let you publish a multi-file page with a single link that stays updated.")),
					h.Div(
						h.Class("seo-grid"),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Multi-file pages")),
							h.P(g.Text("Share a full package without sending multiple links.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Update anytime")),
							h.P(g.Text("Replace files without changing the URL.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Share openly or unlisted")),
							h.P(g.Text("Choose public or private sharing modes.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("How no-login sharing works")),
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
							h.P(g.Text("Copy the link and send it to anyone.")),
						),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 3")),
							h.P(g.Text("Recipients download instantly without signing up.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("No-login sharing FAQ")),
					h.Div(
						h.Class("seo-faq"),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Can I still control access?")),
							h.P(g.Text("Yes. You can add passwords or revoke links anytime.")),
						),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Will the link change if I update files?")),
							h.P(g.Text("Drop Pages keep the same URL even when you update files.")),
						),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Do viewers need Arkive?")),
							h.P(g.Text("No. Anyone with the link can access the files.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-cta"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Share without the logins.")),
					h.P(g.Text("Create an Arkive account and send your first link today.")),
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
