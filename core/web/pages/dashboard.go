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
		Title:   "Arkive · Dashboard",
		CSS:     []string{"/web/pages/dashboard.css"},
		JS:      []string{"/static/monetag-onclick.js", "/static/monetag-vignette.js"},
		HideNav: true,
		Body: h.Main(
			h.Class("dashboard"),
			h.Div(
				h.Class("dashboard-shell"),
				components.DashboardSidebar(),
				h.Div(h.Class("sidebar-scrim"), g.Attr("aria-hidden", "true")),
				h.Div(
					h.Class("dashboard-content"),
					h.Div(
						h.Class("page-header"),
						h.Div(
							h.Class("page-title"),
							h.H1(g.Text("Dashboard")),
							h.P(g.Text("Quick access to uploads and ongoing transfers.")),
						),
						h.Div(
							h.Class("page-actions"),
							h.Button(
								h.Class("button secondary sidebar-toggle"),
								h.Type("button"),
								h.ID("sidebar-toggle"),
								g.Attr("aria-controls", "dashboard-sidebar"),
								g.Attr("aria-expanded", "false"),
								g.Text("Menu"),
							),
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
