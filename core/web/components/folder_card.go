package components

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type FolderCardProps struct {
	ID                  string
	Href                string
	EncryptedNameBase64 string
	EncryptedMetaBase64 string
}

func FolderCard(props FolderCardProps) g.Node {
	return h.Article(
		h.Class("files-card files-card-folder"),
		g.Attr("data-entry-id", props.ID),
		g.Attr("data-entry-type", "folder"),
		g.Attr("data-selectable-entry", ""),
		g.Attr("data-folder-item", props.ID),
		g.Attr("data-folder-name-b64", props.EncryptedNameBase64),
		g.Attr("data-folder-meta-b64", props.EncryptedMetaBase64),
		g.Attr("data-folder-open", props.Href),
		g.Attr("tabindex", "0"),
		h.Div(
			h.Class("files-card-preview"),
			h.Span(
				h.Class("files-folder-emblem"),
				lucide.FolderClosed(
					h.Class("files-lucide files-lucide-type files-card-icon"),
					g.Attr("aria-hidden", "true"),
				),
			),
		),
		h.Div(
			h.Class("files-card-main"),
			h.Span(
				h.Class("files-name"),
				g.Attr("data-folder-field", "name"),
				g.Text("Encrypted folder"),
			),
		),
	)
}

func FolderRow(id, href, encryptedNameBase64, encryptedMetaBase64 string) g.Node {
	return h.Tr(
		h.Class("files-row files-row-folder"),
		g.Attr("data-entry-id", id),
		g.Attr("data-entry-type", "folder"),
		g.Attr("data-selectable-entry", ""),
		g.Attr("data-folder-item", id),
		g.Attr("data-folder-name-b64", encryptedNameBase64),
		g.Attr("data-folder-meta-b64", encryptedMetaBase64),
		g.Attr("data-folder-open", href),
		h.Td(
			h.Class("files-cell files-select-cell"),
			h.Input(
				h.Type("checkbox"),
				g.Attr("data-entry-checkbox", id),
				g.Attr("aria-label", "Select folder"),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-icon"),
			h.Span(
				h.Class("files-type-icon files-folder-emblem"),
				lucide.FolderClosed(
					h.Class("files-lucide files-lucide-type"),
					g.Attr("aria-hidden", "true"),
				),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-name"),
			h.Div(
				h.Class("files-meta"),
				h.Span(
					h.Class("files-name"),
					g.Attr("data-folder-field", "name"),
					g.Text("Encrypted folder"),
				),
			),
		),
		h.Td(
			h.Class("files-cell files-cell-type"),
			h.Span(h.Class("files-code"), g.Text("Directory")),
		),
		h.Td(
			h.Class("files-cell files-cell-size"),
			h.Span(h.Class("files-code"), g.Text("-")),
		),
		h.Td(
			h.Class("files-cell files-cell-modified"),
			h.Span(h.Class("files-code"), g.Text("--")),
		),
		h.Td(
			h.Class("files-cell files-cell-actions"),
			h.A(
				h.Class("files-action-button"),
				h.Href(href),
				h.Title("Open folder"),
				lucide.FolderOpen(
					h.Class("files-lucide files-lucide-action"),
					g.Attr("aria-hidden", "true"),
				),
				h.Span(h.Class("files-action-label"), g.Text("Open")),
			),
		),
	)
}
