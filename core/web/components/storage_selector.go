package components

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type StorageSelectorOption struct {
	Value   string
	Label   string
	Checked bool
}

const StorageSelectorCSS = "/web/components/storage_selector.css"

func StorageSelector(name string, options ...StorageSelectorOption) g.Node {
	nodes := make([]g.Node, 0, len(options))
	for _, option := range options {
		nodes = append(nodes, storageSelectorOption(name, option))
	}

	return g.Group([]g.Node{
		InlineStyle(StorageSelectorCSS),
		h.Div(
			h.Class("storage-selector"),
			g.Group(nodes),
		),
	})
}

func storageSelectorOption(name string, option StorageSelectorOption) g.Node {
	id := name + "-" + option.Value
	return h.Label(
		h.Class("storage-selector-option"),
		h.Input(
			g.Attr("type", "radio"),
			g.Attr("id", id),
			g.Attr("name", name),
			g.Attr("value", option.Value),
			g.If(option.Checked, g.Attr("checked", "checked")),
		),
		h.Span(
			h.Class("storage-selector-copy"),
			storageSelectorIcon(option.Value),
			h.Strong(g.Text(option.Label)),
		),
	)
}

func storageSelectorIcon(value string) g.Node {
	switch value {
	case "s3":
		return lucide.Cloud(
			h.Class("storage-selector-lucide"),
			g.Attr("aria-hidden", "true"),
		)
	default:
		return lucide.HardDrive(
			h.Class("storage-selector-lucide"),
			g.Attr("aria-hidden", "true"),
		)
	}
}
