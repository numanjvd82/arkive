package pages

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web/components"
)

func renderShareDialog() g.Node {
	return components.Dialog(components.DialogProps{
		BackdropID:  "file-share-backdrop",
		TitleID:     "file-share-title",
		DialogClass: "share-modal",
		Header: h.Div(
			h.Class("dialog-header share-modal-header"),
			h.H2(
				h.Class("share-modal-title"),
				g.Attr("id", "file-share-title"),
				lucide.Share2(
					h.Class("share-modal-lucide share-modal-lucide-accent"),
					g.Attr("aria-hidden", "true"),
				),
				h.Span(g.Text("Share Link")),
			),
			h.Button(
				h.Class("share-dialog-close"),
				h.Type("button"),
				g.Attr("id", "share-close-button"),
				g.Attr("aria-label", "Close"),
				lucide.X(
					h.Class("share-modal-lucide"),
					g.Attr("aria-hidden", "true"),
				),
			),
		),
		Body: h.Div(
			h.Class("share-dialog"),
			h.Div(
				h.Class("share-hero"),
				h.Div(
					h.Class("share-hero-copy"),
					h.P(
						h.Class("share-eyebrow"),
						g.Text("Public share link"),
					),
					h.Div(
						h.Class("share-file-pill"),
						lucide.File(
							h.Class("share-modal-lucide share-file-pill-icon"),
							g.Attr("aria-hidden", "true"),
						),
						h.Span(g.Attr("id", "share-file-name"), g.Text("Encrypted file")),
					),
				),
			),
			h.Div(
				h.Class("share-section"),
				h.Label(
					h.Class("form-label"),
					g.Attr("for", "share-link-input"),
					g.Text("Full secure link"),
				),
				h.Div(
					h.Class("share-link-row"),
					h.Input(
						h.Class("form-input share-link-input"),
						g.Attr("id", "share-link-input"),
						g.Attr("name", "share-link"),
						g.Attr("type", "text"),
						g.Attr("readonly", "readonly"),
						g.Attr("placeholder", "Create the share to generate its full link"),
					),
					h.Button(
						h.Class("share-copy-button"),
						h.Type("button"),
						g.Attr("id", "share-copy-button"),
						g.Attr("aria-label", "Copy share link"),
						lucide.Copy(
							h.Class("share-modal-lucide"),
							g.Attr("aria-hidden", "true"),
						),
					),
				),
				h.P(
					h.Class("share-inline-note"),
					h.Span(h.Class("share-inline-note-strong"), g.Text("Important: ")),
					g.Text("the "),
					h.Code(g.Text("#")),
					g.Text(" fragment at the end contains the browser secret. If it is missing, the link will open but the file cannot be decrypted."),
				),
			),
			h.Div(
				h.Class("share-section"),
				h.Div(
					h.Class("share-section-heading share-section-heading-row"),
					h.Span(g.Text("Access")),
					components.Tooltip(components.TooltipProps{
						Class:   "share-heading-tooltip",
						Tooltip: "Password protection adds a second gate. The full link still includes the browser secret fragment for zero-knowledge decryption.",
					}),
				),
				h.Div(
					h.Class("share-setting share-setting-primary"),
					h.Div(
						h.Class("share-setting-main"),
						h.Div(
							h.Class("share-setting-title-row"),
							h.Label(
								h.Class("share-setting-label"),
								g.Attr("for", "share-password-toggle"),
								lucide.KeyRound(
									h.Class("share-modal-lucide share-setting-icon"),
									g.Attr("aria-hidden", "true"),
								),
								h.Span(g.Text("Require a password")),
							),
							components.Tooltip(components.TooltipProps{
								Class:   "share-setting-tooltip",
								Tooltip: "Use this when you do not want the fragment link alone to be enough. Recipients will need both the full link and the password.",
							}),
						),
						h.P(
							h.Class("share-setting-copy"),
							g.Text("Turn this on if the link should not be usable by anyone who receives it."),
						),
					),
					h.Label(
						h.Class("switch"),
						h.Input(
							h.Type("checkbox"),
							g.Attr("id", "share-password-toggle"),
						),
						h.Span(h.Class("switch-track"), h.Span(h.Class("switch-thumb"))),
					),
					h.Div(
						h.Class("share-password-field"),
						h.Input(
							h.Class("form-input"),
							g.Attr("id", "share-password"),
							g.Attr("name", "share-password"),
							g.Attr("type", "password"),
							g.Attr("placeholder", "Create a strong password"),
							g.Attr("autocomplete", "new-password"),
						),
						h.P(
							h.Class("share-password-helper"),
							g.Attr("id", "share-password-helper"),
							g.Text("Use at least 8 characters with lowercase, uppercase, and a symbol."),
						),
						h.P(
							h.Class("share-password-strength"),
							g.Attr("id", "share-password-strength"),
							g.Text(""),
						),
					),
				),
			),
			components.Accordion(components.AccordionProps{
				ID:    "share-advanced-settings",
				Class: "share-advanced",
				Summary: h.Div(
					h.Class("share-advanced-summary"),
					h.Div(
						h.Class("share-advanced-copy"),
						h.Span(h.Class("share-advanced-title"), g.Text("Advanced settings")),
						h.Span(h.Class("share-advanced-subtitle"), g.Text("Expiry and future delivery controls.")),
					),
				),
				Body: h.Div(
					h.Class("share-advanced-body"),
					h.Div(
						h.Class("share-setting share-setting-stack"),
						h.Div(
							h.Class("share-setting-main"),
							h.Div(
								h.Class("share-setting-title-row"),
								h.Label(
									h.Class("share-setting-label"),
									lucide.Link2(
										h.Class("share-modal-lucide share-setting-icon"),
										g.Attr("aria-hidden", "true"),
									),
									h.Span(g.Text("Link type")),
								),
							),
							h.P(
								h.Class("share-setting-copy"),
								g.Text("Choose whether this link stays stable or is consumed after first access."),
							),
						),
						h.Div(
							h.Class("share-mode-selector"),
							shareModeOption("stable", "Stable link", "Normal public link. It stays usable until revoked or expired.", true),
							shareModeOption("once", "One-time link", "Expires after the first attempt to view or download.", false),
						),
					),
					h.Div(
						h.Class("share-setting"),
						h.Div(
							h.Class("share-setting-main"),
							h.Div(
								h.Class("share-setting-title-row"),
								h.Label(
									h.Class("share-setting-label"),
									g.Attr("for", "share-expiry-toggle"),
									lucide.CalendarDays(
										h.Class("share-modal-lucide share-setting-icon"),
										g.Attr("aria-hidden", "true"),
									),
									h.Span(g.Text("Set an expiry")),
								),
								components.Tooltip(components.TooltipProps{
									Class:   "share-setting-tooltip",
									Tooltip: "Expired links stop resolving. Use this for temporary deliveries or review windows.",
								}),
							),
							h.P(
								h.Class("share-setting-copy"),
								g.Text("Optional. Leave it off if the share should stay stable until you revoke it."),
							),
						),
						h.Label(
							h.Class("switch"),
							h.Input(
								h.Type("checkbox"),
								g.Attr("id", "share-expiry-toggle"),
							),
							h.Span(h.Class("switch-track"), h.Span(h.Class("switch-thumb"))),
						),
						h.Div(
							h.Class("share-expiry-fields"),
							h.Select(
								h.Class("form-input share-expiry-select"),
								g.Attr("id", "share-expiry-select"),
								h.Option(g.Attr("value", "custom"), g.Text("Custom date")),
								h.Option(g.Attr("value", "1d"), g.Text("In 24 hours")),
								h.Option(g.Attr("value", "7d"), g.Text("In 7 days")),
								h.Option(g.Attr("value", "30d"), g.Text("In 30 days")),
							),
							h.Div(
								h.Class("share-expiry-custom"),
								h.Input(
									h.Class("form-input"),
									g.Attr("id", "share-expiry-custom"),
									g.Attr("name", "share-expiry-custom"),
									g.Attr("type", "date"),
								),
								h.Input(
									h.Class("form-input"),
									g.Attr("id", "share-expiry-time"),
									g.Attr("name", "share-expiry-time"),
									g.Attr("type", "time"),
								),
							),
						),
					),
				),
			}),
			h.Div(
				h.Class("share-status"),
				h.Div(
					h.Class("share-status-main"),
					h.Span(h.Class("share-status-label"), g.Text("Status")),
					h.Span(h.Class("share-status-value"), g.Attr("id", "share-status"), g.Text("Ready to create")),
				),
				h.Span(h.Class("share-save-state"), g.Attr("id", "share-save-state"), g.Text("")),
			),
			h.P(h.Class("share-state-note"), g.Attr("id", "share-state-note"), g.Text("")),
			h.P(h.Class("form-error share-error"), g.Attr("id", "share-error"), g.Text("")),
		),
		ActionsClass: "share-dialog-actions",
		Actions: g.Group([]g.Node{
			h.Button(
				h.Class("button secondary"),
				h.Type("button"),
				g.Attr("id", "share-revoke-button"),
				lucide.Ban(
					h.Class("button-lucide"),
					g.Attr("aria-hidden", "true"),
				),
				g.Text("Revoke Link"),
			),
			h.Button(
				h.Class("button danger-outline"),
				h.Type("button"),
				g.Attr("id", "share-delete-button"),
				lucide.Trash2(
					h.Class("button-lucide"),
					g.Attr("aria-hidden", "true"),
				),
				g.Text("Delete Link"),
			),
			h.Button(
				h.Class("button primary"),
				h.Type("button"),
				g.Attr("id", "share-save-button"),
				lucide.Save(
					h.Class("button-lucide"),
					g.Attr("aria-hidden", "true"),
				),
				g.Text("Save Link"),
			),
		}),
	})
}

func shareModeOption(value, title, description string, checked bool) g.Node {
	id := "share-mode-" + value
	return h.Label(
		h.Class("share-mode-option"),
		h.Input(
			h.Type("radio"),
			g.Attr("id", id),
			g.Attr("name", "share-link-type"),
			g.Attr("value", value),
			g.If(checked, g.Attr("checked", "checked")),
		),
		h.Span(
			h.Class("share-mode-copy"),
			h.Span(
				h.Class("share-mode-title"),
				g.Text(title),
			),
			h.Span(
				h.Class("share-mode-description"),
				g.Text(description),
			),
		),
	)
}
