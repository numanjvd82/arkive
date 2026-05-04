package pages

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
)

type FilesPageProps struct {
	Ctx           PageContext
	Files         []models.File
	ArchivedCount int64
	Query         url.Values
	Sort          string
	Page          int
	PageSize      int
	TotalFiles    int
}

func FilesPage(props FilesPageProps) web.Page {
	archivedCount := props.ArchivedCount

	return web.Page{
		Title:      "Arkive · Files",
		Robots:     RobotsNoIndex,
		CSS:        []string{"/web/pages/files.css"},
		JS:         []string{"/static/files.js"},
		AuthLayout: true,
		User:       props.Ctx.User,
		ActiveNav:  "files",

		Body: g.Group([]g.Node{
			components.InlineStyle(components.InputCSS),
			components.InlineStyle(components.DataTableCSS),
			h.Main(
				h.Class("files-page"),
				h.Div(
					h.Class("container"),
					h.Div(
						h.Class("page-header"),
						h.Div(
							h.Class("page-title"),
							h.H1(g.Text("Files")),
							h.P(g.Text("Manage and secure your encrypted volumes.")),
						),
						h.Div(
							h.Class("page-actions"),
							renderListControls(props),
							h.A(
								h.Class("files-upload-link"),
								h.Href("/dashboard"),
								lucide.Upload(
									h.Class("files-lucide files-lucide-action"),
									g.Attr("aria-hidden", "true"),
								),
								h.Span(g.Text("Upload")),
							),
						),
					),
					renderArchivedBanner(archivedCount),
					h.Section(
						h.Class("files-table-panel"),
						renderCompletedList(props.Files),
					),
					renderPagination(props),
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

func renderArchivedBanner(count int64) g.Node {
	if count <= 0 {
		return nil
	}
	message := "You have " + format.Commas(count) + " archived files. Log in to restore them (free restores up to 2 GB/day)."
	return h.Div(
		h.Class("files-banner"),
		h.Span(h.Class("files-banner-title"), g.Text("Archived files")),
		h.Span(h.Class("files-banner-body"), g.Text(message)),
	)
}

func renderListControls(props FilesPageProps) g.Node {
	options := []components.SortOption{
		{Label: "Updated (newest)", Value: "updated_desc"},
		{Label: "Updated (oldest)", Value: "updated_asc"},
		{Label: "Name (A-Z)", Value: "name_asc"},
		{Label: "Name (Z-A)", Value: "name_desc"},
		{Label: "Size (smallest)", Value: "size_asc"},
		{Label: "Size (largest)", Value: "size_desc"},
	}
	return h.Div(
		h.Class("files-controls"),
		components.SortSelect(components.SortSelectProps{
			Label:   "Sort",
			Value:   props.Sort,
			Options: options,
			BaseURL: "/files",
			Query:   props.Query,
		}),
	)
}

func renderCompletedList(files []models.File) g.Node {
	if len(files) == 0 {
		return h.Div(
			h.Class("data-table-wrap files-table-wrap"),
			h.Div(
				h.Class("files-empty"),
				lucide.FolderOpen(
					h.Class("files-lucide files-lucide-empty"),
					g.Attr("aria-hidden", "true"),
				),
				h.H2(g.Text("No completed uploads yet.")),
				h.P(g.Text("Upload a file from the dashboard to start building your vault.")),
				h.A(
					h.Class("files-empty-link"),
					h.Href("/dashboard"),
					g.Text("Go to upload"),
				),
			),
		)
	}

	rows := make([]g.Node, 0, len(files))
	for _, file := range files {
		rows = append(rows, renderFileRow(file))
	}

	return h.Div(
		h.Class("data-table-wrap files-table-wrap"),
		h.Table(
			h.Class("data-table files-table"),
			h.THead(
				h.Tr(
					h.Th(g.Text("Name")),
					h.Th(g.Text("Type")),
					h.Th(h.Class("files-align-right"), g.Text("Size")),
					h.Th(h.Class("files-align-right"), g.Text("Modified")),
					h.Th(h.Class("files-align-center"), g.Text("Actions")),
				),
			),
			h.TBody(g.Group(rows)),
		),
	)
}

func renderFileRow(file models.File) g.Node {
	previewable := isPreviewableContentType(file.ContentType)
	timestamp := formatTime(file.UpdatedAt)
	relative := format.RelativeTime(file.UpdatedAt)
	viewURL := ""
	if previewable {
		viewURL = fmt.Sprintf("/files/%s/view", file.ID)
	}

	return h.Tr(
		h.Class("files-row"),
		g.Attr("data-file-row", file.ID),
		h.Td(
			h.Class("files-cell files-cell-name"),
			h.Span(
				h.Class("files-type-icon"),
				fileTypeGlyph(file),
			),
			h.Div(
				h.Class("files-meta"),
				h.Span(h.Class("files-name"), g.Text(file.Filename)),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-type"),
			h.Span(
				h.Class("files-code"),
				h.Title(fileSubtitle(file)),
				g.Text(fileSubtitle(file)),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-size"),
			h.Span(h.Class("files-code"), g.Text(format.Bytes(file.SizeBytes))),
		),
		h.Td(
			h.Class("files-cell files-cell-modified"),
			h.Span(h.Class("files-code"), h.Title(timestamp), g.Text(relative)),
		),
		h.Td(
			h.Class("files-cell files-cell-actions"),
			renderActionLink(
				"Share",
				"",
				"button",
				g.Group([]g.Node{
					h.Type("button"),
					g.Attr("data-file-action", "share"),
					g.Attr("data-file-id", file.ID),
				}),
				lucide.Share2(
					h.Class("files-lucide files-lucide-action"),
					g.Attr("aria-hidden", "true"),
				),
			),
			renderActionLink(
				"View",
				viewURL,
				"a",
				nil,
				lucide.Eye(
					h.Class("files-lucide files-lucide-action"),
					g.Attr("aria-hidden", "true"),
				),
			),
			renderActionLink(
				"Download",
				fmt.Sprintf("/api/files/%s/download", file.ID),
				"a",
				nil,
				lucide.Download(
					h.Class("files-lucide files-lucide-action"),
					g.Attr("aria-hidden", "true"),
				),
			),
			renderActionLink(
				"Delete",
				"",
				"button",
				g.Group([]g.Node{
					h.Type("button"),
					g.Attr("data-file-action", "delete"),
					g.Attr("data-file-id", file.ID),
					g.Attr("data-file-name", file.Filename),
				}),
				lucide.Trash2(
					h.Class("files-lucide files-lucide-action"),
					g.Attr("aria-hidden", "true"),
				),
			),
		),
	)
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

func fileSubtitle(file models.File) string {
	contentType := strings.TrimSpace(file.ContentType)
	if contentType == "" {
		contentType = "Unknown type"
	}
	return contentType
}

func fileTypeIcon(file models.File) string {
	contentType := strings.TrimSpace(strings.ToLower(file.ContentType))
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return "file-image"
	case strings.HasPrefix(contentType, "video/"):
		return "file-video"
	case strings.HasPrefix(contentType, "audio/"):
		return "file-audio"
	case strings.Contains(contentType, "zip") || strings.Contains(contentType, "rar") || strings.Contains(contentType, "tar"):
		return "file-archive"
	case strings.Contains(contentType, "pdf") || strings.Contains(contentType, "word") || strings.Contains(contentType, "officedocument"):
		return "file-doc"
	case strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "json") || strings.Contains(contentType, "xml"):
		return "file-text"
	default:
		return "file"
	}
}

func fileTypeGlyph(file models.File) g.Node {
	switch fileTypeIcon(file) {
	case "file-image":
		return lucide.Image(h.Class("files-lucide files-lucide-type"), g.Attr("aria-hidden", "true"))
	case "file-video":
		return lucide.Film(h.Class("files-lucide files-lucide-type"), g.Attr("aria-hidden", "true"))
	case "file-audio":
		return lucide.Music4(h.Class("files-lucide files-lucide-type"), g.Attr("aria-hidden", "true"))
	case "file-archive":
		return lucide.Archive(h.Class("files-lucide files-lucide-type"), g.Attr("aria-hidden", "true"))
	case "file-doc":
		return lucide.FileText(h.Class("files-lucide files-lucide-type"), g.Attr("aria-hidden", "true"))
	case "file-text":
		return lucide.Code(h.Class("files-lucide files-lucide-type"), g.Attr("aria-hidden", "true"))
	default:
		return lucide.File(h.Class("files-lucide files-lucide-type"), g.Attr("aria-hidden", "true"))
	}
}

func renderActionLink(label, href, kind string, attrs g.Node, icon g.Node) g.Node {
	classes := "files-action-button"
	if label == "Delete" {
		classes += " is-danger"
	}
	if label == "View" && strings.TrimSpace(href) == "" {
		classes += " is-disabled"
	}

	if kind == "a" {
		if strings.TrimSpace(href) == "" {
			return h.Span(
				h.Class(classes),
				g.Attr("title", "Preview unavailable"),
				g.Attr("aria-disabled", "true"),
				icon,
				h.Span(h.Class("files-action-label"), g.Text(label)),
			)
		}
		return h.A(
			h.Class(classes),
			h.Href(href),
			g.Attr("title", label),
			icon,
			h.Span(h.Class("files-action-label"), g.Text(label)),
		)
	}

	return h.Button(
		h.Class(classes),
		attrs,
		g.Attr("title", label),
		icon,
		h.Span(h.Class("files-action-label"), g.Text(label)),
	)
}

func renderPagination(props FilesPageProps) g.Node {
	if props.TotalFiles <= 0 {
		return nil
	}
	return components.Pagination(components.PaginationProps{
		TotalRecords: props.TotalFiles,
		CurrentPage:  props.Page,
		PageSize:     props.PageSize,
		BaseURL:      "/files",
		Query:        props.Query,
		PageSizes:    []int{25, 50, 100},
	})
}
