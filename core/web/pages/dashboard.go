package pages

import (
	"fmt"
	"math"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
)

type DashboardPageProps struct {
	Ctx            PageContext
	RecentFiles    []models.File
	TotalFiles     int
	UploadSettings models.UploadSettings
}

func DefaultUploadSettings() models.UploadSettings {
	return models.UploadSettings{
		MaxQueueItems: 300,
	}
}

func DashboardPage(props DashboardPageProps) web.Page {
	user := props.Ctx.User
	uploadSettings := props.UploadSettings
	defaultUploadSettings := DefaultUploadSettings()
	if uploadSettings.MaxQueueItems <= 0 {
		uploadSettings.MaxQueueItems = defaultUploadSettings.MaxQueueItems
	}
	usedBytes := int64(0)
	quotaBytes := int64(0)
	isUnlimitedQuota := true
	usagePercent := 0.0
	if user != nil {
		usedBytes = user.UsedBytes
		quotaBytes = user.QuotaBytes
		if quotaBytes > 0 && quotaBytes != math.MaxInt64 {
			isUnlimitedQuota = false
			usagePercent = (float64(usedBytes) / float64(quotaBytes)) * 100
			if usagePercent > 100 {
				usagePercent = 100
			}
		}
	}

	return web.Page{
		Title:              "Arkive · Dashboard",
		Robots:             RobotsNoIndex,
		CSS:                []string{"/web/pages/dashboard.css"},
		AuthLayout:         true,
		RequireVaultUnlock: true,
		User:               props.Ctx.User,
		ActiveNav:          "dashboard",
		SearchPlaceholder:  "Search system...",
		Body: h.Main(
			h.Class("dashboard"),
			h.Div(
				h.Class("container dashboard-shell"),
				components.InlineStyle(components.DataTableCSS),
				h.Div(
					h.Class("dashboard-summary"),
					h.Div(
						h.Class("dashboard-card summary-card"),
						h.Span(h.Class("summary-label"), g.Text("Stored Files")),
						h.Span(h.Class("summary-value summary-value-files"), g.Text(format.Commas(int64(props.TotalFiles)))),
					),
					h.Div(
						h.Class("dashboard-card summary-card"),
						h.Div(
							h.Class("summary-header"),
							h.Span(h.Class("summary-label"), g.Text("Storage Used")),
							h.Span(h.Class("summary-percent"), g.Text(formatPercentText(usagePercent))),
						),
						h.Div(
							h.Class("summary-storage"),
							h.Span(h.Class("summary-value summary-value-storage"), g.Text(format.Bytes(usedBytes))),
							h.Span(
								h.Class("summary-quota"),
								g.If(!isUnlimitedQuota, g.Text("/ "+format.Bytes(quotaBytes))),
								g.If(isUnlimitedQuota, g.Text("/ Unlimited")),
							),
						),
						g.If(!isUnlimitedQuota, h.Div(
							h.Class("dashboard-progress"),
							h.Span(
								h.Class("dashboard-progress-bar"),
								g.Attr("style", "width: "+fmt.Sprintf("%.1f%%", usagePercent)),
							),
						)),
					),
				),
				h.Section(
					h.Class("dashboard-upload"),
					h.ID("upload-panel"),
					components.UploadControls(components.UploadControlsProps{
						InputRequired: true,
						InputLabel:    "Secure Payload Drop",
						InputHelper:   "Drag and drop files here to begin encrypted transfer. All data is zero-knowledge encrypted client-side before transmission.",
						StatusText:    "Select files manually to start a secure upload.",
						MaxQueueItems: uploadSettings.MaxQueueItems,
					}),
				),
				h.Section(
					h.Class("dashboard-activity"),
					h.ID("recent-activity"),
					h.Div(
						h.Class("dashboard-section-header"),
						h.H2(g.Text("Recent Activity")),
					),
					h.Div(
						h.Class("data-table-wrap activity-table-wrap"),
						h.Table(
							h.Class("data-table activity-table"),
							h.THead(
								h.Tr(
									h.Th(g.Text("File Name")),
									h.Th(g.Text("Size")),
									h.Th(g.Text("Modified")),
								),
							),
							h.TBody(
								g.Group(renderDashboardRows(props.RecentFiles)),
							),
						),
					),
				),
			),
		),
	}
}

func renderDashboardRows(files []models.File) []g.Node {
	if len(files) == 0 {
		return []g.Node{
			h.Tr(
				h.Class("activity-empty-row"),
				h.Td(
					g.Attr("colspan", "3"),
					g.Text("No completed files yet."),
				),
			),
		}
	}

	rows := make([]g.Node, 0, len(files))
	for _, file := range files {
		rows = append(rows, h.Tr(
			h.Class("activity-row"),
			h.Td(
				h.Div(
					h.Class("activity-file"),
					fileIcon(),
					h.Span(h.Class("activity-file-name"), g.Text("Encrypted file")),
					lucide.Lock(
						h.Class("dashboard-lucide dashboard-lucide-lock"),
						g.Attr("aria-hidden", "true"),
					),
				),
			),
			h.Td(h.Class("mono"), g.Text("Encrypted")),
			h.Td(h.Class("activity-muted"), g.Text(format.RelativeTime(file.UpdatedAt))),
		))
	}
	return rows
}

func fileIcon() g.Node {
	return lucide.File(h.Class("dashboard-lucide activity-file-icon"), g.Attr("aria-hidden", "true"))
}

func formatPercentText(value float64) string {
	if value <= 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", value)
}
