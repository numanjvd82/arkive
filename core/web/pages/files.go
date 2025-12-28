package pages

import (
	"fmt"
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
	return web.Page{
		Title: "Arkive · Files",
		CSS:   []string{"/web/pages/files.css"},
		JS:    []string{"/static/files.js"},
		Body: h.Main(
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
							Text:    "Back to dashboard",
							Href:    "/dashboard",
							Variant: "secondary",
						}),
					),
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
		),
	}
}

func renderCompletedList(files []models.File) g.Node {
	if len(files) == 0 {
		return h.P(h.Class("files-empty"), g.Text("No completed uploads yet."))
	}

	rows := make([]g.Node, 0, len(files))
	for _, file := range files {
		previewable := isPreviewableContentType(file.ContentType)
		rows = append(rows, h.Div(
			h.Class("files-row"),
			g.Attr("data-file-row", file.ID),
			h.Div(
				h.Class("files-meta"),
				h.Span(h.Class("files-name"), g.Text(file.Filename)),
				h.Span(h.Class("files-sub"), g.Text(fmt.Sprintf("%s • Completed", format.Bytes(file.SizeBytes)))),
			),
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
				h.Span(h.Class("files-time"), g.Text(formatTime(file.UpdatedAt))),
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
