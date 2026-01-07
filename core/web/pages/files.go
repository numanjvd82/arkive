package pages

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
)

type FilesPageProps struct {
	Ctx   PageContext
	Files []models.File
}

func FilesPage(props FilesPageProps) web.Page {
	totalSize := int64(0)
	lastUpdated := time.Time{}
	for _, file := range props.Files {
		totalSize += file.SizeBytes
		if file.UpdatedAt.After(lastUpdated) {
			lastUpdated = file.UpdatedAt
		}
	}
	fileCount := len(props.Files)
	lastActivity := "No activity yet"
	if !lastUpdated.IsZero() {
		lastActivity = formatTime(lastUpdated)
	}

	return web.Page{
		Title:      "Arkive · Files",
		CSS:        []string{"/web/pages/files.css"},
		JS:         []string{"/static/files.js", "/static/monetag-onclick.js", "/static/monetag-vignette.js"},
		AuthLayout: true,
		User:       props.Ctx.User,

		Body: g.Group([]g.Node{
			components.InlineStyle(components.InputCSS),
			h.Main(
				h.Class("files-page"),
				h.Div(
					h.Class("container"),
					h.Div(
						h.Class("page-header"),
						h.Div(
							h.Class("page-title"),
							h.H1(g.Text("Files")),
							h.P(g.Text("Browse completed uploads and manage your stored files.")),
						),
						h.Div(
							h.Class("page-actions"),
							components.Button(components.ButtonProps{
								Text:    "New upload",
								Href:    "/dashboard",
								Variant: "primary",
							}),
							components.Button(components.ButtonProps{
								Text:    "Dashboard",
								Href:    "/dashboard",
								Variant: "secondary",
							}),
						),
					),
					h.Section(
						h.Class("files-summary"),
						components.Card(components.CardProps{
							Title:    "Stored files",
							Subtitle: "Completed uploads",
							Class:    "summary-card",
							Body: []g.Node{
								h.Span(h.Class("summary-value"), g.Text(fmt.Sprintf("%d", fileCount))),
								h.Span(h.Class("summary-meta"), g.Text(fmt.Sprintf("%s total", format.Bytes(totalSize)))),
							},
						}),
						components.Card(components.CardProps{
							Title:    "Storage used",
							Subtitle: "Across your workspace",
							Class:    "summary-card",
							Body: []g.Node{
								h.Span(h.Class("summary-value"), g.Text(format.Bytes(totalSize))),
								h.Span(h.Class("summary-meta"), g.Text("Based on completed uploads")),
							},
						}),
						components.Card(components.CardProps{
							Title:    "Last activity",
							Subtitle: "Most recent upload",
							Class:    "summary-card",
							Body: []g.Node{
								h.Span(h.Class("summary-value is-small"), g.Text(lastActivity)),
								h.Span(h.Class("summary-meta"), g.Text("Updates as files complete")),
							},
						}),
					),
					h.Section(
						h.Class("files-panels"),
						h.Section(
							h.Class("panel files-list"),
							h.Div(
								h.Class("panel-header"),
								h.H2(g.Text("Completed files")),
								h.P(g.Text("Everything you have finished uploading.")),
							),
							renderCompletedList(props.Files),
						),
					),
				),
				components.Dialog(components.DialogProps{
					BackdropID: "file-delete-backdrop",
					TitleID:    "file-delete-title",
					Title:      "Delete file?",
					Body:       h.P(g.Attr("id", "file-delete-meta"), g.Text("This will permanently delete the file. This action cannot be undone.")),
					Actions: h.Div(
						h.Class("dialog-actions"),
						h.Button(
							h.Class("button secondary"),
							h.Type("button"),
							g.Attr("id", "file-delete-cancel"),
							g.Text("Cancel"),
						),
						h.Button(
							h.Class("button danger"),
							h.Type("button"),
							g.Attr("id", "file-delete-confirm"),
							g.Text("Delete"),
						),
					),
				}),
				components.Dialog(components.DialogProps{
					BackdropID: "file-share-backdrop",
					TitleID:    "file-share-title",
					Title:      "Share file",
					Body: h.Div(
						h.Class("share-dialog"),
						h.Button(
							h.Class("share-dialog-close"),
							h.Type("button"),
							g.Attr("id", "share-close-button"),
							g.Attr("aria-label", "Close"),
							components.Icon(components.IconProps{
								Name:       "x",
								Size:       "16",
								Decorative: true,
							}),
						),
						h.P(h.Class("share-subtitle"), g.Text("Link is generated instantly. Changes save automatically.")),
						h.Div(
							h.Class("share-link-row"),
							h.Div(
								h.Class("share-link-field"),
								h.Label(
									h.Class("form-label"),
									g.Attr("for", "share-link-input"),
									g.Text("Share link"),
								),
								h.Input(
									h.Class("form-input"),
									g.Attr("id", "share-link-input"),
									g.Attr("name", "share-link"),
									g.Attr("type", "text"),
									g.Attr("readonly", "readonly"),
									g.Attr("placeholder", "Generating link..."),
								),
							),
							h.Button(
								h.Class("button secondary"),
								h.Type("button"),
								g.Attr("id", "share-copy-button"),
								g.Text("Copy"),
							),
						),
						h.Div(
							h.Class("share-status"),
							h.Div(
								h.Class("share-status-main"),
								h.Span(h.Class("share-status-label"), g.Text("Status")),
								h.Span(h.Class("share-status-value"), g.Attr("id", "share-status"), g.Text("Preparing")),
							),
							h.Span(h.Class("share-save-state"), g.Attr("id", "share-save-state"), g.Text("")),
						),
						h.Div(
							h.Class("share-options"),
							h.Div(
								h.Class("form-field"),
								h.Label(h.Class("form-label"), g.Attr("for", "share-expiry-select"), g.Text("Expiry")),
								h.Select(
									h.Class("form-input"),
									g.Attr("id", "share-expiry-select"),
									h.Option(g.Attr("value", "never"), g.Text("Never")),
									h.Option(g.Attr("value", "1d"), g.Text("In 24 hours")),
									h.Option(g.Attr("value", "7d"), g.Text("In 7 days")),
									h.Option(g.Attr("value", "30d"), g.Text("In 30 days")),
									h.Option(g.Attr("value", "custom"), g.Text("Custom date")),
								),
								h.Div(
									h.Class("share-expiry-custom"),
									h.Label(h.Class("form-label"), g.Attr("for", "share-expiry-custom"), g.Text("Custom expiry")),
									h.Input(
										h.Class("form-input"),
										g.Attr("id", "share-expiry-custom"),
										g.Attr("name", "share-expiry-custom"),
										g.Attr("type", "datetime-local"),
									),
								),
							),
							h.Div(
								h.Class("form-field"),
								h.Label(h.Class("form-label"), g.Attr("for", "share-password-toggle"), g.Text("Password")),
								h.Div(
									h.Class("share-password-toggle"),
									h.Input(
										g.Attr("type", "checkbox"),
										g.Attr("id", "share-password-toggle"),
									),
									h.Label(g.Attr("for", "share-password-toggle"), g.Text("Require a password")),
								),
								h.Div(
									h.Class("share-password-field"),
									h.Div(
										h.Class("form-field"),
										h.Label(
											h.Class("form-label"),
											g.Attr("for", "share-password"),
											g.Text("Password"),
										),
										h.Div(
											h.Class("password-wrapper"),
											h.Input(
												h.Class("form-input password-input"),
												g.Attr("id", "share-password"),
												g.Attr("name", "share-password"),
												g.Attr("type", "password"),
												g.Attr("placeholder", "Set a password"),
											),
											h.Button(
												h.Class("password-toggle"),
												g.Attr("type", "button"),
												g.Attr("aria-label", "Show password"),
												g.Attr("aria-pressed", "false"),
												g.Attr("data-target", "share-password"),
												g.Attr("data-visible", "false"),
												h.Span(
													h.Class("icon-eye"),
													components.Icon(components.IconProps{
														Name:       "eye-open",
														Size:       "md",
														Decorative: true,
													}),
												),
												h.Span(
													h.Class("icon-eye-off"),
													components.Icon(components.IconProps{
														Name:       "eye-closed",
														Size:       "md",
														Decorative: true,
													}),
												),
											),
										),
									),
								),
							),
						),
						h.P(h.Class("form-error share-error"), g.Attr("id", "share-error"), g.Text("")),
					),
					Actions: h.Div(
						h.Class("dialog-actions share-dialog-actions"),
						h.Button(
							h.Class("button secondary"),
							h.Type("button"),
							g.Attr("id", "share-reset-button"),
							g.Text("Reset link"),
						),
						h.Button(
							h.Class("button danger"),
							h.Type("button"),
							g.Attr("id", "share-revoke-button"),
							g.Text("Revoke access"),
						),
						h.Button(
							h.Class("button danger"),
							h.Type("button"),
							g.Attr("id", "share-delete-button"),
							g.Text("Delete link"),
						),
					),
				}),
			),
		}),
	}
}

func renderCompletedList(files []models.File) g.Node {
	if len(files) == 0 {
		return h.Div(
			h.Class("files-empty"),
			h.P(g.Text("No completed uploads yet.")),
			components.Button(components.ButtonProps{
				Text:    "Upload your first file",
				Href:    "/dashboard",
				Variant: "secondary",
			}),
		)
	}

	rows := make([]g.Node, 0, len(files))
	rows = append(rows, h.Div(
		h.Class("files-row files-row-head"),
		h.Span(g.Text("File")),
		h.Span(g.Text("Size")),
		h.Span(g.Text("Updated")),
		h.Span(g.Text("Actions")),
	))
	for _, file := range files {
		previewable := isPreviewableContentType(file.ContentType)
		rows = append(rows, h.Div(
			h.Class("files-row"),
			g.Attr("data-file-row", file.ID),
			h.Div(
				h.Class("files-file"),
				h.Span(h.Class("files-badge"), g.Text(fileTypeLabel(file))),
				h.Div(
					h.Class("files-meta"),
					h.Span(h.Class("files-name"), g.Text(file.Filename)),
					h.Span(h.Class("files-sub"), g.Text(fileSubtitle(file))),
				),
			),
			h.Span(h.Class("files-size"), g.Text(format.Bytes(file.SizeBytes))),
			h.Span(h.Class("files-time"), g.Text(formatTime(file.UpdatedAt))),
			h.Div(
				h.Class("files-actions"),
				h.Div(
					h.Class("files-action-buttons"),
					h.Button(
						h.Class("button secondary"),
						h.Type("button"),
						g.Attr("data-file-action", "share"),
						g.Attr("data-file-id", file.ID),
						g.Text("Share"),
					),
					g.If(!previewable, h.Button(
						h.Class("button secondary is-disabled"),
						h.Type("button"),
						g.Attr("disabled", "disabled"),
						g.Text("View"),
					)),
					g.If(previewable, h.A(
						h.Class("button secondary"),
						h.Href(fmt.Sprintf("/files/%s/view", file.ID)),
						g.Attr("data-file-action", "view"),
						g.Attr("data-file-id", file.ID),
						g.Text("View"),
					)),
					h.Button(
						h.Class("button danger"),
						h.Type("button"),
						g.Attr("data-file-action", "delete"),
						g.Attr("data-file-id", file.ID),
						g.Attr("data-file-name", file.Filename),
						g.Text("Delete"),
					),
				),
			),
		))
	}

	return h.Div(h.Class("files-rows"), g.Group(rows))
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("Jan 2, 2006 15:04")
}

func isPreviewableContentType(contentType string) bool {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	return strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")
}

func fileTypeLabel(file models.File) string {
	ext := strings.TrimSpace(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
	if ext == "" {
		parts := strings.Split(strings.TrimSpace(file.ContentType), "/")
		if len(parts) > 1 {
			ext = parts[len(parts)-1]
		}
	}
	ext = strings.ToUpper(ext)
	if ext == "" {
		return "FILE"
	}
	if len(ext) > 5 {
		return "FILE"
	}
	return ext
}

func fileSubtitle(file models.File) string {
	contentType := strings.TrimSpace(file.ContentType)
	if contentType == "" {
		contentType = "Unknown type"
	}
	return fmt.Sprintf("%s • %s", contentType, format.Bytes(file.SizeBytes))
}
