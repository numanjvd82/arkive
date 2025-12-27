package pages

import (
	"fmt"
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
		),
	}
}

func renderCompletedList(files []models.File) g.Node {
	if len(files) == 0 {
		return h.P(h.Class("files-empty"), g.Text("No completed uploads yet."))
	}

	rows := make([]g.Node, 0, len(files))
	for _, file := range files {
		rows = append(rows, h.Div(
			h.Class("files-row"),
			h.Div(
				h.Class("files-meta"),
				h.Span(h.Class("files-name"), g.Text(file.Filename)),
				h.Span(h.Class("files-sub"), g.Text(fmt.Sprintf("%s • Completed", format.Bytes(file.SizeBytes)))),
			),
			h.Div(
				h.Class("files-actions"),
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
