package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const AdBlockModalCSS = "/web/components/adblock_modal.css"
const AdBlockModalJS = "/web/components/adblock_modal.js"

func AdBlockModal() g.Node {
	return g.Group([]g.Node{
		InlineStyle(AdBlockModalCSS),
		InlineScript(AdBlockModalJS),
		Dialog(DialogProps{
			BackdropID: "adblock-backdrop",
			TitleID:    "adblock-title",
			Title:      "Ad Blocker Detected - Please whitelist arkive.sh to support free uploads & downloads.",
			Body: h.Div(
				h.Class("adblock-body"),
				h.Button(
					h.Class("adblock-close"),
					h.Type("button"),
					g.Attr("aria-label", "Close"),
					g.Attr("id", "adblock-close"),
					Icon(IconProps{Name: "x", Size: "16", Decorative: true}),
				),
				h.P(h.Class("adblock-eyebrow label"), g.Text("Support arkive.sh")),
				h.P(
					h.Class("adblock-note"),
					h.Span(
						h.Class("adblock-note-icon"),
						Icon(IconProps{Name: "info", Size: "14", Decorative: true}),
					),
					g.Text("We only use non-intrusive ads. No popups, no auto audio."),
				),
				h.P(
					h.Class("adblock-warning"),
					g.Attr("id", "adblock-warning"),
					g.Attr("hidden", "hidden"),
					g.Text("Disable your ad blocker to continue."),
				),
				h.Div(
					h.Class("adblock-steps"),
					g.Attr("id", "adblock-steps"),
					g.Attr("hidden", "hidden"),
					h.Ol(
						h.Li(g.Text("Click your ad blocker icon")),
						h.Li(g.Text("Select \"Allow on this site\"")),
						h.Li(g.Text("Refresh the page")),
					),
				),
			),
			Actions: h.Div(
				h.Class("dialog-actions adblock-actions"),
				h.Button(
					h.Class("button primary"),
					g.Attr("type", "button"),
					g.Attr("id", "adblock-help"),
					g.Text("How to whitelist"),
				),
				h.Button(
					h.Class("button secondary"),
					g.Attr("type", "button"),
					g.Attr("id", "adblock-continue"),
					g.Text("Continue"),
				),
			),
		}),
	})
}
