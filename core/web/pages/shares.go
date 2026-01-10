package pages

import (
	"fmt"
	"strings"
	"time"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/services/shares"
	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
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
		JS:         []string{"/static/shares.js", "/static/monetag-onclick.js", "/static/monetag-vignette.js"},
		AuthLayout: true,
		User:       props.Ctx.User,
		Body: h.Main(
			h.Class("shares-page"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("page-header"),
					h.Div(
						h.Class("page-title"),
						h.H1(g.Text("Shares")),
						h.P(g.Text("Track every public link you have created and manage access.")),
					),
					h.Div(
						h.Class("page-actions"),
						components.Button(components.ButtonProps{
							Text:    "Share a file",
							Href:    "/files",
							Variant: "primary",
						}),
						components.Button(components.ButtonProps{
							Text:    "Files",
							Href:    "/files",
							Variant: "secondary",
						}),
					),
				),
				h.Section(
					h.Class("shares-summary"),
					components.Card(components.CardProps{
						Title:    "Total shares",
						Subtitle: "All time",
						Class:    "summary-card",
						Body: []g.Node{
							h.Span(h.Class("summary-value"), g.Text(fmt.Sprintf("%d", total))),
							h.Span(h.Class("summary-meta"), g.Text("Created links")),
						},
					}),
					components.Card(components.CardProps{
						Title:    "Active",
						Subtitle: "Publicly available",
						Class:    "summary-card",
						Body: []g.Node{
							h.Span(h.Class("summary-value"), g.Text(fmt.Sprintf("%d", activeCount))),
							h.Span(h.Class("summary-meta"), g.Text(fmt.Sprintf("%d expiring soon", expiringSoonCount))),
						},
					}),
					components.Card(components.CardProps{
						Title:    "Restricted",
						Subtitle: "Expired or revoked",
						Class:    "summary-card",
						Body: []g.Node{
							h.Span(h.Class("summary-value"), g.Text(fmt.Sprintf("%d", expiredCount+revokedCount))),
							h.Span(h.Class("summary-meta"), g.Text(fmt.Sprintf("%d revoked", revokedCount))),
						},
					}),
				),
				h.Section(
					h.Class("shares-panels"),
					h.Section(
						h.Class("panel shares-list"),
						h.Div(
							h.Class("panel-header"),
							h.H2(g.Text("Shared files")),
							h.P(g.Text("Links are unique per file and update instantly.")),
						),
						renderShareList(props.Shares),
					),
				),
			),
			components.Dialog(components.DialogProps{
				BackdropID: "share-action-backdrop",
				TitleID:    "share-action-title",
				Title:      "Update share?",
				Body:       h.P(g.Attr("id", "share-action-meta"), g.Text("")),
				Actions: h.Div(
					h.Class("dialog-actions"),
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
				),
			}),
		),
	}
}

func renderShareList(items []models.ShareWithFile) g.Node {
	if len(items) == 0 {
		return h.Div(
			h.Class("shares-empty"),
			h.P(g.Text("No shared links yet.")),
			components.Button(components.ButtonProps{
				Text:    "Share a file",
				Href:    "/files",
				Variant: "secondary",
			}),
		)
	}

	rows := make([]g.Node, 0, len(items)+1)
	rows = append(rows, h.Div(
		h.Class("shares-row shares-row-head"),
		h.Span(g.Text("File")),
		h.Span(g.Text("Status")),
		h.Span(g.Text("Expiry")),
		h.Span(g.Text("Link")),
		h.Span(g.Text("Actions")),
	))

	now := time.Now()
	for _, item := range items {
		status, expired := shareStatus(item, now)
		statusLabel := status
		if status == shares.ShareStatusActive && expired {
			statusLabel = "expired"
		}
		statusClass := "status-pill " + statusLabel
		expiresLabel := "Never"
		if item.ExpiresAt != nil {
			expiresLabel = item.ExpiresAt.Format("Jan 2, 2006")
		}
		protectedLabel := "Public"
		if item.PasswordHash != nil {
			protectedLabel = "Password"
		}
		sharePath := "/s/" + item.Token

		rows = append(rows, h.Div(
			h.Class("shares-row"),
			g.Attr("data-share-row", item.ID),
			h.Div(
				h.Class("shares-file"),
				h.Span(h.Class("shares-badge"), g.Text(fileTypeLabelFromShare(item))),
				h.Div(
					h.Class("shares-meta"),
					h.Span(h.Class("shares-name"), g.Text(item.FileName)),
					h.Span(h.Class("shares-sub"), g.Text(fmt.Sprintf("%s • %s", protectedLabel, format.Bytes(item.FileSizeBytes)))),
				),
			),
			h.Div(
				h.Class("shares-status"),
				h.Span(h.Class(statusClass), g.Attr("data-share-status", item.ID), g.Text(titleCase(statusLabel))),
				h.Span(h.Class("shares-sub"), g.Text("Created "+formatTime(item.CreatedAt))),
			),
			h.Span(h.Class("shares-expiry"), g.Text(expiresLabel)),
			h.Div(
				h.Class("shares-link"),
				h.Span(h.Class("shares-link-text"), g.Text(sharePath)),
				h.Button(
					h.Class("button secondary shares-copy"),
					h.Type("button"),
					g.Attr("data-share-copy", sharePath),
					g.Attr("aria-label", "Copy share link"),
					g.Text("Copy"),
				),
			),
			h.Div(
				h.Class("shares-actions"),
				h.A(
					h.Class("button secondary"),
					h.Href(sharePath),
					g.Attr("target", "_blank"),
					g.Attr("rel", "noreferrer"),
					g.Text("Open"),
				),
				g.If(statusLabel == shares.ShareStatusActive, h.Button(
					h.Class("button secondary"),
					h.Type("button"),
					g.Attr("data-share-action", "revoke"),
					g.Attr("data-share-id", item.ID),
					g.Attr("data-share-file", item.FileName),
					g.Text("Revoke"),
				)),
				h.Button(
					h.Class("button danger"),
					h.Type("button"),
					g.Attr("data-share-action", "delete"),
					g.Attr("data-share-id", item.ID),
					g.Attr("data-share-file", item.FileName),
					g.Text("Delete"),
				),
			),
		))
	}

	return h.Div(h.Class("shares-rows"), g.Group(rows))
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

func titleCase(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
