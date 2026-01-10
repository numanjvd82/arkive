package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type DropPagesPageProps struct {
	Ctx PageContext
}

func DropPagesPage(props DropPagesPageProps) web.Page {
	_ = props
	return web.Page{
		Title:         "Drop Pages | Arkive",
		Description:   "Create a Drop Page to share anything: files, code, media kits, assignments, or slideshows. One link you can keep updating.",
		CanonicalPath: "/drop-pages",
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
					h.Span(h.Class("seo-eyebrow"), g.Text("Drop Pages")),
					h.H1(g.Text("One Drop Page for anything you want to share.")),
					h.P(g.Text("Drop Pages are shareable pages you can keep updating. Share files, code samples, media kits, portfolios, or class materials with a single link.")),
					h.Div(
						h.Class("seo-actions"),
						h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Create a Drop Page")),
						h.A(h.Class("button secondary"), h.Href("/contact"), g.Text("Talk to us")),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Made for creators, teams, and students.")),
					h.P(g.Text("Drop Pages are flexible. Use them for any content that fits Arkive's guidelines.")),
					h.Div(
						h.Class("seo-grid"),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Developers")),
							h.P(g.Text("Share code snippets, demos, and project files in one page.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Designers")),
							h.P(g.Text("Publish portfolios, mockups, and image sets for clients.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Students")),
							h.P(g.Text("Upload assignments, slideshows, and study packs.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Everything lives in one stable link.")),
					h.P(g.Text("Update content without re-sending links. Keep your Drop Page current as projects evolve.")),
					h.Div(
						h.Class("seo-grid"),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Multi-file collections")),
							h.P(g.Text("Add files, reorder them, and keep your page organized.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Optional privacy")),
							h.P(g.Text("Public, unlisted, or password-protected pages.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("View + download insights")),
							h.P(g.Text("See how often your page and files are accessed.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("How Drop Pages work")),
					h.Div(
						h.Class("seo-steps"),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 1")),
							h.P(g.Text("Create a Drop Page with a title and slug.")),
						),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 2")),
							h.P(g.Text("Add files, reorder them, and set privacy.")),
						),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 3")),
							h.P(g.Text("Share the stable link anywhere.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Drop Pages FAQ")),
					h.Div(
						h.Class("seo-faq"),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Can I use Drop Pages for more than files?")),
							h.P(g.Text("Yes. Use Drop Pages for portfolios, code, media kits, slides, or any content within Arkive guidelines.")),
						),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Will the link stay the same?")),
							h.P(g.Text("Yes. Update your page without changing the URL.")),
						),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Can I make it private?")),
							h.P(g.Text("Yes. Choose unlisted or password-protected pages.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-cta"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Create your first Drop Page.")),
					h.P(g.Text("Start sharing anything with a single link.")),
					h.Div(
						h.Class("seo-actions"),
						h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Create account")),
						h.A(h.Class("button secondary"), h.Href("/pricing"), g.Text("See pricing")),
					),
				),
			),
			marketingFooter(),
		),
	}
}
