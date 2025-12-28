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
	limitLabel := format.Bytes(quotaBytes)

	return web.Page{
		Title:   "Arkive · Dashboard",
		CSS:     []string{"/web/pages/dashboard.css"},
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
							h.P(g.Text("Quick access to storage, uploads, and ongoing transfers.")),
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
						h.Section(
							h.Class("panel storage-panel"),
							h.ID("storage-panel"),
							h.Div(
								h.Class("panel-header storage-header"),
								h.H2(g.Text("Storage")),
								components.Tooltip(components.TooltipProps{
									Class:    "storage-tooltip",
									IconName: "info",
									IconSize: "18",
									Tooltip:  fmt.Sprintf("Used: %s\nLimit: %s", usedLabel, limitLabel),
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
		),
	}
}
