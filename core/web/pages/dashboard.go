package pages

import (
	"fmt"
	"math"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
)

type DashboardPageProps struct {
	Ctx PageContext
}

func DashboardPage(props DashboardPageProps) web.Page {
	var usedBytes int64
	var reservedBytes int64
	var quotaBytes int64
	if props.Ctx.User != nil {
		usedBytes = props.Ctx.User.UsedBytes
		reservedBytes = props.Ctx.User.ReservedBytes
		quotaBytes = props.Ctx.User.QuotaBytes
	}

	totalUsed := usedBytes + reservedBytes
	percent := 0
	if quotaBytes > 0 {
		percent = int(math.Round(float64(totalUsed) * 100 / float64(quotaBytes)))
		if percent < 0 {
			percent = 0
		}
		if percent > 100 {
			percent = 100
		}
	}
	quotaLabel := fmt.Sprintf("%s / %s", format.Bytes(totalUsed), format.Bytes(quotaBytes))

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
					h.Class("dashboard-quota"),
					h.Div(
						h.Class("quota-meta"),
						h.Span(h.Class("quota-label"), g.Text("Storage used")),
						h.Span(h.Class("quota-value"), g.Text(quotaLabel)),
					),
					h.Div(
						h.Class("quota-progress"),
						components.ProgressBar(components.ProgressBarProps{
							Value: percent,
						}),
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
