package pages

import (
	"fmt"
	"strings"
	"time"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/services/shares"
	"arkive/core/web"
	"arkive/core/web/components"
)

type SharesPageProps struct {
	Ctx    PageContext
	Shares []models.ShareWithFile
}

func SharesPage(props SharesPageProps) web.Page {
	now := time.Now()
	total := len(props.Shares)
	activeCount := 0
	expiredCount := 0
	revokedCount := 0
	expiringSoonCount := 0
	for _, item := range props.Shares {
		status, expired := shareStatus(item, now)
		switch status {
		case shares.ShareStatusRevoked:
			revokedCount++
		case shares.ShareStatusActive:
			if expired {
				expiredCount++
			} else {
				activeCount++
			}
		}
		if status == shares.ShareStatusActive && !expired && item.ExpiresAt != nil && item.ExpiresAt.After(now) && item.ExpiresAt.Before(now.Add(7*24*time.Hour)) {
			expiringSoonCount++
		}
	}

	return web.Page{
		Title:      "Arkive · Shares",
		Robots:     RobotsNoIndex,
		CSS:        []string{"/web/pages/shares.css"},
		JS:         []string{"/static/shares.js"},
		AuthLayout: true,
		User:       props.Ctx.User,
		ActiveNav:  "shares",
		Body: g.Group([]g.Node{
			components.InlineStyle(components.DataTableCSS),
			h.Main(
				h.Class("shares-page"),
				h.Div(
					h.Class("container"),
					h.Div(
						h.Class("page-header"),
						h.Div(
							h.Class("page-title"),
							h.H1(g.Text("Public Shares")),
						),
					),
					h.Section(
						h.Class("shares-summary"),
						renderShareSummaryCard(
							"Total Active Shares",
							fmt.Sprintf("%d", activeCount),
							"primary",
							lucide.Share2(
								h.Class("shares-lucide shares-lucide-summary"),
								g.Attr("aria-hidden", "true"),
							),
						),
						renderShareSummaryCard(
							"Restricted Links",
							fmt.Sprintf("%d", expiredCount+revokedCount),
							"warning",
							lucide.Lock(
								h.Class("shares-lucide shares-lucide-summary"),
								g.Attr("aria-hidden", "true"),
							),
						),
					),
					renderShareList(props.Shares, total, expiringSoonCount, revokedCount),
				),
			),
			components.Dialog(components.DialogProps{
				BackdropID: "share-action-backdrop",
				TitleID:    "share-action-title",
				Title:      "Update share?",
				Body:       h.P(g.Attr("id", "share-action-meta"), g.Text("")),
				Actions: g.Group([]g.Node{
					h.Button(
						h.Class("button secondary"),
						h.Type("button"),
						g.Attr("id", "share-action-cancel"),
						g.Text("Cancel"),
					),
					h.Button(
						h.Class("button danger"),
						h.Type("button"),
						g.Attr("id", "share-action-confirm"),
						g.Text("Confirm"),
					),
				}),
			}),
		}),
	}
}

func renderShareSummaryCard(title, value, variant string, icon g.Node) g.Node {
	return h.Div(
		h.Class("shares-summary-card shares-summary-card-"+variant),
		h.Span(
			h.Class("shares-summary-icon"),
			icon,
		),
		h.Div(
			h.Class("shares-summary-copy"),
			h.Span(h.Class("shares-summary-title"), g.Text(title)),
			h.Span(h.Class("shares-summary-value"), g.Text(value)),
		),
	)
}

func renderShareList(items []models.ShareWithFile, total, expiringSoonCount, revokedCount int) g.Node {
	if len(items) == 0 {
		return h.Div(
			h.Class("data-table-wrap shares-table-wrap"),
			h.Div(
				h.Class("shares-empty"),
				lucide.Share2(
					h.Class("shares-lucide shares-lucide-empty"),
					g.Attr("aria-hidden", "true"),
				),
				h.H2(g.Text("No shared links yet.")),
				h.P(g.Text("Create a public link from the files page to start sharing.")),
				h.A(
					h.Class("shares-empty-link"),
					h.Href("/files"),
					g.Text("Go to files"),
				),
			),
		)
	}

	rows := make([]g.Node, 0, len(items))
	now := time.Now()
	for _, item := range items {
		status, expired := shareStatus(item, now)
		statusLabel := status
		if status == shares.ShareStatusActive && expired {
			statusLabel = "expired"
		}
		expiresLabel := "Never"
		if item.ExpiresAt != nil {
			expiresLabel = item.ExpiresAt.Format("Jan 2, 2006")
		}
		sharePath := "/s/" + item.Token
		shareDisplay := compactSharePath(sharePath)
		statusTone := statusLabel
		if statusLabel == shares.ShareStatusActive && item.PasswordHash != nil {
			statusTone = "restricted"
		}

		rows = append(rows, h.Tr(
			h.Class("shares-row"),
			g.Attr("data-share-row", item.ID),
			g.Attr("id", "share-"+item.ID),
			h.Td(
				h.Class("shares-file"),
				shareFileIcon(item),
				h.Span(h.Class("shares-name"), g.Text(item.FileName)),
			),
			h.Td(
				h.Class("shares-status"),
				h.Span(
					h.Class("status-pill "+statusTone),
					g.Attr("data-share-status", item.ID),
					g.Text(titleCase(statusLabel)),
				),
			),
			h.Td(h.Class("shares-expiry"), g.Text(expiresLabel)),
			h.Td(
				h.Class("shares-link-cell"),
				h.Span(
					h.Class("shares-link-text"),
					h.Title(sharePath),
					g.Text(shareDisplay),
				),
			),
			h.Td(
				h.Class("shares-actions-cell"),
				h.Div(
					h.Class("shares-actions"),
					h.Button(
						h.Class("shares-copy-button"),
						h.Type("button"),
						g.Attr("data-share-copy", sharePath),
						g.Attr("aria-label", "Copy share link"),
						lucide.Copy(
							h.Class("shares-lucide shares-lucide-copy"),
							g.Attr("aria-hidden", "true"),
						),
					),
					h.Button(
						h.Class("shares-action-link"),
						h.Type("button"),
						g.Attr("data-share-action", "delete"),
						g.Attr("data-share-id", item.ID),
						g.Attr("data-share-file", item.FileName),
						g.Text("Delete"),
					),
				),
			),
		))
	}

	return h.Div(
		h.Class("data-table-wrap shares-table-wrap"),
		h.Table(
			h.Class("data-table shares-table"),
			h.THead(
				h.Tr(
					h.Th(g.Text("File Name")),
					h.Th(g.Text("Status")),
					h.Th(g.Text("Expiry Date")),
					h.Th(g.Text("Share URL")),
					h.Th(h.Class("shares-align-right"), g.Text("Actions")),
				),
			),
			h.TBody(g.Group(rows)),
		),
	)
}

func shareStatus(item models.ShareWithFile, now time.Time) (string, bool) {
	if item.Status == shares.ShareStatusRevoked {
		return shares.ShareStatusRevoked, false
	}
	if item.ExpiresAt != nil && !item.ExpiresAt.After(now) {
		return shares.ShareStatusActive, true
	}
	return shares.ShareStatusActive, false
}

func fileTypeLabelFromShare(item models.ShareWithFile) string {
	name := strings.TrimSpace(item.FileName)
	ext := ""
	if name != "" {
		parts := strings.Split(name, ".")
		if len(parts) > 1 {
			ext = parts[len(parts)-1]
		}
	}
	if ext == "" {
		content := strings.TrimSpace(item.FileContentType)
		if strings.Contains(content, "/") {
			ext = strings.Split(content, "/")[1]
		}
	}
	ext = strings.ToUpper(ext)
	if ext == "" || len(ext) > 5 {
		return "FILE"
	}
	return ext
}

func shareFileIcon(item models.ShareWithFile) g.Node {
	contentType := strings.TrimSpace(strings.ToLower(item.FileContentType))
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return lucide.Image(h.Class("shares-lucide shares-lucide-file"), g.Attr("aria-hidden", "true"))
	case strings.Contains(contentType, "zip") || strings.Contains(contentType, "tar"):
		return lucide.FileArchive(h.Class("shares-lucide shares-lucide-file"), g.Attr("aria-hidden", "true"))
	case item.PasswordHash != nil:
		return lucide.Lock(h.Class("shares-lucide shares-lucide-file"), g.Attr("aria-hidden", "true"))
	default:
		return lucide.FileText(h.Class("shares-lucide shares-lucide-file"), g.Attr("aria-hidden", "true"))
	}
}

func compactSharePath(path string) string {
	path = strings.TrimSpace(path)
	if len(path) <= 18 {
		return path
	}
	return path[:8] + "..." + path[len(path)-6:]
}

func titleCase(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
