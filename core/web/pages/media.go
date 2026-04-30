package pages

import (
	"fmt"
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
	"arkive/pkg/video"
)

type MediaViewPageProps struct {
	Ctx      PageContext
	File     models.File
	ViewURL  string
	IsImage  bool
	IsVideo  bool
	Viewable bool
	Large    bool
}

func MediaViewPage(props MediaViewPageProps) web.Page {
	file := props.File
	contentType := strings.TrimSpace(file.ContentType)
	headerChips := []g.Node{
		h.Span(h.Class("chip"), g.Text(format.Bytes(file.SizeBytes))),
	}
	if contentType != "" {
		headerChips = append(headerChips, h.Span(h.Class("chip chip-muted"), g.Text(contentType)))
	}
	if !file.UpdatedAt.IsZero() {
		headerChips = append(headerChips, h.Span(h.Class("chip chip-muted"), g.Text(formatTime(file.UpdatedAt))))
	}

	return web.Page{
		Title:      fmt.Sprintf("Arkive · %s", file.Filename),
		Robots:     RobotsNoIndex,
		CSS:        buildMediaCSS(props),
		JS:         buildMediaJS(props),
		AuthLayout: true,
		User:       props.Ctx.User,
		Body: h.Main(
			h.Class("media-view"),
			h.Div(
				h.Class("container"),
				h.Section(
					h.Class("media-header"),
					h.Div(
						h.Class("media-title"),
						h.P(h.Class("media-eyebrow"), g.Text("Media preview")),
						h.H1(g.Text(file.Filename)),
						h.Div(h.Class("media-chips"), g.Group(headerChips)),
					),
					h.Div(
						h.Class("media-actions"),
						renderMediaActions(file.ID, props),
						components.Button(components.ButtonProps{
							Text:    "Back to files",
							Href:    "/files",
							Variant: "secondary",
						}),
					),
				),
				h.Section(
					h.Class("media-shell"),
					h.Div(
						h.Class("media-main"),
						g.If(props.IsVideo && props.Large, h.Div(
							h.Class("media-alert is-warning"),
							h.Span(g.Text("This video is very large. Streaming may be slow. Download recommended.")),
						)),
						h.Div(
							h.Class("media-frame"),
							h.Div(
								h.Class(mediaFrameClass(props)),
								renderMedia(props),
							),
						),
						g.If(!props.Viewable, h.Div(
							h.Class("media-alert"),
							h.Span(g.Text("Preview available for images and videos only.")),
						)),
						g.If(props.Viewable && props.ViewURL == "", h.Div(
							h.Class("media-alert"),
							h.Span(g.Text("Preview link is unavailable. Try again later.")),
						)),
					),
					h.Aside(
						h.Class("media-sidebar"),
						h.Div(
							h.Class("media-panel"),
							h.H3(g.Text("Details")),
							h.Div(
								h.Class("media-meta"),
								metaRow("Filename", file.Filename),
								metaRow("Type", fallbackText(contentType, "Unknown")),
								metaRow("Size", format.Bytes(file.SizeBytes)),
								metaRow("Resolution", video.FormatResolution(file.VideoWidth, file.VideoHeight, props.IsVideo)),
								metaRow("Duration", video.FormatDuration(file.VideoDurationSeconds, props.IsVideo)),
								metaRow("Updated", fallbackText(formatTime(file.UpdatedAt), "Not available")),
							),
						),
					),
				),
			),
			g.If(props.IsImage, components.Lightbox()),
		),
	}
}

func renderMedia(props MediaViewPageProps) g.Node {
	if props.Viewable && props.ViewURL != "" {
		switch {
		case props.IsVideo:
			nodes := []g.Node{
				h.Class("media-video plyr"),
				h.Controls(),
				g.Attr("playsinline", "playsinline"),
				g.Attr("data-video-element", "true"),
			}
			if props.Large {
				nodes = append(nodes, h.Preload("none"), g.Attr("data-video-src", props.ViewURL))
			} else {
				nodes = append(nodes, g.Attr("src", props.ViewURL))
			}
			return h.Video(nodes...)
		case props.IsImage:
			return h.Div(
				h.Class("media-image-wrap"),
				h.Img(
					h.Class("media-image"),
					h.Src(props.ViewURL),
					h.Alt(props.File.Filename),
					g.Attr("data-lightbox-trigger", "true"),
					g.Attr("data-lightbox-src", props.ViewURL),
					g.Attr("data-lightbox-title", props.File.Filename),
					g.Attr("loading", "lazy"),
				),
				h.Button(
					h.Class("media-fullscreen-button"),
					g.Attr("type", "button"),
					g.Attr("aria-label", "Open full screen"),
					g.Attr("data-lightbox-src", props.ViewURL),
					g.Attr("data-lightbox-title", props.File.Filename),
					components.Icon(components.IconProps{Name: "fullscreen", Size: "18", Decorative: true}),
				),
			)
		}
	}

	return h.Div(
		h.Class("media-placeholder"),
		h.Span(g.Text("Preview unavailable")),
		h.P(g.Text("Download the file to view it locally.")),
	)
}

func mediaFrameClass(props MediaViewPageProps) string {
	className := "media-frame-inner"
	if props.IsImage {
		className += " is-image"
	}
	return className
}

func buildMediaCSS(props MediaViewPageProps) []string {
	css := []string{"/web/pages/media.css"}
	if props.IsVideo {
		css = append([]string{"https://cdn.plyr.io/3.7.8/plyr.css"}, css...)
	}
	return css
}

func buildMediaJS(props MediaViewPageProps) []string {
	js := []string{"/static/media.js"}
	if props.IsVideo {
		js = append([]string{
			"https://cdn.plyr.io/3.7.8/plyr.polyfilled.js",
			"/static/plyr.js",
		}, js...)
	}
	return js
}

func renderMediaActions(fileID string, props MediaViewPageProps) g.Node {
	downloadClass := "button secondary"
	if props.IsVideo && props.Large {
		downloadClass = "button primary"
	}
	nodes := []g.Node{
		h.Button(
			h.Class(downloadClass),
			g.Attr("type", "button"),
			g.Attr("data-download-id", fileID),
			g.Text("Download"),
		),
	}
	if props.IsVideo && props.Large && props.ViewURL != "" {
		nodes = append(nodes, h.Button(
			h.Class("button secondary"),
			g.Attr("type", "button"),
			g.Attr("data-video-action", "play"),
			g.Text("Play (may buffer)"),
		))
	}
	return g.Group(nodes)
}

func metaRow(label, value string) g.Node {
	return h.Div(
		h.Class("meta-row"),
		h.Span(h.Class("meta-label"), g.Text(label)),
		h.Span(h.Class("meta-value"), g.Text(value)),
	)
}

func fallbackText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
