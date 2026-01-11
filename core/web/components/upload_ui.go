package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type UploadControlsProps struct {
	InputID       string
	InputName     string
	InputLabel    string
	InputHelper   string
	StatusText    string
	InputRequired bool
}

type UploadResumeBannerProps struct {
	ID string
}

const UploadUICSS = "/web/components/upload_ui.css"
const UploadUIJS = "/web/components/uploads.js"

func UploadControls(props UploadControlsProps) g.Node {
	inputID := props.InputID
	if inputID == "" {
		inputID = "upload-file"
	}
	inputName := props.InputName
	if inputName == "" {
		inputName = "file"
	}
	statusText := props.StatusText
	if statusText == "" {
		statusText = ""
	}

	return g.Group([]g.Node{
		InlineStyle(UploadUICSS),
		InlineScript(UploadUIJS),
		h.Div(
			h.Class("upload-dropzone"),
			g.Attr("id", "upload-dropzone"),
			g.Attr("role", "button"),
			g.Attr("tabindex", "0"),
			h.Div(
				h.Class("dropzone-icon"),
				h.Span(g.Text("⇪")),
			),
			h.Div(
				h.Class("dropzone-copy"),
				h.P(
					h.Class("dropzone-title"),
					g.Text("Drop files or folders here or "),
					h.Button(
						h.Class("dropzone-action"),
						h.Type("button"),
						g.Attr("id", "upload-browse-files"),
						g.Text("Browse files"),
					),
					g.Text(" or "),
					h.Button(
						h.Class("dropzone-action"),
						h.Type("button"),
						g.Attr("id", "upload-browse-folders"),
						g.Text("Browse folder"),
					),
				),
				h.P(
					h.Class("dropzone-sub"),
					g.Text("Fast, resumable uploads. Files queue automatically."),
				),
			),
		),
		h.Input(
			h.Type("file"),
			h.ID(inputID),
			h.Name(inputName),
			h.Class("upload-input is-hidden"),
			g.Attr("multiple", "multiple"),
			g.If(props.InputRequired, g.Attr("required", "required")),
			g.If(props.InputHelper != "", g.Attr("aria-describedby", inputID+"-helper")),
		),
		h.Input(
			h.Type("file"),
			h.ID("upload-folder"),
			h.Name("folder"),
			h.Class("upload-input is-hidden"),
			g.Attr("webkitdirectory", ""),
			g.Attr("directory", ""),
			g.Attr("mozdirectory", ""),
			g.Attr("multiple", "multiple"),
		),
		h.Div(
			h.Class("upload-chip is-hidden"),
			g.Attr("id", "upload-chip"),
			h.Span(h.Class("chip-name"), g.Attr("id", "upload-chip-name")),
			h.Span(h.Class("chip-size"), g.Attr("id", "upload-chip-size")),
			h.Button(
				h.Class("chip-clear"),
				h.Type("button"),
				g.Attr("id", "upload-chip-clear"),
				g.Text("Change"),
			),
		),
		h.Div(
			h.Class("upload-actions"),
			h.Div(
				h.Class("upload-controls is-hidden"),
				g.Attr("id", "upload-controls"),
				h.Button(
					h.Class("button secondary"),
					h.Type("button"),
					g.Attr("id", "upload-pause"),
					g.Text("Pause"),
				),
				h.Button(
					h.Class("button secondary"),
					h.Type("button"),
					g.Attr("id", "upload-resume"),
					g.Text("Resume"),
				),
				h.Button(
					h.Class("button danger"),
					h.Type("button"),
					g.Attr("id", "upload-abort"),
					g.Text("Cancel"),
				),
			),
		),

		h.Div(
			h.Class("upload-meta"),
			h.Span(h.Class("upload-meta-title"), g.Attr("id", "upload-meta-title")),
			h.Span(
				h.Class("upload-meta-detail"),
				g.Attr("id", "upload-meta-detail"),
				g.Text(""),
			),
			Tooltip(TooltipProps{
				ID:       "upload-meta-tooltip",
				Class:    "upload-meta-tooltip",
				IconName: "info",
				IconSize: "18",
			}),
		),
		h.P(
			h.Class("upload-status is-hidden"),
			g.Attr("id", "upload-status"),
			g.Text(statusText),
		),
		h.Div(
			h.Class("upload-queue"),
			g.Attr("id", "upload-queue"),
			h.Div(
				h.Class("queue-header"),
				h.Span(h.Class("queue-title"), g.Text("Queue")),
				h.Span(h.Class("queue-meta"), g.Attr("id", "upload-queue-meta"), g.Text("0 items")),
			),
			h.Div(
				h.Class("queue-list"),
				g.Attr("id", "upload-queue-list"),
			),
			h.P(
				h.Class("queue-empty"),
				g.Attr("id", "upload-queue-empty"),
				g.Text("Nothing queued yet."),
			),
		),
		Dialog(DialogProps{
			BackdropID: "upload-confirm-backdrop",
			TitleID:    "upload-confirm-title",
			Title:      "Start upload?",
			Body:       h.P(g.Attr("id", "upload-confirm-meta")),
			Actions: h.Div(
				h.Class("dialog-actions"),
				h.Button(
					h.Class("button secondary"),
					h.Type("button"),
					g.Attr("id", "upload-confirm-cancel"),
					g.Text("Cancel"),
				),
				h.Button(
					h.Class("button primary"),
					h.Type("button"),
					g.Attr("id", "upload-confirm-start"),
					g.Text("Start upload"),
				),
			),
		}),
	})
}

func UploadResumeBanner(props UploadResumeBannerProps) g.Node {
	id := props.ID
	if id == "" {
		id = "upload-resume-banner"
	}

	return g.Group([]g.Node{
		InlineStyle(UploadUICSS),
		h.Div(
			h.Class("upload-resume-banner is-hidden"),
			g.Attr("id", id),
			h.Div(
				h.Class("resume-content"),
				h.Span(h.Class("resume-title"), g.Text("Upload paused")),
				h.Span(h.Class("resume-meta"), g.Attr("id", "resume-banner-meta")),
				h.P(
					h.Class("resume-note"),
					g.Text("To resume, select the same file again. Your browser requires this."),
				),
			),
			h.Div(
				h.Class("resume-actions"),
				h.Button(
					h.Class("button primary"),
					h.Type("button"),
					g.Attr("id", "resume-banner-resume"),
					g.Text("Resume"),
				),
				h.Button(
					h.Class("button secondary"),
					h.Type("button"),
					g.Attr("id", "resume-banner-cancel"),
					g.Text("Cancel"),
				),
			),
		),
	})
}
