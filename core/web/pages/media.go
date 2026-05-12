package pages

import (
	"encoding/hex"
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
)

type MediaViewPageProps struct {
	Ctx  PageContext
	File models.File
}

func MediaViewPage(props MediaViewPageProps) web.Page {
	file := props.File
	integrityHash := "Unavailable"
	if len(file.EncryptedHash) > 0 {
		integrityHash = hex.EncodeToString(file.EncryptedHash)
	}

	return web.Page{
		Title:              "Arkive · Encrypted file",
		Robots:             RobotsNoIndex,
		CSS:                buildMediaCSS(props),
		JS:                 buildMediaJS(props),
		AuthLayout:         true,
		RequireVaultUnlock: true,
		User:               props.Ctx.User,
		ActiveNav:          "files",
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
							h.Div(
								h.Class("media-stage"),
								h.Div(
									h.Class("media-frame"),
									h.Div(
										h.Class(mediaFrameClass(props)),
										g.Attr("data-media-stage", "true"),
										renderMedia(props),
									),
								),
							),
						),
						h.Aside(
							h.Class("media-sidebar"),
							h.Div(
								h.Class("media-sidebar-card"),
								h.Div(
									h.Class("media-sidebar-head"),
									h.H1(g.Attr("data-media-title", "true"), g.Text("Encrypted file")),
									h.Div(
										h.Class("media-chips"),
										h.Span(h.Class("chip chip-muted"), g.Attr("data-media-chip-type", "true"), g.Text("Encrypted")),
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
								metaRow("Dimensions", "Encrypted", "media-dimensions"),
								metaRow("File Size", "Encrypted", "media-size"),
								metaRow("Uploaded", fallbackText(formatTime(file.CreatedAt), "Not available"), "media-uploaded"),
								metaRow("MIME Type", "Encrypted", "media-mime"),
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
										h.Span(h.Class("media-hash-label"), g.Attr("data-media-hash-label", "true"), g.Text("BLAKE3")),
										components.CopyButton(components.CopyButtonProps{
											Text:           "Copy",
											TargetID:       "media-hash-value",
											Variant:        "secondary",
											Icon:           "copy",
											AriaLabel:      "Copy integrity hash",
											SuccessTitle:   "Copied",
											SuccessMessage: "Integrity hash copied.",
										}),
									),
									h.Code(h.Class("media-hash-value"), g.Attr("id", "media-hash-value"), g.Attr("data-media-hash", "true"), g.Text(integrityHash)),
								),
								h.P(
									h.Class("media-panel-note"),
									g.Attr("data-media-hash-note", "true"),
									g.Text("Encrypted object integrity hash."),
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
								g.Attr("data-media-file-name", "Encrypted file"),
								components.Button(components.ButtonProps{
									Text:    "Download File",
									Variant: "primary",
									Icon:    "download",
									Class:   "media-action-primary",
									ID:      "media-download-button",
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
			),
			components.Lightbox(),
		}),
	}
}

func renderMedia(props MediaViewPageProps) g.Node {
	return h.Div(
		h.Class("media-placeholder"),
		g.Attr("data-media-placeholder", "true"),
		lucide.EyeOff(
			h.Class("media-placeholder-lucide"),
			g.Attr("aria-hidden", "true"),
		),
		h.Span(g.Text("Preparing secure preview")),
		h.P(g.Text("File metadata stays encrypted until decrypted in your browser.")),
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
	return "media-frame-inner"
}

func buildMediaCSS(props MediaViewPageProps) []string {
	return []string{
		"/static/vendor/plyr/plyr.css",
		"/web/pages/media.css",
	}
}

func buildMediaJS(props MediaViewPageProps) []string {
	return []string{
		"/static/vendor/plyr/plyr.polyfilled.js",
		"/static/plyr.js",
		"/static/media.js",
	}
}

func metaRow(label, value, id string) g.Node {
	return h.Div(
		h.Class("meta-row"),
		h.Span(h.Class("meta-label"), g.Text(label)),
		h.Span(h.Class("meta-value"), g.If(id != "", g.Attr("data-media-field", id)), g.Text(value)),
	)
}

func fallbackText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
