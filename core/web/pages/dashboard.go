package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

func DashboardPage() web.Page {
	return web.Page{
		Title: "Arkive · Dashboard",
		CSS:   []string{"/web/pages/dashboard.css"},
		JS:    []string{"/static/uploads.js"},
		Body: h.Main(
			h.Class("dashboard"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("dashboard-header"),
					h.Div(
						h.Class("dashboard-brand"),
						h.Span(h.Class("logo-dot")),
						h.Span(h.Class("logo-text"), g.Text("Arkive")),
					),
					h.Div(
						h.Class("dashboard-actions"),
						components.Button(components.ButtonProps{
							Text:    "Files",
							Href:    "/files",
							Variant: "secondary",
						}),
						h.Form(
							h.Method("post"),
							h.Action("/logout"),
							h.Button(h.Class("button secondary"), h.Type("submit"), g.Text("Logout")),
						),
					),
				),
				h.Section(
					h.Class("dashboard-hero"),
					h.H1(g.Text("Welcome back.")),
					h.P(g.Text("Start a direct-to-R2 upload. Resume from the Files page anytime.")),
				),
				h.Section(
					h.Class("dashboard-upload"),
					components.Card(components.CardProps{
						Title:    "Upload files",
						Subtitle: "Multipart uploads go straight to R2. Resume from the Files page if needed.",
						Class:    "upload-card",
						Body: []g.Node{
							components.UploadControls(components.UploadControlsProps{
								InputLabel:    "Choose a file",
								InputHelper:   "Max 1GB. Files over 200MB use multipart chunks.",
								InputRequired: true,
							}),
							h.P(
								h.Class("upload-status"),
								g.Text("Need to resume? Open Files to continue any pending upload."),
							),
						},
					}),
				),
				h.Div(
					h.Class("dashboard-grid"),
					dashboardCard("Collections", "3 active"),
					dashboardCard("Shared links", "12 saved"),
					dashboardCard("Last updated", "2 hours ago"),
				),
			),
		),
	}
}

func dashboardCard(title, body string) g.Node {
	return h.Div(
		h.Class("dashboard-card"),
		h.Span(h.Class("dashboard-label"), g.Text(title)),
		h.P(h.Class("dashboard-value"), g.Text(body)),
	)
}
