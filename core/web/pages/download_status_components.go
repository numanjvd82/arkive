package pages

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func renderDownloadStatus(queueClass string) g.Node {
	return g.Group([]g.Node{
		h.Div(g.Attr("id", "download-warning")),
		h.Div(
			h.Class("upload-queue "+queueClass),
			g.Attr("id", "download-queue"),
			g.Attr("hidden", "hidden"),
			h.Div(
				h.Class("queue-header"),
				h.H3(h.Class("queue-title"), g.Text("Download")),
				h.Span(h.Class("queue-meta"), g.Attr("id", "download-queue-meta"), g.Text("1 item active")),
			),
			h.Div(
				h.Class("queue-list"),
				h.Div(
					h.Class("queue-item"),
					g.Attr("id", "download-queue-item"),
					h.Div(
						h.Class("queue-item-top"),
						h.Div(
							h.Class("queue-item-file"),
							h.Span(
								h.Class("queue-item-icon"),
								lucide.Download(
									g.Attr("aria-hidden", "true"),
								),
							),
							h.Span(h.Class("queue-item-name"), g.Attr("id", "download-file-label"), g.Text("Encrypted file")),
							h.Span(h.Class("queue-item-badge is-queued"), g.Attr("id", "download-status-badge"), g.Text("queued")),
						),
						h.Div(
							h.Class("queue-item-actions"),
							h.Button(
								h.Class("queue-item-action is-cancel"),
								h.Type("button"),
								g.Attr("id", "download-cancel"),
								g.Attr("aria-label", "Cancel download"),
								g.Attr("title", "Cancel download"),
								lucide.Trash2(
									g.Attr("aria-hidden", "true"),
								),
							),
						),
					),
					h.Div(
						h.Class("queue-item-track"),
						h.Span(h.Class("queue-item-fill"), g.Attr("id", "download-progress-fill"), g.Attr("style", "width: 0%")),
					),
					h.Div(
						h.Class("queue-item-meta"),
						h.Span(h.Class("queue-item-progress mono"), g.Attr("id", "download-progress-text"), g.Text("0 B / 0 B")),
						h.Span(h.Class("queue-item-speed mono"), g.Attr("id", "download-status-detail"), g.Text("queued")),
					),
				),
			),
		),
	})
}
