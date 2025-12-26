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
)

func FilesPage(files []models.File) web.Page {
	return web.Page{
		Title: "Arkive · Files",
		CSS:   []string{"/web/pages/files.css"},
		JS:    []string{"/static/uploads.js"},
		Body: h.Main(
			h.Class("files-page"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("files-header"),
					h.Div(
						h.Class("files-title"),
						h.H1(g.Text("Pending uploads")),
						h.P(g.Text("Resume from the same device and browser session.")),
					),
					components.Button(components.ButtonProps{
						Text:    "Back to dashboard",
						Href:    "/dashboard",
						Variant: "secondary",
					}),
				),
				h.Section(
					h.Class("files-instructions"),
					components.Card(components.CardProps{
						Title: "Resume instructions",
						Body: []g.Node{
							h.P(g.Text("To resume, select the same file again below. Resume only works from the same device and browser that started the upload. Do not clear the cache or localStorage.")),
							components.UploadInput(components.UploadInputProps{
								ID:       "upload-file",
								Name:     "file",
								Label:    "Choose the same file",
								Helper:   "Max 1GB. Files over 200MB use multipart chunks.",
								Required: true,
							}),
							h.Div(
								h.Class("upload-actions"),
								h.Button(
									h.Class("button primary"),
									h.Type("button"),
									g.Attr("id", "upload-start"),
									g.Text("Start upload"),
								),
								h.Button(
									h.Class("icon-button"),
									h.Type("button"),
									g.Attr("id", "upload-pause"),
									g.Attr("disabled", "disabled"),
									components.Icon(components.IconProps{
										Name:       "pause",
										Size:       "md",
										Title:      "Pause upload",
										AriaLabel:  "Pause upload",
										Decorative: false,
									}),
								),
								h.Button(
									h.Class("icon-button"),
									h.Type("button"),
									g.Attr("id", "upload-resume"),
									g.Attr("disabled", "disabled"),
									components.Icon(components.IconProps{
										Name:       "play",
										Size:       "md",
										Title:      "Resume upload",
										AriaLabel:  "Resume upload",
										Decorative: false,
									}),
								),
								h.Button(
									h.Class("icon-button"),
									h.Type("button"),
									g.Attr("id", "upload-abort"),
									g.Attr("disabled", "disabled"),
									components.Icon(components.IconProps{
										Name:       "x",
										Size:       "md",
										Title:      "Abort upload",
										AriaLabel:  "Abort upload",
										Decorative: false,
									}),
								),
							),
							components.ProgressBar(components.ProgressBarProps{
								ID:    "upload-progress",
								Value: 0,
								Label: "Progress",
							}),
							h.P(
								h.Class("upload-status"),
								g.Attr("id", "upload-status"),
								g.Text("No uploads yet."),
							),
						},
					}),
				),
				h.Section(
					h.Class("files-list"),
					components.Card(components.CardProps{
						Title:    "Uploads in progress",
						Subtitle: "Only pending and uploading items are shown.",
						Body: []g.Node{
							renderPendingList(files),
						},
					}),
				),
			),
		),
	}
}

func renderPendingList(files []models.File) g.Node {
	if len(files) == 0 {
		return h.P(h.Class("files-empty"), g.Text("No pending uploads right now."))
	}

	rows := make([]g.Node, 0, len(files))
	for _, file := range files {
		rows = append(rows, h.Div(
			h.Class("files-row"),
			h.Div(
				h.Class("files-meta"),
				h.Span(h.Class("files-name"), g.Text(file.Filename)),
				h.Span(h.Class("files-sub"), g.Text(fmt.Sprintf("%s • %s", formatBytes(file.SizeBytes), titleCase(file.Status)))),
			),
			h.Div(
				h.Class("files-actions"),
				h.Span(h.Class("files-time"), g.Text(formatTime(file.UpdatedAt))),
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

	return h.Div(h.Class("files-rows"), g.Group(rows))
}

func formatBytes(size int64) string {
	if size <= 0 {
		return "0 B"
	}
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	if size >= GB {
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	}
	if size >= MB {
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	}
	if size >= KB {
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	}
	return fmt.Sprintf("%d B", size)
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
