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
	usedLabel := format.Bytes(usedBytes)
	reservedLabel := format.Bytes(reservedBytes)
	limitLabel := format.Bytes(quotaBytes)

	return web.Page{
		Title: "Arkive · Dashboard",
		CSS:   []string{"/web/pages/dashboard.css"},
		JS:    []string{"/static/uploads.js"},
		Body: h.Main(
			h.Class("dashboard"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("page-header"),
					h.Div(
						h.Class("page-title"),
						h.H1(g.Text("Dashboard")),
						h.P(g.Text("Quick access to storage, uploads, and ongoing transfers.")),
					),
					h.Div(
						h.Class("page-actions"),
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
					h.Class("dashboard-panels"),
					h.Section(
						h.Class("panel quota-panel"),
						h.Div(
							h.Class("panel-header"),
							h.H2(g.Text("Storage")),
							h.P(g.Text("Track usage across active and in-progress uploads.")),
						),
						h.Div(
							h.Class("quota-summary"),
							h.Span(h.Class("quota-title"), g.Text("Storage used")),
							h.Span(h.Class("quota-value"), g.Text(quotaLabel)),
						),
						components.ProgressBar(components.ProgressBarProps{
							Value: percent,
							Label: "Usage",
						}),
						h.Div(
							h.Class("quota-breakdown"),
							quotaStat("Used", usedLabel),
							quotaStat("Reserved", reservedLabel),
							quotaStat("Limit", limitLabel),
						),
					),
					h.Section(
						h.Class("panel upload-panel"),
						h.Div(
							h.Class("panel-header"),
							h.H2(g.Text("Start an upload")),
							h.P(g.Text("Uploads go straight to R2. Pause or resume when needed.")),
						),
						components.UploadControls(components.UploadControlsProps{
							InputLabel:    "Choose a file",
							InputHelper:   "Up to 1GB. Files over 200MB use multipart chunks.",
							InputRequired: true,
						}),
						h.P(
							h.Class("upload-hint"),
							g.Text("Need to resume? Open Files to continue a pending upload."),
						),
					),
				),
			),
		),
	}
}

func quotaStat(label, value string) g.Node {
	return h.Div(
		h.Class("quota-stat"),
		h.Span(h.Class("stat-label"), g.Text(label)),
		h.Span(h.Class("stat-value"), g.Text(value)),
	)
}
