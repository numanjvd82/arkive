package pages

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"
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
	integrityHash := placeholderIntegrityHash(file)

	return web.Page{
		Title:      fmt.Sprintf("Arkive · %s", file.Filename),
		Robots:     RobotsNoIndex,
		CSS:        buildMediaCSS(props),
		JS:         buildMediaJS(props),
		AuthLayout: true,
		User:       props.Ctx.User,
		ActiveNav:  "files",
		Body: g.Group([]g.Node{
			components.InlineStyle(components.InputCSS),
			h.Main(
				h.Class("media-view"),
				h.Div(
					h.Class("container"),
					h.Section(
						h.Class("media-shell"),
						h.Div(
							h.Class("media-main"),
							g.If(props.IsVideo && props.Large, h.Div(
								h.Class("media-alert is-warning"),
								lucide.TriangleAlert(
									h.Class("media-alert-lucide"),
									g.Attr("aria-hidden", "true"),
								),
								h.Span(g.Text("This video is very large. Streaming may be slow. Download is recommended.")),
							)),
							h.Div(
								h.Class("media-stage"),
								h.Div(
									h.Class("media-frame"),
									h.Div(
										h.Class(mediaFrameClass(props)),
										renderMedia(props),
									),
								),
							),
							g.If(!props.Viewable, h.Div(
								h.Class("media-alert"),
								lucide.Info(
									h.Class("media-alert-lucide"),
									g.Attr("aria-hidden", "true"),
								),
								h.Span(g.Text("Preview is available for images and videos only.")),
							)),
							g.If(props.Viewable && props.ViewURL == "", h.Div(
								h.Class("media-alert"),
								lucide.Info(
									h.Class("media-alert-lucide"),
									g.Attr("aria-hidden", "true"),
								),
								h.Span(g.Text("Preview link is unavailable. Try again later.")),
							)),
						),
						h.Aside(
							h.Class("media-sidebar"),
							h.Div(
								h.Class("media-sidebar-card"),
								h.Div(
									h.Class("media-sidebar-head"),
									h.H1(g.Text(file.Filename)),
									h.Div(
										h.Class("media-chips"),
										h.Span(h.Class("chip chip-muted"), g.Text(mediaChipLabel(contentType))),
										h.Span(h.Class("chip"), g.Text("ZERO-KNOWLEDGE")),
									),
								),
							),
							renderMediaPanel(
								"File Specifications",
								lucide.FileText(
									h.Class("media-panel-lucide"),
									g.Attr("aria-hidden", "true"),
								),
								metaRow("Dimensions", fallbackText(video.FormatResolution(file.VideoWidth, file.VideoHeight, props.IsImage || props.IsVideo), "Not available")),
								metaRow("File Size", format.Bytes(file.SizeBytes)),
								metaRow("Uploaded", fallbackText(formatTime(file.CreatedAt), "Not available")),
								metaRow("MIME Type", fallbackText(contentType, "Unknown")),
							),
							renderMediaPanel(
								"Integrity Hash",
								lucide.ShieldCheck(
									h.Class("media-panel-lucide"),
									g.Attr("aria-hidden", "true"),
								),
								h.Div(
									h.Class("media-hash-block"),
									h.Div(
										h.Class("media-hash-top"),
										h.Span(h.Class("media-hash-label"), g.Text("SHA-256")),
										components.CopyButton(components.CopyButtonProps{
											Text:           "Copy",
											Value:          integrityHash,
											Variant:        "secondary",
											Icon:           "copy",
											AriaLabel:      "Copy integrity hash",
											SuccessTitle:   "Copied",
											SuccessMessage: "Integrity hash copied.",
										}),
									),
									h.Code(h.Class("media-hash-value"), g.Text(integrityHash)),
								),
								h.P(
									h.Class("media-panel-note"),
									g.Text("Placeholder digest for the current internal view until full integrity verification lands."),
								),
							),
							renderMediaPanel(
								"Privacy Boundary",
								lucide.Lock(
									h.Class("media-panel-lucide"),
									g.Attr("aria-hidden", "true"),
								),
								h.Ul(
									h.Class("media-privacy-list"),
									h.Li(g.Text("No location metadata is shown here.")),
									h.Li(g.Text("No storage bucket, object key, or backend path is exposed.")),
									h.Li(g.Text("Everything is decrypted on the client side.")),
									h.Li(g.Text("Only display-safe file metadata is rendered in Core.")),
								),
							),
							h.Div(
								h.Class("media-actions-panel"),
								g.Attr("data-media-file-id", file.ID),
								g.Attr("data-media-file-name", file.Filename),
								components.Button(components.ButtonProps{
									Text:    "Download File",
									Href:    fmt.Sprintf("/api/files/%s/download", file.ID),
									Variant: "primary",
									Icon:    "download",
									Class:   "media-action-primary",
								}),
								h.Div(
									h.Class("media-actions-row"),
									components.Button(components.ButtonProps{
										Text:    "Share",
										Variant: "secondary",
										Icon:    "share",
										Class:   "media-action-split",
										ID:      "media-share-button",
									}),
									components.Button(components.ButtonProps{
										Text:    "Delete",
										Variant: "danger-outline",
										Icon:    "trash",
										Class:   "media-action-split",
										ID:      "media-delete-button",
									}),
								),
							),
						),
					),
				),
				g.If(props.IsImage, components.Lightbox()),
			),
		}),
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
					lucide.Expand(
						h.Class("media-fullscreen-lucide"),
						g.Attr("aria-hidden", "true"),
					),
				),
			)
		}
	}

	return h.Div(
		h.Class("media-placeholder"),
		lucide.EyeOff(
			h.Class("media-placeholder-lucide"),
			g.Attr("aria-hidden", "true"),
		),
		h.Span(g.Text("Preview unavailable")),
		h.P(g.Text("Download the file to inspect it locally.")),
	)
}

func renderMediaPanel(title string, icon g.Node, children ...g.Node) g.Node {
	nodes := []g.Node{
		h.Div(
			h.Class("media-panel-header"),
			icon,
			h.H2(g.Text(title)),
		),
	}
	nodes = append(nodes, children...)
	return h.Section(
		h.Class("media-sidebar-card media-panel"),
		g.Group(nodes),
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
		css = append([]string{"/static/vendor/plyr/plyr.css"}, css...)
	}
	return css
}

func buildMediaJS(props MediaViewPageProps) []string {
	js := []string{"/static/media.js"}
	if props.IsVideo {
		js = append([]string{
			"/static/vendor/plyr/plyr.polyfilled.js",
			"/static/plyr.js",
		}, js...)
	}
	return js
}

func metaRow(label, value string) g.Node {
	return h.Div(
		h.Class("meta-row"),
		h.Span(h.Class("meta-label"), g.Text(label)),
		h.Span(h.Class("meta-value"), g.Text(value)),
	)
}

func mediaChipLabel(contentType string) string {
	if strings.TrimSpace(contentType) == "" {
		return "UNKNOWN"
	}
	return strings.ToUpper(contentType)
}

func placeholderIntegrityHash(file models.File) string {
	source := strings.Join([]string{
		file.ID,
		file.Filename,
		file.ContentType,
		strconv.FormatInt(file.SizeBytes, 10),
		file.CreatedAt.UTC().Format("2006-01-02T15:04:05.000000000Z07:00"),
		file.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000000000Z07:00"),
	}, "|")
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])
}

func fallbackText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
