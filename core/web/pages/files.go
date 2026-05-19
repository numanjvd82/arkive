package pages

import (
	"encoding/base64"
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
	Path          []models.Folder
	Folders       []models.Folder
	Files         []models.File
	ArchivedCount int64
	Query         url.Values
	ViewMode      string
	CurrentFolder *string
	Page          int
	PageSize      int
	TotalEntries  int
}

func FilesPage(props FilesPageProps) web.Page {
	currentFolderAttr := g.Node(nil)
	if props.CurrentFolder != nil {
		currentFolderAttr = g.Attr("data-current-folder-id", *props.CurrentFolder)
	}

	return web.Page{
		Title:              "Arkive · Files",
		Robots:             RobotsNoIndex,
		CSS:                []string{"/web/pages/files.css"},
		ModuleJS:           []string{"/static/files.js"},
		AuthLayout:         true,
		RequireVaultUnlock: true,
		User:               props.Ctx.User,
		ActiveNav:          "files",
		Body: g.Group([]g.Node{
			components.InlineStyle(components.InputCSS),
			components.InlineStyle(components.DataTableCSS),
			h.Main(
				h.Class("files-page"),
				currentFolderAttr,
				g.Attr("data-files-view-mode", props.ViewMode),
				h.Div(
					h.Class("container"),
					renderFolderLocation(props.Path, props.ViewMode),
					h.Div(
						h.Class("page-header"),
						h.Div(
							h.Class("page-title"),
							h.H1(g.Text("Files")),
							h.P(g.Text("Manage and secure your encrypted volumes.")),
						),
						h.Div(
							h.Class("page-actions"),
							renderFilesViewToggle(props.Query, props.ViewMode, entriesBaseURL(props.CurrentFolder)),
							h.Button(
								h.Class("button secondary files-header-button"),
								h.Type("button"),
								g.Attr("id", "new-folder-button"),
								lucide.FolderPlus(
									h.Class("files-lucide files-lucide-action"),
									g.Attr("aria-hidden", "true"),
								),
								h.Span(g.Text("New Folder")),
							),
							h.A(
								h.Class("files-upload-link"),
								h.Href(uploadHref(props.CurrentFolder)),
								lucide.Upload(
									h.Class("files-lucide files-lucide-action"),
									g.Attr("aria-hidden", "true"),
								),
								h.Span(g.Text("Upload")),
							),
						),
					),
					renderArchivedBanner(props.ArchivedCount),
					h.Section(
						h.Class("files-table-panel"),
						renderEntriesList(props),
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
				components.FolderDialog(),
				components.MoveDialog(),
				renderShareDialog(),
			),
		}),
	}
}

func renderFolderLocation(path []models.Folder, viewMode string) g.Node {
	nodes := []g.Node{
		h.A(
			h.Class("files-location-link"),
			h.Href(filesNavURL("/files", viewMode)),
			lucide.House(
				h.Class("files-lucide files-location-home"),
				g.Attr("aria-hidden", "true"),
			),
			h.Span(g.Text("Root")),
		),
	}

	for index, folder := range path {
		nodes = append(nodes,
			lucide.ChevronRight(
				h.Class("files-lucide files-location-separator"),
				g.Attr("aria-hidden", "true"),
			),
		)

		if index == len(path)-1 {
			nodes = append(nodes,
				h.Span(
					h.Class("files-location-current"),
					g.Attr("data-folder-breadcrumb", folder.ID),
					g.Attr("data-folder-meta-b64", base64.StdEncoding.EncodeToString(folder.EncryptedMetadata)),
					g.Attr("data-folder-name-b64", base64.StdEncoding.EncodeToString(folder.EncryptedName)),
					g.Text("Encrypted folder"),
				),
			)
			continue
		}

		nodes = append(nodes,
			h.A(
				h.Class("files-location-link"),
				h.Href(filesNavURL("/folders/"+folder.ID, viewMode)),
				g.Attr("data-folder-breadcrumb", folder.ID),
				g.Attr("data-folder-meta-b64", base64.StdEncoding.EncodeToString(folder.EncryptedMetadata)),
				g.Attr("data-folder-name-b64", base64.StdEncoding.EncodeToString(folder.EncryptedName)),
				g.Text("Encrypted folder"),
			),
		)
	}
	return h.Div(
		h.Class("files-location"),
		g.Group(nodes),
	)
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

func renderEntriesList(props FilesPageProps) g.Node {
	if len(props.Folders) == 0 && len(props.Files) == 0 {
		return h.Div(
			h.Class("data-table-wrap files-table-wrap"),
			h.Div(
				h.Class("files-empty"),
				lucide.FolderOpen(
					h.Class("files-lucide files-lucide-empty"),
					g.Attr("aria-hidden", "true"),
				),
				h.H2(g.Text("This folder is empty.")),
				h.P(g.Text("Create a folder or upload a file to start organizing your vault.")),
				h.A(
					h.Class("files-empty-link"),
					h.Href(uploadHref(props.CurrentFolder)),
					g.Text("Go to upload"),
				),
			),
		)
	}

	pagination := &components.PaginationProps{
		TotalRecords: props.TotalEntries,
		CurrentPage:  props.Page,
		PageSize:     props.PageSize,
		BaseURL:      entriesBaseURL(props.CurrentFolder),
		Query:        props.Query,
		PageSizes:    []int{25, 50, 100},
	}

	return h.Div(
		h.Class("files-browser"),
		renderEntriesBrowserToolbar(pagination),
		g.If(props.ViewMode == "grid", renderGridList(props.Folders, props.Files, props.ViewMode)),
		g.If(props.ViewMode != "grid", renderTableList(props.Folders, props.Files, props.ViewMode)),
		renderFilesPagination(pagination),
	)
}

func renderEntriesBrowserToolbar(pagination *components.PaginationProps) g.Node {
	return h.Div(
		h.Class("files-browser-toolbar"),
		h.Div(
			h.Class("files-browser-actions"),
			g.Attr("id", "entries-selection-toolbar"),
			g.Attr("hidden", "hidden"),
			h.Div(
				h.Class("files-table-actions"),
				h.Button(
					h.Class("button secondary"),
					h.Type("button"),
					g.Attr("id", "move-entries-selected"),
					g.Text("Move selected"),
				),
				h.Span(
					h.Class("files-selection-count"),
					g.Attr("id", "entries-selection-count"),
					g.Text("0 selected"),
				),
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

func renderTableList(folders []models.Folder, files []models.File, viewMode string) g.Node {
	rows := make([]g.Node, 0, len(folders)+len(files))
	for _, folder := range folders {
		rows = append(rows, components.FolderRow(
			folder.ID,
			filesNavURL("/folders/"+folder.ID, viewMode),
			base64.StdEncoding.EncodeToString(folder.EncryptedName),
			base64.StdEncoding.EncodeToString(folder.EncryptedMetadata),
		))
	}
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
						g.Attr("id", "entries-select-all"),
						g.Attr("aria-label", "Select visible entries"),
					),
				),
				h.Th(h.Class("files-icon-column")),
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

func renderGridList(folders []models.Folder, files []models.File, viewMode string) g.Node {
	cards := make([]g.Node, 0, len(folders)+len(files))
	for _, folder := range folders {
		cards = append(cards, components.FolderCard(components.FolderCardProps{
			ID:                  folder.ID,
			Href:                filesNavURL("/folders/"+folder.ID, viewMode),
			EncryptedNameBase64: base64.StdEncoding.EncodeToString(folder.EncryptedName),
			EncryptedMetaBase64: base64.StdEncoding.EncodeToString(folder.EncryptedMetadata),
		}))
	}
	for _, file := range files {
		cards = append(cards, renderFileCard(file))
	}

	return h.Div(
		h.Class("files-grid-wrap"),
		h.Div(
			h.Class("files-grid"),
			g.Group(cards),
		),
	)
}

func renderFilesViewToggle(query url.Values, viewMode, baseURL string) g.Node {
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
			h.Href(buildFilesViewURL(baseURL, query, "list")),
			g.Attr("data-files-view-link", "list"),
			g.Attr("aria-label", "List view"),
			lucide.List(
				h.Class("files-lucide files-lucide-action"),
				g.Attr("aria-hidden", "true"),
			),
		),
		h.A(
			h.Class(gridClass),
			h.Href(buildFilesViewURL(baseURL, query, "grid")),
			g.Attr("data-files-view-link", "grid"),
			g.Attr("aria-label", "Grid view"),
			lucide.Grid2x2(
				h.Class("files-lucide files-lucide-action"),
				g.Attr("aria-hidden", "true"),
			),
		),
	)
}

func buildFilesViewURL(baseURL string, query url.Values, viewMode string) string {
	next := cloneQuery(query)
	next.Set("view", viewMode)
	next.Del("page")
	if encoded := next.Encode(); encoded != "" {
		return baseURL + "?" + encoded
	}
	return baseURL
}

func entriesBaseURL(currentFolder *string) string {
	if currentFolder == nil || strings.TrimSpace(*currentFolder) == "" {
		return "/files"
	}
	return "/folders/" + strings.TrimSpace(*currentFolder)
}

func uploadHref(currentFolder *string) string {
	if currentFolder == nil || strings.TrimSpace(*currentFolder) == "" {
		return "/dashboard"
	}
	return "/dashboard?folder=" + url.QueryEscape(strings.TrimSpace(*currentFolder))
}

func filesNavURL(baseURL, viewMode string) string {
	mode := strings.TrimSpace(viewMode)
	if mode != "grid" {
		mode = "list"
	}
	return baseURL + "?view=" + url.QueryEscape(mode)
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
		g.Attr("data-entry-id", file.ID),
		g.Attr("data-entry-type", "file"),
		g.Attr("data-selectable-entry", ""),
		g.Attr("data-file-row", file.ID),
		g.Attr("data-file-open", viewURL),
		g.Attr("data-file-item", file.ID),
		g.Attr("data-file-name", ""),
		g.Attr("id", "file-"+file.ID),
		h.Td(
			h.Class("files-cell files-select-cell"),
			h.Input(
				h.Type("checkbox"),
				g.Attr("data-entry-checkbox", file.ID),
				g.Attr("data-file-select", file.ID),
				g.Attr("aria-label", "Select file"),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-icon"),
			h.Span(
				h.Class("files-type-icon"),
				g.Attr("data-file-field", "icon"),
				fileTypeGlyph(file),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-name"),
			h.Div(
				h.Class("files-meta"),
				h.Div(
					h.Class("files-name-row"),
					h.Span(
						h.Class("files-name files-skeleton files-skeleton-name"),
						g.Attr("data-file-field", "name"),
						g.Attr("aria-hidden", "true"),
					),
					lucide.Lock(
						h.Class("files-lucide files-file-lock"),
						g.Attr("aria-hidden", "true"),
					),
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
		g.Attr("data-entry-id", file.ID),
		g.Attr("data-entry-type", "file"),
		g.Attr("data-selectable-entry", ""),
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
