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
	Ctx                PageContext
	Files              []models.File
	MultipartThreshold int64
}

func FilesPage(props FilesPageProps) web.Page {
	return web.Page{
		Title: "Arkive · Files",
		CSS:   []string{"/web/pages/files.css"},
		JS:    []string{"/static/uploads.js"},
		Body: h.Main(
			h.Class("files-page"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("page-header"),
					h.Div(
						h.Class("page-title"),
						h.H1(g.Text("Files")),
						h.P(g.Text("Resume multipart uploads and manage in-progress transfers.")),
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
						h.Class("panel files-resume"),
						h.Div(
							h.Class("panel-header"),
							h.H2(g.Text("Resume an upload")),
							h.P(g.Text("Select the same file to reconnect and keep uploading.")),
						),
						components.UploadControls(components.UploadControlsProps{
							InputLabel:    "Choose the same file",
							InputHelper:   "Up to 1GB. Files over 200MB use multipart chunks.",
							InputRequired: true,
						}),
					),
					h.Section(
						h.Class("panel files-list"),
						h.Div(
							h.Class("panel-header"),
							h.H2(g.Text("Pending uploads")),
							h.P(g.Text("Only multipart uploads in progress are listed here.")),
						),
						renderPendingList(props.Files, props.MultipartThreshold),
					),
				),
			),
		),
	}
}

func renderPendingList(files []models.File, multipartThreshold int64) g.Node {
	if len(files) == 0 {
		return h.P(h.Class("files-empty"), g.Text("No pending uploads right now."))
	}

	rows := make([]g.Node, 0, len(files))
	for _, file := range files {
		canResume := file.SizeBytes > multipartThreshold
		if !canResume {
			continue
		}
		rows = append(rows, h.Div(
			h.Class("files-row"),
			g.Attr("data-file-id", file.ID),
			h.Div(
				h.Class("files-meta"),
				h.Span(h.Class("files-name"), g.Text(file.Filename)),
				h.Span(h.Class("files-sub"), g.Text(fmt.Sprintf("%s • %s", format.Bytes(file.SizeBytes), titleCase(file.Status)))),
			),
			h.Div(
				h.Class("files-actions"),
				h.Span(h.Class("files-time"), g.Text(formatTime(file.UpdatedAt))),
				h.Button(
					h.Class("icon-button"),
					g.Attr("type", "button"),
					g.Attr("data-resume-id", file.ID),
					components.Icon(components.IconProps{
						Name:       "play",
						Size:       "md",
						Title:      "Resume upload",
						AriaLabel:  "Resume upload",
						Decorative: false,
					}),
				),
				h.Button(
					h.Class("icon-button danger file-abort"),
					g.Attr("type", "button"),
					g.Attr("data-file-id", file.ID),
					components.Icon(components.IconProps{
						Name:       "x",
						Size:       "md",
						Title:      "Abort upload",
						AriaLabel:  "Abort upload",
						Decorative: false,
					}),
				),
			),
		))
	}

	if len(rows) == 0 {
		return h.P(h.Class("files-empty"), g.Text("No multipart uploads to resume right now."))
	}
	return h.Div(h.Class("files-rows"), g.Group(rows))
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("Jan 2, 2006 15:04")
}

func titleCase(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	runes := []rune(value)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}
