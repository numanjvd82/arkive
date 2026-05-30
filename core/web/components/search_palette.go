package components

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const SearchPaletteCSS = "/web/components/search_palette.css"

func SearchPalette(placeholder string) g.Node {
	if placeholder == "" {
		placeholder = "Search files, shares, or settings..."
	}

	return g.Group([]g.Node{
		InlineStyle(SearchPaletteCSS),
		h.Div(
			h.Class("search-panel is-hidden"),
			h.ID("search-panel"),
			g.Attr("aria-hidden", "true"),
			h.Div(
				h.Class("search-panel-shell"),
				h.Div(
					h.Class("search-panel-header"),
					lucide.Search(
						h.Class("search-panel-lucide search-panel-icon"),
						g.Attr("aria-hidden", "true"),
					),
					h.Input(
						h.Class("search-panel-input"),
						h.Type("search"),
						h.ID("search-panel-input"),
						g.Attr("placeholder", placeholder),
						g.Attr("aria-label", placeholder),
						g.Attr("autocomplete", "off"),
					),
					h.Kbd(h.Class("search-panel-kbd"), g.Text("ESC")),
				),
				h.Div(
					h.Class("search-panel-results"),
					h.Div(
						h.Class("search-section"),
						g.Attr("data-search-category", "results"),
						h.Div(h.Class("search-section-title"), g.Text("Results")),
						h.Div(h.Class("search-section-list"), g.Attr("id", "search-results-items")),
					),
					h.Div(
						h.Class("search-section"),
						g.Attr("data-search-category", "shares"),
						h.Div(h.Class("search-section-title"), g.Text("Shares")),
						h.Div(h.Class("search-section-list"), g.Attr("id", "search-results-shares")),
					),
					h.Div(
						h.Class("search-section"),
						g.Attr("data-search-category", "settings"),
						h.Div(h.Class("search-section-title"), g.Text("Settings")),
						h.Div(h.Class("search-section-list"), g.Attr("id", "search-results-settings")),
					),
					h.Div(
						h.Class("search-loading is-hidden"),
						h.ID("search-loading"),
						g.Text("Searching..."),
					),
					h.Div(
						h.Class("search-empty is-hidden"),
						h.ID("search-empty"),
						g.Text("No results found."),
					),
				),
				h.Div(
					h.Class("search-panel-footer"),
					h.Div(
						h.Class("search-footer-hint"),
						h.Kbd(h.Class("search-panel-kbd"), g.Text("↑↓")),
						h.Span(g.Text("to navigate")),
					),
					h.Div(
						h.Class("search-footer-hint"),
						h.Kbd(h.Class("search-panel-kbd"), g.Text("↵")),
						h.Span(g.Text("to select")),
					),
				),
			),
		),
	})
}
