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
					h.Form(
						h.Method("post"),
						h.Action("/logout"),
						h.Button(h.Class("button secondary"), h.Type("submit"), g.Text("Logout")),
					),
				),
				h.Section(
					h.Class("dashboard-hero"),
					h.H1(g.Text("Welcome back.")),
					h.P(g.Text("Drag in a file to start a resumable, direct-to-R2 upload.")),
				),
				h.Section(
					h.Class("dashboard-upload"),
					components.Card(components.CardProps{
						Title:    "Upload files",
						Subtitle: "Multipart uploads go straight to R2 with automatic resume.",
						Class:    "upload-card",
						Body: []g.Node{
							components.UploadInput(components.UploadInputProps{
								ID:       "upload-file",
								Name:     "file",
								Label:    "Choose a file",
								Helper:   "We will split it into 10MB chunks (25MB for very large files).",
								Required: true,
							}),
							h.Div(
								h.Class("upload-actions"),
								h.Button(
									h.Class("button primary"),
									h.Type("button"),
									g.Attr("id", "upload-start"),
									g.Text("Start upload"),
								),
								h.Button(
									h.Class("button secondary"),
									h.Type("button"),
									g.Attr("id", "upload-abort"),
									g.Attr("disabled", "disabled"),
									g.Text("Abort"),
								),
							),
							components.ProgressBar(components.ProgressBarProps{
								ID:    "upload-progress",
								Value: 0,
								Label: "Progress",
							}),
							h.P(
								h.Class("upload-status"),
								g.Attr("id", "upload-status"),
								g.Text("No uploads yet."),
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
