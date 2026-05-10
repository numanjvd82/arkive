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
	ViewMode      string
	Page          int
	PageSize      int
	TotalFiles    int
}

func FilesPage(props FilesPageProps) web.Page {
	archivedCount := props.ArchivedCount

	return web.Page{
		Title:              "Arkive · Files",
		Robots:             RobotsNoIndex,
		CSS:                []string{"/web/pages/files.css"},
		JS:                 []string{"/static/files.js"},
		AuthLayout:         true,
		RequireVaultUnlock: true,
		User:               props.Ctx.User,
		ActiveNav:          "files",

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
							renderFilesViewToggle(props.Query, props.ViewMode),
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
						renderCompletedList(props),
					),
				),
				components.Dialog(components.DialogProps{
					BackdropID: "file-delete-backdrop",
					TitleID:    "file-delete-title",
					Title:      "Delete file?",
					Body:       h.P(g.Attr("id", "file-delete-meta"), g.Text("This will permanently delete the file. This action cannot be undone.")),
					Actions: g.Group([]g.Node{
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
					}),
				}),
				components.Dialog(components.DialogProps{
					BackdropID:  "file-share-backdrop",
					TitleID:     "file-share-title",
					DialogClass: "share-modal",
					Header: h.Div(
						h.Class("dialog-header share-modal-header"),
						h.H2(
							h.Class("share-modal-title"),
							g.Attr("id", "file-share-title"),
							lucide.Share2(
								h.Class("share-modal-lucide share-modal-lucide-accent"),
								g.Attr("aria-hidden", "true"),
							),
							h.Span(g.Text("Share File: ")),
							h.Span(g.Attr("id", "share-file-name"), g.Text("")),
						),
						h.Button(
							h.Class("share-dialog-close"),
							h.Type("button"),
							g.Attr("id", "share-close-button"),
							g.Attr("aria-label", "Close"),
							lucide.X(
								h.Class("share-modal-lucide"),
								g.Attr("aria-hidden", "true"),
							),
						),
					),
					Body: h.Div(
						h.Class("share-dialog"),
						h.Div(
							h.Class("share-section"),
							h.Label(
								h.Class("form-label"),
								g.Attr("for", "share-link-input"),
								g.Text("Share Link"),
							),
							h.Div(
								h.Class("share-link-row"),
								h.Input(
									h.Class("form-input share-link-input"),
									g.Attr("id", "share-link-input"),
									g.Attr("name", "share-link"),
									g.Attr("type", "text"),
									g.Attr("readonly", "readonly"),
									g.Attr("placeholder", "Link will appear after saving"),
								),
								h.Button(
									h.Class("share-copy-button"),
									h.Type("button"),
									g.Attr("id", "share-copy-button"),
									g.Attr("aria-label", "Copy share link"),
									lucide.Copy(
										h.Class("share-modal-lucide"),
										g.Attr("aria-hidden", "true"),
									),
								),
							),
						),
						h.Div(
							h.Class("share-section"),
							h.Div(
								h.Class("share-section-heading"),
								h.Span(g.Text("Access Settings")),
							),
							h.Div(
								h.Class("share-setting"),
								h.Div(
									h.Class("share-setting-main"),
									h.Label(
										h.Class("share-setting-label"),
										g.Attr("for", "share-password-toggle"),
										lucide.KeyRound(
											h.Class("share-modal-lucide share-setting-icon"),
											g.Attr("aria-hidden", "true"),
										),
										h.Span(g.Text("Password Protection")),
									),
								),
								h.Label(
									h.Class("switch"),
									h.Input(
										h.Type("checkbox"),
										g.Attr("id", "share-password-toggle"),
									),
									h.Span(h.Class("switch-track"), h.Span(h.Class("switch-thumb"))),
								),
								h.Div(
									h.Class("share-password-field"),
									h.Input(
										h.Class("form-input"),
										g.Attr("id", "share-password"),
										g.Attr("name", "share-password"),
										g.Attr("type", "password"),
										g.Attr("placeholder", "Enter password"),
									),
									h.P(
										h.Class("share-password-helper"),
										g.Attr("id", "share-password-helper"),
										g.Text("Leave blank to keep the existing password."),
									),
								),
							),
							h.Div(
								h.Class("share-setting"),
								h.Div(
									h.Class("share-setting-main"),
									h.Label(
										h.Class("share-setting-label"),
										g.Attr("for", "share-expiry-toggle"),
										lucide.CalendarDays(
											h.Class("share-modal-lucide share-setting-icon"),
											g.Attr("aria-hidden", "true"),
										),
										h.Span(g.Text("Expiry Date")),
									),
								),
								h.Label(
									h.Class("switch"),
									h.Input(
										h.Type("checkbox"),
										g.Attr("id", "share-expiry-toggle"),
									),
									h.Span(h.Class("switch-track"), h.Span(h.Class("switch-thumb"))),
								),
								h.Div(
									h.Class("share-expiry-fields"),
									h.Div(
										h.Class("share-expiry-custom"),
										h.Input(
											h.Class("form-input"),
											g.Attr("id", "share-expiry-custom"),
											g.Attr("name", "share-expiry-custom"),
											g.Attr("type", "date"),
										),
										h.Input(
											h.Class("form-input"),
											g.Attr("id", "share-expiry-time"),
											g.Attr("name", "share-expiry-time"),
											g.Attr("type", "time"),
										),
									),
									h.Select(
										h.Class("share-expiry-select"),
										g.Attr("id", "share-expiry-select"),
										h.Option(g.Attr("value", "custom"), g.Text("Custom date")),
										h.Option(g.Attr("value", "1d"), g.Text("In 24 hours")),
										h.Option(g.Attr("value", "7d"), g.Text("In 7 days")),
										h.Option(g.Attr("value", "30d"), g.Text("In 30 days")),
									),
								),
							),
							h.Div(
								h.Class("share-setting share-setting-inline"),
								h.Div(
									h.Class("share-setting-main"),
									h.Label(
										h.Class("share-setting-label"),
										g.Attr("for", "share-burn-toggle"),
										lucide.Flame(
											h.Class("share-modal-lucide share-setting-icon"),
											g.Attr("aria-hidden", "true"),
										),
										h.Span(g.Text("Burn After Reading (Single Download)")),
									),
								),
								h.Label(
									h.Class("switch"),
									h.Input(
										h.Type("checkbox"),
										g.Attr("id", "share-burn-toggle"),
									),
									h.Span(h.Class("switch-track"), h.Span(h.Class("switch-thumb"))),
								),
							),
						),
						h.Div(
							h.Class("share-status"),
							h.Div(
								h.Class("share-status-main"),
								h.Span(h.Class("share-status-label"), g.Text("Status")),
								h.Span(h.Class("share-status-value"), g.Attr("id", "share-status"), g.Text("Not shared")),
							),
							h.Span(h.Class("share-save-state"), g.Attr("id", "share-save-state"), g.Text("")),
						),
						h.P(h.Class("form-error share-error"), g.Attr("id", "share-error"), g.Text("")),
					),
					ActionsClass: "share-dialog-actions",
					Actions: g.Group([]g.Node{
						h.Button(
							h.Class("button danger-outline"),
							h.Type("button"),
							g.Attr("id", "share-delete-button"),
							lucide.Trash2(
								h.Class("button-lucide"),
								g.Attr("aria-hidden", "true"),
							),
							g.Text("Delete Link"),
						),
						h.Button(
							h.Class("button primary"),
							h.Type("button"),
							g.Attr("id", "share-save-button"),
							lucide.Save(
								h.Class("button-lucide"),
								g.Attr("aria-hidden", "true"),
							),
							g.Text("Update Share Settings"),
						),
					}),
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

func renderCompletedList(props FilesPageProps) g.Node {
	if len(props.Files) == 0 {
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

	pagination := &components.PaginationProps{
		TotalRecords: props.TotalFiles,
		CurrentPage:  props.Page,
		PageSize:     props.PageSize,
		BaseURL:      "/files",
		Query:        props.Query,
		PageSizes:    []int{25, 50, 100},
	}

	return h.Div(
		h.Class("files-browser"),
		renderFilesBrowserToolbar(pagination, true),
		g.If(props.ViewMode == "grid", renderGridList(props.Files)),
		g.If(props.ViewMode != "grid", renderTableList(props.Files)),
		renderFilesPagination(pagination),
	)
}

func renderFilesBrowserToolbar(pagination *components.PaginationProps, showActions bool) g.Node {
	return h.Div(
		h.Class("files-browser-toolbar"),
		g.If(showActions,
			h.Div(
				h.Class("files-browser-actions"),
				g.Attr("id", "files-selection-toolbar"),
				g.Attr("hidden", "hidden"),
				renderTableActions(),
			),
		),
		h.Div(
			h.Class("files-browser-pagination"),
			components.Pagination(*pagination),
		),
	)
}

func renderFilesPagination(pagination *components.PaginationProps) g.Node {
	return h.Div(
		h.Class("files-browser-toolbar files-browser-toolbar-bottom"),
		h.Div(
			h.Class("files-browser-pagination"),
			components.Pagination(*pagination),
		),
	)
}

func renderTableList(files []models.File) g.Node {
	rows := make([]g.Node, 0, len(files))
	for _, file := range files {
		rows = append(rows, renderFileRow(file))
	}

	return components.DataTable(components.DataTableProps{
		WrapClass:  "files-table-wrap",
		TableClass: "files-table",
		Header: h.THead(
			h.Tr(
				h.Th(
					h.Class("files-select-cell"),
					h.Input(
						h.Type("checkbox"),
						g.Attr("id", "files-select-all"),
						g.Attr("aria-label", "Select all files"),
					),
				),
				h.Th(g.Text("Name")),
				h.Th(g.Text("Type")),
				h.Th(h.Class("files-align-right"), g.Text("Size")),
				h.Th(h.Class("files-align-right"), g.Text("Modified")),
				h.Th(h.Class("files-align-center"), g.Text("Actions")),
			),
		),
		Body: h.TBody(g.Group(rows)),
	})
}

func renderGridList(files []models.File) g.Node {
	cards := make([]g.Node, 0, len(files))
	for _, file := range files {
		cards = append(cards, renderFileCard(file))
	}

	return g.Group([]g.Node{
		h.Div(
			h.Class("files-grid-wrap"),
			h.Div(
				h.Class("files-grid"),
				g.Group(cards),
			),
		),
		renderGridContextMenu(),
	})
}

func renderTableActions() g.Node {
	return h.Div(
		h.Class("files-table-actions"),
		h.Button(
			h.Class("button danger files-bulk-delete"),
			h.Type("button"),
			g.Attr("id", "files-delete-selected"),
			g.Attr("disabled", "disabled"),
			g.Text("Delete selected"),
		),
		h.Span(
			h.Class("files-selection-count"),
			g.Attr("id", "files-selection-count"),
			g.Text("0 selected"),
		),
	)
}

func renderFilesViewToggle(query url.Values, viewMode string) g.Node {
	listClass := "files-view-toggle-button"
	gridClass := "files-view-toggle-button"
	if viewMode == "grid" {
		gridClass += " is-active"
	} else {
		listClass += " is-active"
	}
	return h.Div(
		h.Class("files-view-toggle"),
		g.Attr("role", "group"),
		g.Attr("aria-label", "Choose file layout"),
		h.A(
			h.Class(listClass),
			h.Href(buildFilesViewURL(query, "list")),
			g.Attr("data-files-view-link", "list"),
			lucide.List(
				h.Class("files-lucide files-lucide-action"),
				g.Attr("aria-hidden", "true"),
			),
			h.Span(g.Text("List")),
		),
		h.A(
			h.Class(gridClass),
			h.Href(buildFilesViewURL(query, "grid")),
			g.Attr("data-files-view-link", "grid"),
			lucide.Grid2x2(
				h.Class("files-lucide files-lucide-action"),
				g.Attr("aria-hidden", "true"),
			),
			h.Span(g.Text("Grid")),
		),
	)
}

func buildFilesViewURL(query url.Values, viewMode string) string {
	next := cloneQuery(query)
	next.Set("view", viewMode)
	next.Del("page")
	if encoded := next.Encode(); encoded != "" {
		return "/files?" + encoded
	}
	return "/files"
}

func cloneQuery(values url.Values) url.Values {
	if values == nil {
		return url.Values{}
	}
	cloned := url.Values{}
	for key, items := range values {
		copied := make([]string, len(items))
		copy(copied, items)
		cloned[key] = copied
	}
	return cloned
}

func renderFileRow(file models.File) g.Node {
	timestamp := formatTime(file.UpdatedAt)
	relative := format.RelativeTime(file.UpdatedAt)
	viewURL := fmt.Sprintf("/files/%s/view", file.ID)

	return h.Tr(
		h.Class("files-row"),
		g.Attr("aria-busy", "true"),
		g.Attr("data-file-row", file.ID),
		g.Attr("data-file-item", file.ID),
		g.Attr("data-file-name", ""),
		g.Attr("id", "file-"+file.ID),
		h.Td(
			h.Class("files-cell files-select-cell"),
			h.Input(
				h.Type("checkbox"),
				g.Attr("data-file-select", file.ID),
				g.Attr("aria-label", "Select file"),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-name"),
			h.Span(
				h.Class("files-type-icon"),
				g.Attr("data-file-field", "icon"),
				fileTypeGlyph(file),
			),
			h.Div(
				h.Class("files-meta"),
				h.Span(
					h.Class("files-name files-skeleton files-skeleton-name"),
					g.Attr("data-file-field", "name"),
					g.Attr("aria-hidden", "true"),
				),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-type"),
			h.Span(
				h.Class("files-code files-skeleton files-skeleton-type"),
				g.Attr("data-file-field", "type"),
				g.Attr("aria-hidden", "true"),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-size"),
			h.Span(
				h.Class("files-code"),
				g.Attr("data-file-field", "size"),
				g.Text("Encrypted"),
			),
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
					g.Attr("data-file-name", ""),
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
				g.Attr("data-file-action", "view"),
				lucide.Eye(
					h.Class("files-lucide files-lucide-action"),
					g.Attr("aria-hidden", "true"),
				),
			),
			renderActionLink(
				"Download",
				"",
				"button",
				g.Group([]g.Node{
					h.Type("button"),
					g.Attr("data-file-action", "download"),
					g.Attr("data-file-id", file.ID),
				}),
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
					g.Attr("data-file-name", ""),
				}),
				lucide.Trash2(
					h.Class("files-lucide files-lucide-action"),
					g.Attr("aria-hidden", "true"),
				),
			),
		),
	)
}

func renderFileCard(file models.File) g.Node {
	viewURL := fmt.Sprintf("/files/%s/view", file.ID)

	return h.Article(
		h.Class("files-card"),
		g.Attr("aria-busy", "true"),
		g.Attr("data-file-card", file.ID),
		g.Attr("data-file-item", file.ID),
		g.Attr("data-file-name", ""),
		g.Attr("data-file-grid-select", file.ID),
		g.Attr("data-file-open", viewURL),
		g.Attr("tabindex", "0"),
		h.Div(
			h.Class("files-card-preview"),
			g.Attr("data-file-preview", "true"),
			h.Span(
				h.Class("files-type-icon files-card-icon"),
				g.Attr("data-file-field", "icon"),
				fileTypeGlyph(file),
			),
		),
		h.Div(
			h.Class("files-card-main"),
			h.Span(
				h.Class("files-name files-skeleton files-skeleton-name"),
				g.Attr("data-file-field", "name"),
				g.Attr("aria-hidden", "true"),
			),
		),
	)
}

func renderGridContextMenu() g.Node {
	return h.Div(
		h.Class("files-context-menu"),
		g.Attr("id", "files-grid-context-menu"),
		g.Attr("hidden", "hidden"),
		h.Button(
			h.Class("files-context-menu-item"),
			h.Type("button"),
			g.Attr("data-grid-menu-action", "open"),
			g.Text("Open"),
		),
		h.Button(
			h.Class("files-context-menu-item"),
			h.Type("button"),
			g.Attr("data-grid-menu-action", "rename"),
			g.Text("Rename"),
		),
		h.Button(
			h.Class("files-context-menu-item"),
			h.Type("button"),
			g.Attr("data-grid-menu-action", "move"),
			g.Text("Move"),
		),
		h.Div(h.Class("files-context-menu-divider")),
		h.Button(
			h.Class("files-context-menu-item"),
			h.Type("button"),
			g.Attr("data-grid-menu-action", "properties"),
			g.Text("Properties"),
		),
		h.Div(h.Class("files-context-menu-divider")),
		h.Button(
			h.Class("files-context-menu-item"),
			h.Type("button"),
			g.Attr("data-grid-space-action", "new-folder"),
			g.Text("New Folder"),
		),
		h.Button(
			h.Class("files-context-menu-item"),
			h.Type("button"),
			g.Attr("data-grid-space-action", "upload-here"),
			g.Text("Upload Here"),
		),
	)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("Jan 2, 2006 15:04")
}

func fileTypeIcon(file models.File) string {
	return "file"
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
			attrs,
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
