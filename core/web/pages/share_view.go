package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
)

type PublicShareViewProps struct {
	File        models.File
	ViewURL     string
	DownloadURL string
	IsImage     bool
	IsVideo     bool
	Viewable    bool
	ShareURL    string
}

func PublicShareViewPage(props PublicShareViewProps) web.Page {
	file := props.File
	headerChips := []g.Node{
		h.Span(h.Class("chip"), g.Text("Encrypted")),
	}
	if !file.UpdatedAt.IsZero() {
		headerChips = append(headerChips, h.Span(h.Class("chip chip-muted"), g.Text(formatTime(file.UpdatedAt))))
	}

	isImage := props.IsImage
	isVideo := props.IsVideo
	largeVideo := false

	return web.Page{
		Title:   "Arkive · Shared file",
		Robots:  RobotsNoIndex,
		CSS:     append(buildMediaCSS(MediaViewPageProps{}), "/web/pages/share.css"),
		JS:      buildMediaJS(MediaViewPageProps{}),
		HideNav: true,
		Body: g.Group([]g.Node{
			components.InlineStyle(components.InputCSS),
			h.Main(
				h.Class("media-view"),
				h.Div(
					h.Class("container"),
					h.Div(
						h.Class("share-topbar"),
						components.BrandLogo(components.BrandLogoProps{
							Href:  "/",
							Class: "share-brand",
						}),
						h.Div(
							h.Class("share-badge"),
							h.Span(g.Text("Shared via Arkive")),
						),
					),
					h.Section(
						h.Class("media-header"),
						h.Div(
							h.Class("media-title"),
							h.P(h.Class("media-eyebrow"), g.Text("Media preview")),
							h.H1(g.Text("Encrypted file")),
							h.Div(h.Class("media-chips"), g.Group(headerChips)),
						),
						h.Div(
							h.Class("media-actions"),
							renderShareActions(props, largeVideo),
						),
					),
					h.Section(
						h.Class("media-shell"),
						h.Div(
							h.Class("media-main"),
							g.If(isVideo && largeVideo, h.Div(
								h.Class("media-alert is-warning"),
								h.Span(g.Text("This video is very large. Streaming may be slow. Download recommended.")),
							)),
							h.Div(
								h.Class("media-frame"),
								h.Div(
									h.Class(mediaFrameClass(MediaViewPageProps{})),
									renderShareMedia(props),
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
							metaRow("Filename", "Encrypted", ""),
							metaRow("Type", "Encrypted", ""),
							metaRow("Size", "Encrypted", ""),
							metaRow("Resolution", "Encrypted", ""),
							metaRow("Duration", "Encrypted", ""),
									metaRow("Updated", fallbackText(formatTime(file.UpdatedAt), "Not available"), ""),
								),
							),
							h.Div(
								h.Class("media-panel"),
								h.H3(g.Text("Share link")),
								h.Div(
									h.Class("share-link-panel"),
									h.Input(
										h.Class("form-input"),
										g.Attr("id", "public-share-link"),
										g.Attr("type", "text"),
										g.Attr("readonly", "readonly"),
										g.Attr("value", props.ShareURL),
									),
									components.CopyButton(components.CopyButtonProps{
										Text:     "Copy link",
										TargetID: "public-share-link",
										Variant:  "secondary",
									}),
								),
								h.P(
									h.Class("share-link-hint"),
									g.Text("Share this link with anyone you want to access the file."),
								),
							),
						),
					),
				),
				g.If(isImage, components.Lightbox()),
			),
		}),
	}
}

func renderShareActions(props PublicShareViewProps, largeVideo bool) g.Node {
	downloadClass := "button secondary"
	if props.IsVideo && largeVideo {
		downloadClass = "button primary"
	}
	nodes := []g.Node{
		h.A(
			h.Class(downloadClass),
			h.Href(props.DownloadURL),
			g.Text("Download"),
		),
	}
	if props.IsVideo && largeVideo && props.ViewURL != "" {
		nodes = append(nodes, h.Button(
			h.Class("button secondary"),
			g.Attr("type", "button"),
			g.Attr("data-video-action", "play"),
			g.Text("Play (may buffer)"),
		))
	}
	return g.Group(nodes)
}

func renderShareMedia(props PublicShareViewProps) g.Node {
	if props.Viewable && props.ViewURL != "" {
		switch {
		case props.IsVideo:
			nodes := []g.Node{
				h.Class("media-video plyr"),
				h.Controls(),
				g.Attr("playsinline", "playsinline"),
				g.Attr("data-video-element", "true"),
			}
			nodes = append(nodes, g.Attr("src", props.ViewURL))
			return h.Video(nodes...)
		case props.IsImage:
			return h.Div(
				h.Class("media-image-wrap"),
				h.Img(
					h.Class("media-image"),
					h.Src(props.ViewURL),
								h.Alt("Encrypted file"),
					g.Attr("data-lightbox-trigger", "true"),
					g.Attr("data-lightbox-src", props.ViewURL),
								g.Attr("data-lightbox-title", "Encrypted file"),
					g.Attr("loading", "lazy"),
				),
				h.Button(
					h.Class("media-fullscreen-button"),
					g.Attr("type", "button"),
					g.Attr("aria-label", "Open full screen"),
					g.Attr("data-lightbox-src", props.ViewURL),
								g.Attr("data-lightbox-title", "Encrypted file"),
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
