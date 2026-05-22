package pages

import (
	"encoding/hex"
	"path"
	"strings"
	"time"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web/components"
	"arkive/pkg/format"
)

func buildPublicShareJS() []string {
	return []string{
		"/static/vendor/plyr/plyr.polyfilled.js",
	}
}

func buildPublicShareModuleJS() []string {
	return []string{
		"/static/plyr.js",
		"/static/public_share.js",
	}
}

func renderPublicShareHeader() g.Node {
	return h.Div(
		h.Class("public-share-header"),
		lucide.Lock(
			h.Class("public-share-header-icon"),
			g.Attr("aria-hidden", "true"),
		),
		h.Span(h.Class("public-share-header-text"), g.Text("Arkive Core Secure Gateway")),
	)
}

func renderPublicShareCard(props PublicShareViewProps) g.Node {
	return h.Section(
		h.Class("public-share-card"),
		g.Attr("data-public-share-token", props.Token),
		renderPublicSharePreview(props),
		h.Div(
			h.Class("public-share-meta"),
			h.Div(
				h.Class("public-share-copy"),
				h.Div(
					h.Class("public-share-name-row"),
					publicShareTypeIcon(props),
					h.H1(
						h.Class("public-share-name"),
						g.Attr("data-public-share-name", "true"),
						g.Text(publicShareName(props)),
					),
				),
				h.Div(
					h.Class("public-share-facts"),
					renderPublicShareFact(
						lucide.Database(
							h.Class("public-share-fact-icon"),
							g.Attr("aria-hidden", "true"),
						),
						publicShareSizeText(props),
						"public-share-fact-label",
						"data-public-share-size",
					),
					renderPublicShareSeparator(),
					renderPublicShareFact(
						lucide.Clock3(
							h.Class("public-share-fact-icon"),
							g.Attr("aria-hidden", "true"),
						),
						publicShareSharedText(props.SharedAt),
					),
					renderPublicShareSeparator(),
					renderPublicShareFact(
						lucide.ShieldCheck(
							h.Class("public-share-fact-icon is-verified"),
							g.Attr("aria-hidden", "true"),
						),
						"Zero-Knowledge Encrypted",
						"public-share-fact-label is-verified",
					),
				),
			),
			h.Div(
				h.Class("public-share-actions"),
				h.A(
					h.Class("public-share-download"),
					g.Attr("id", "public-share-download"),
					h.Href(props.DownloadURL),
					lucide.Download(
						h.Class("public-share-download-icon"),
						g.Attr("aria-hidden", "true"),
					),
					h.Span(g.Text("Download File")),
				),
				renderDownloadStatus("public-share-download-queue"),
			),
		),
	)
}

func renderPublicSharePreview(props PublicShareViewProps) g.Node {
	return h.Div(
		h.Class("public-share-preview"),
		g.Attr("id", "public-share-preview"),
		renderPublicSharePreviewMedia(props),
		h.Div(
			h.Class("public-share-badge"),
			h.Span(g.Text(publicShareExtensionLabel(props))),
		),
	)
}

func renderPublicSharePreviewMedia(props PublicShareViewProps) g.Node {
	switch {
	case props.Viewable && props.ViewURL != "" && props.IsImage:
		return h.Div(
			h.Class("public-share-image-wrap"),
			h.Img(
				h.Class("public-share-image"),
				h.Src(props.ViewURL),
				h.Alt(publicShareName(props)),
				g.Attr("data-lightbox-trigger", "true"),
				g.Attr("data-lightbox-src", props.ViewURL),
				g.Attr("data-lightbox-title", publicShareName(props)),
				g.Attr("loading", "lazy"),
			),
			h.Button(
				h.Class("media-fullscreen-button"),
				g.Attr("type", "button"),
				g.Attr("aria-label", "Open full screen"),
				g.Attr("data-lightbox-src", props.ViewURL),
				g.Attr("data-lightbox-title", publicShareName(props)),
				components.Icon(components.IconProps{Name: "fullscreen", Size: "18", Decorative: true}),
			),
		)
	case props.Viewable && props.ViewURL != "" && props.IsVideo:
		return h.Video(
			h.Class("public-share-video plyr"),
			h.Controls(),
			g.Attr("playsinline", "playsinline"),
			g.Attr("data-video-element", "true"),
			g.Attr("src", props.ViewURL),
		)
	case strings.TrimSpace(props.PreviewText) != "":
		return h.Pre(
			h.Class("public-share-text"),
			g.Text(props.PreviewText),
		)
	default:
		return h.Div(
			h.Class("public-share-empty"),
			lucide.EyeOff(
				h.Class("public-share-empty-icon"),
				g.Attr("aria-hidden", "true"),
			),
			h.Span(g.Text("Preview unavailable")),
			h.P(g.Text("Download the file to view it locally.")),
		)
	}
}

func renderPublicShareChecksum(props PublicShareViewProps) g.Node {
	return h.Section(
		h.Class("public-share-checksum"),
		h.Span(h.Class("public-share-detail-label"), g.Text("Checksum")),
		h.Code(h.Class("public-share-detail-value"), g.Text(publicShareChecksum(props))),
	)
}

func renderPublicShareFooter() g.Node {
	return h.Footer(
		h.Class("public-share-footer"),
		h.P(
			h.Class("public-share-footer-copy"),
			g.Text("Secure, self-hosted file sharing. Powered by "),
			h.Span(h.Class("public-share-footer-strong"), g.Text("Arkive Core.")),
		),
		h.Div(
			h.Class("public-share-footer-icons"),
			lucide.Shield(
				h.Class("public-share-footer-icon"),
				g.Attr("aria-hidden", "true"),
			),
			lucide.Terminal(
				h.Class("public-share-footer-icon"),
				g.Attr("aria-hidden", "true"),
			),
			lucide.KeyRound(
				h.Class("public-share-footer-icon"),
				g.Attr("aria-hidden", "true"),
			),
		),
	)
}

func renderPublicShareFact(icon g.Node, text string, classNames ...string) g.Node {
	labelClass := "public-share-fact-label"
	attrName := ""
	if len(classNames) > 0 && strings.TrimSpace(classNames[0]) != "" {
		labelClass = strings.TrimSpace(classNames[0])
	}
	if len(classNames) > 1 && strings.TrimSpace(classNames[1]) != "" {
		attrName = strings.TrimSpace(classNames[1])
	}
	return h.Div(
		h.Class("public-share-fact"),
		icon,
		h.Span(
			h.Class(labelClass),
			g.If(attrName != "", g.Attr(attrName, "true")),
			g.Text(text),
		),
	)
}

func renderPublicShareSeparator() g.Node {
	return h.Span(h.Class("public-share-separator"), g.Attr("aria-hidden", "true"))
}

func publicShareTypeIcon(props PublicShareViewProps) g.Node {
	switch {
	case props.IsImage:
		return lucide.Image(
			h.Class("public-share-name-icon"),
			g.Attr("aria-hidden", "true"),
		)
	case props.IsVideo:
		return lucide.Film(
			h.Class("public-share-name-icon"),
			g.Attr("aria-hidden", "true"),
		)
	case strings.TrimSpace(props.PreviewText) != "":
		return lucide.FileText(
			h.Class("public-share-name-icon"),
			g.Attr("aria-hidden", "true"),
		)
	default:
		return lucide.File(
			h.Class("public-share-name-icon"),
			g.Attr("aria-hidden", "true"),
		)
	}
}

func publicShareName(props PublicShareViewProps) string {
	name := strings.TrimSpace(props.FileName)
	if name != "" {
		return name
	}
	return "Encrypted file"
}

func publicShareExtensionLabel(props PublicShareViewProps) string {
	name := strings.TrimSpace(props.FileName)
	if name != "" {
		ext := strings.TrimPrefix(strings.ToUpper(path.Ext(name)), ".")
		if ext != "" {
			return ext
		}
	}

	mime := strings.TrimSpace(strings.ToLower(props.MimeType))
	if mime != "" && strings.Contains(mime, "/") {
		parts := strings.SplitN(mime, "/", 2)
		suffix := strings.ToUpper(strings.TrimSpace(parts[1]))
		if suffix != "" {
			return suffix
		}
	}

	switch {
	case props.IsImage:
		return "IMAGE"
	case props.IsVideo:
		return "VIDEO"
	case strings.TrimSpace(props.PreviewText) != "":
		return "TEXT"
	default:
		return "FILE"
	}
}

func publicShareSizeText(props PublicShareViewProps) string {
	if props.PlaintextSize > 0 {
		return format.Bytes(props.PlaintextSize)
	}
	return "Encrypted"
}

func publicShareSharedText(sharedAt time.Time) string {
	if sharedAt.IsZero() {
		return "Shared recently"
	}
	return "Shared " + format.RelativeTime(sharedAt)
}

func publicShareChecksum(props PublicShareViewProps) string {
	if len(props.File.EncryptedHash) == 0 {
		return "BLAKE3 unavailable"
	}
	return "BLAKE3: " + hex.EncodeToString(props.File.EncryptedHash)
}
