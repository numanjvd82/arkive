package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type DashboardPageProps struct {
	Ctx PageContext
}

func DashboardPage(props DashboardPageProps) web.Page {
	return web.Page{
		Title: "Arkive · Dashboard",
		CSS:   []string{"/web/pages/dashboard.css"},
		// JS:      []string{"/static/monetag-onclick.js", "/static/monetag-vignette.js"},
		AuthLayout: true,
		User:       props.Ctx.User,
		Body: h.Main(
			h.Class("dashboard"),
			h.Div(
				h.Class("dashboard-shell"),
				h.Div(
					h.Class("dashboard-content"),
					h.Div(
						h.Class("page-header"),
						h.Div(
							h.Class("page-title"),
							h.H1(g.Text("Dashboard")),
							h.P(g.Text("Quick access to uploads and ongoing transfers.")),
						),
					),
					components.UploadResumeBanner(components.UploadResumeBannerProps{}),
					h.Section(
						h.Class("dashboard-panels"),
						h.Section(
							h.Class("panel upload-panel"),
							h.ID("upload-panel"),
							h.Div(
								h.Class("panel-header"),
								h.H2(g.Text("Upload")),
								h.P(g.Text("Fast, resumable uploads with a single click.")),
							),
							components.UploadControls(components.UploadControlsProps{
								InputRequired: true,
							}),
						),
					),
				),
			),
		),
	}
}
