package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type ShareLargeFilesPageProps struct {
	Ctx PageContext
}

func ShareLargeFilesPage(props ShareLargeFilesPageProps) web.Page {
	_ = props
	return web.Page{
		Title:         "Share Large Files | Arkive",
		Description:   "Share large files fast with stable links, generous storage, and no-login downloads. Built for creators and teams.",
		CanonicalPath: "/share-large-files",
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
					h.Span(h.Class("seo-eyebrow"), g.Text("Large file sharing")),
					h.H1(g.Text("Share large files without limits.")),
					h.P(g.Text("Send big files with stable links, fast delivery, and no-login downloads. Arkive keeps your files available while your account is active.")),
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
					h.H2(g.Text("Built for heavy files and fast delivery.")),
					h.P(g.Text("Arkive is tuned for large transfers with generous storage and a simple sharing flow.")),
					h.Div(
						h.Class("seo-grid"),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("10 GB included")),
							h.P(g.Text("Start free with generous storage and fair use limits.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Fast uploads")),
							h.P(g.Text("Optimized for quick uploads and downloads.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Stable links")),
							h.P(g.Text("Share links that stay live while your account is active.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Drop Pages make large transfers simple.")),
					h.P(g.Text("Drop Pages (Share Pages) bundle multiple large files into a single page you can update without changing the link.")),
					h.Div(
						h.Class("seo-grid"),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("One link for many files")),
							h.P(g.Text("Create a collection for clients or teams.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Keep it updated")),
							h.P(g.Text("Add or reorder files without re-sending links.")),
						),
						h.Div(
							h.Class("seo-card"),
							h.H3(g.Text("Optional password")),
							h.P(g.Text("Protect pages when sharing sensitive work.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("How to share large files")),
					h.Div(
						h.Class("seo-steps"),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 1")),
							h.P(g.Text("Upload large files or create a Drop Page.")),
						),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 2")),
							h.P(g.Text("Copy the share link and send it anywhere.")),
						),
						h.Div(
							h.Class("seo-step"),
							h.Span(g.Text("Step 3")),
							h.P(g.Text("Recipients download without logging in.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-section"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Large file sharing FAQ")),
					h.Div(
						h.Class("seo-faq"),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Can I share multiple large files at once?")),
							h.P(g.Text("Yes. Use Drop Pages to group files into one link.")),
						),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("Do recipients need an account?")),
							h.P(g.Text("No. Anyone with the link can download.")),
						),
						h.Div(
							h.Class("seo-faq-item"),
							h.H3(g.Text("How long are files kept?")),
							h.P(g.Text("Files stay available while your account is active.")),
						),
					),
				),
			),
			h.Section(
				h.Class("seo-cta"),
				h.Div(
					h.Class("container"),
					h.H2(g.Text("Send large files today.")),
					h.P(g.Text("Create an account and start sharing large files in minutes.")),
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
