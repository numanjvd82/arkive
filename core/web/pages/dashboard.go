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
		CSS:   []string{"/web/pages/dashboard.css", "/web/pages/upload.css", "/static/dialog.css", "/static/tooltip.css"},
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
				components.UploadResumeBanner(components.UploadResumeBannerProps{}),
				h.Section(
					h.Class("dashboard-panels"),
					h.Section(
						h.Class("panel upload-panel"),
						h.Div(
							h.Class("panel-header"),
							h.H2(g.Text("Upload")),
							h.P(g.Text("Fast, resumable uploads with a single click.")),
						),
						components.UploadControls(components.UploadControlsProps{
							InputRequired: true,
						}),
					),
					h.Section(
						h.Class("panel storage-panel"),
						h.Div(
							h.Class("panel-header storage-header"),
							h.H2(g.Text("Storage")),
							components.Tooltip(components.TooltipProps{
								Class:    "storage-tooltip",
								IconName: "info",
								IconSize: "18",
								Tooltip:  fmt.Sprintf("Used: %s\nReserved: %s\nLimit: %s", usedLabel, reservedLabel, limitLabel),
							}),
						),
						components.ProgressBar(components.ProgressBarProps{
							Value: percent,
							Label: "Used " + quotaLabel,
						}),
					),
				),
			),
		),
	}
}
