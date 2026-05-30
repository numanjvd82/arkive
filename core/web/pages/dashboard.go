package pages

import (
	"fmt"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
)

type DashboardPageProps struct {
	Ctx             PageContext
	RecentFiles     []models.File
	TotalFiles      int
	CurrentFolder   *string
	StorageSettings models.StorageSettings
	UploadSettings  models.UploadSettings
}

func DefaultUploadSettings() models.UploadSettings {
	return models.UploadSettings{
		MaxQueueItems:    300,
		PartConcurrency:  3,
		StaleUploadHours: 1,
	}
}

func DashboardPage(props DashboardPageProps) web.Page {
	user := props.Ctx.User
	uploadSettings := props.UploadSettings
	currentFolderAttr := g.Node(nil)
	if props.CurrentFolder != nil {
		currentFolderAttr = g.Attr("data-current-folder-id", *props.CurrentFolder)
	}
	defaultUploadSettings := DefaultUploadSettings()
	if uploadSettings.MaxQueueItems <= 0 {
		uploadSettings.MaxQueueItems = defaultUploadSettings.MaxQueueItems
	}
	if uploadSettings.PartConcurrency <= 0 {
		uploadSettings.PartConcurrency = defaultUploadSettings.PartConcurrency
	}
	if uploadSettings.StaleUploadHours <= 0 {
		uploadSettings.StaleUploadHours = defaultUploadSettings.StaleUploadHours
	}
	usedBytes := int64(0)
	maxStorageBytes := props.StorageSettings.MaxStorageBytes
	isUnlimitedQuota := true
	usagePercent := 0.0
	if user != nil {
		usedBytes = user.UsedBytes
		if maxStorageBytes > 0 {
			isUnlimitedQuota = false
			usagePercent = (float64(usedBytes) / float64(maxStorageBytes)) * 100
			if usagePercent > 100 {
				usagePercent = 100
			}
		}
	}

	return web.Page{
		Title:              "Arkive · Vault",
		Robots:             RobotsNoIndex,
		CSS:                []string{"/web/pages/dashboard.css"},
		JS:                 []string{"/static/dashboard.js"},
		AuthLayout:         true,
		RequireVaultUnlock: true,
		User:               props.Ctx.User,
		ActiveNav:          "dashboard",
		SearchPlaceholder:  "Search vault...",
		Body: h.Main(
			h.Class("dashboard"),
			currentFolderAttr,
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
								g.If(!isUnlimitedQuota, g.Text("/ "+format.Bytes(maxStorageBytes))),
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
						InputRequired:   true,
						InputLabel:      "Secure Upload",
						InputHelper:     "Drag and drop files here to begin encrypted upload. Data is encrypted in browser before transfer.",
						StatusText:      "Select files manually to start a secure upload.",
						MaxQueueItems:   uploadSettings.MaxQueueItems,
						PartConcurrency: uploadSettings.PartConcurrency,
					}),
				),
				h.Section(
					h.Class("dashboard-activity"),
					h.ID("recent-activity"),
					h.Div(
						h.Class("dashboard-section-header"),
						h.H2(g.Text("Recent Files")),
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
		viewURL := fmt.Sprintf("/files/%s/view", file.ID)
		rows = append(rows, h.Tr(
			h.Class("activity-row"),
			g.Attr("aria-busy", "true"),
			g.Attr("data-activity-open", viewURL),
			g.Attr("data-file-item", file.ID),
			g.Attr("data-file-name", ""),
			g.Attr("tabindex", "0"),
			h.Td(
				h.Div(
					h.Class("activity-file"),
					fileIcon(),
					h.Span(
						h.Class("activity-file-name activity-skeleton activity-skeleton-name"),
						g.Attr("data-file-field", "name"),
						g.Attr("aria-hidden", "true"),
					),
					lucide.Lock(
						h.Class("dashboard-lucide dashboard-lucide-lock"),
						g.Attr("aria-hidden", "true"),
					),
				),
			),
			h.Td(
				h.Class("mono activity-size"),
				h.Span(
					g.Attr("data-file-field", "size"),
					g.Text("Encrypted"),
				),
			),
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
