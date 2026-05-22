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
			h.Div(
				h.Class("share-modal-title-block"),
				h.H2(
					h.Class("share-modal-title"),
					g.Attr("id", "file-share-title"),
					lucide.Share2(
						h.Class("share-modal-lucide share-modal-lucide-accent"),
						g.Attr("aria-hidden", "true"),
					),
					h.Span(g.Text("Share Resource")),
				),
				h.P(
					h.Class("share-file-line"),
					lucide.File(
						h.Class("share-modal-lucide share-file-pill-icon"),
						g.Attr("aria-hidden", "true"),
					),
					h.Span(g.Attr("id", "share-file-name"), g.Text("Encrypted file")),
				),
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
				h.Class("share-section"),
				h.Label(
					h.Class("share-section-label"),
					g.Attr("for", "share-link-input"),
					g.Text("Public access link"),
				),
				h.Div(
					h.Class("share-link-row"),
					h.Input(
						h.Class("form-input share-link-input"),
						g.Attr("id", "share-link-input"),
						g.Attr("name", "share-link"),
						g.Attr("type", "text"),
						g.Attr("readonly", "readonly"),
						g.Attr("placeholder", "Create share to generate link"),
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
			),
			h.Div(
				h.Class("share-state-banner"),
				h.Div(
					h.Class("share-state-banner-icon"),
					lucide.Info(
						h.Class("share-modal-lucide share-modal-lucide-accent"),
						g.Attr("aria-hidden", "true"),
					),
				),
				h.Div(
					h.Class("share-state-banner-copy"),
					h.P(h.Class("share-status-value"), g.Attr("id", "share-status"), g.Text("Ready to create")),
					h.P(h.Class("share-state-note"), g.Attr("id", "share-state-note"), g.Text("Create link, then configure how it should behave.")),
					h.Span(h.Class("share-save-state"), g.Attr("id", "share-save-state"), g.Text("")),
				),
			),
			h.Div(
				h.Class("share-config"),
				h.Div(
					h.Class("share-section-heading"),
					h.Span(g.Text("Link configuration")),
				),
				h.Div(
					h.Class("share-config-grid"),
					h.Div(
						h.Class("share-config-block"),
						h.Div(
							h.Class("share-config-label"),
							lucide.SlidersHorizontal(
								h.Class("share-modal-lucide share-setting-icon"),
								g.Attr("aria-hidden", "true"),
							),
							h.Span(g.Text("Link type")),
						),
						h.Div(
							h.Class("share-select-wrap"),
							h.Select(
								h.Class("form-input share-select"),
								g.Attr("id", "share-mode-picker"),
								h.Option(g.Attr("value", "stable"), g.Text("Stable link")),
								h.Option(g.Attr("value", "once"), g.Text("One-Time Access")),
							),
							lucide.ChevronDown(
								h.Class("share-modal-lucide share-select-icon"),
								g.Attr("aria-hidden", "true"),
							),
						),
						h.Div(
							h.Class("share-legacy-fields"),
							shareModeOption("stable", "Stable link", "Reusable until revoked or expired.", true),
							shareModeOption("once", "One-time link", "Consumed after first attempt to view or download.", false),
						),
					),
					h.Div(
						h.Class("share-config-block"),
						h.Div(
							h.Class("share-config-label"),
							lucide.CalendarDays(
								h.Class("share-modal-lucide share-setting-icon"),
								g.Attr("aria-hidden", "true"),
							),
							h.Span(g.Text("Expiry date")),
						),
						h.Div(
							h.Div(
								h.Class("share-expiry-shell"),
								h.Input(
									h.Class("form-input share-date-input"),
									g.Attr("id", "share-expiry-custom"),
									g.Attr("name", "share-expiry-custom"),
									g.Attr("type", "date"),
								),
								h.Input(
									h.Class("form-input share-legacy-fields"),
									g.Attr("id", "share-expiry-time"),
									g.Attr("name", "share-expiry-time"),
									g.Attr("type", "time"),
									g.Attr("value", "00:00"),
								),
								h.Input(
									h.Class("share-legacy-fields"),
									h.Type("checkbox"),
									g.Attr("id", "share-expiry-toggle"),
								),
								h.Select(
									h.Class("form-input share-legacy-fields"),
									g.Attr("id", "share-expiry-select"),
									h.Option(g.Attr("value", "custom"), g.Text("Custom date")),
								),
							),
						),
					),
				),
				h.Div(
					h.Class("share-access-card"),
					h.Div(
						h.Class("share-config-label"),
						lucide.KeyRound(
							h.Class("share-modal-lucide share-setting-icon"),
							g.Attr("aria-hidden", "true"),
						),
						h.Span(g.Text("Password protection")),
					),
					h.Div(
						h.Class("share-password-wrap"),
						h.Input(
							h.Class("form-input"),
							g.Attr("id", "share-password"),
							g.Attr("name", "share-password"),
							g.Attr("type", "password"),
							g.Attr("placeholder", "Create strong password"),
							g.Attr("autocomplete", "new-password"),
						),
						h.Button(
							h.Class("share-password-toggle"),
							h.Type("button"),
							g.Attr("id", "share-password-visibility"),
							g.Attr("aria-label", "Show password"),
							lucide.Eye(
								h.Class("share-modal-lucide"),
								g.Attr("aria-hidden", "true"),
							),
						),
						h.Input(
							h.Class("share-legacy-fields"),
							h.Type("checkbox"),
							g.Attr("id", "share-password-toggle"),
						),
					),
					h.Div(
						h.Class("share-password-field share-password-field-visible"),
						h.P(
							h.Class("share-password-helper"),
							g.Attr("id", "share-password-helper"),
							g.Text("Use at least 8 characters with lowercase, uppercase, and symbol."),
						),
						h.P(
							h.Class("share-password-strength"),
							g.Attr("id", "share-password-strength"),
							g.Text(""),
						),
					),
				),
			),
			h.P(h.Class("form-error share-error"), g.Attr("id", "share-error"), g.Text("")),
		),
		ActionsClass: "share-dialog-actions",
		Actions: g.Group([]g.Node{
			h.Div(
				h.Class("share-dialog-actions-row share-dialog-actions-top"),
				components.Button(components.ButtonProps{
					Text:    "Revoke Link",
					Variant: "secondary",
					Class:   "share-footer-action share-footer-link share-footer-danger",
					ID:      "share-revoke-button",
					Type:    "button",
				}),
				components.Button(components.ButtonProps{
					Text:    "Delete Link",
					Variant: "danger-outline",
					Class:   "share-footer-action",
					ID:      "share-delete-button",
					Type:    "button",
					Icon:    "trash",
				}),
			),
			h.Div(
				h.Class("share-dialog-actions-row share-dialog-actions-bottom"),
				components.Button(components.ButtonProps{
					Text:    "Cancel",
					Variant: "secondary",
					Class:   "share-footer-action",
					ID:      "share-cancel-button",
					Type:    "button",
				}),
				components.Button(components.ButtonProps{
					Text:     "Save Link",
					Variant:  "primary",
					Class:    "share-footer-action share-footer-primary",
					ID:       "share-save-button",
					Type:     "button",
					BusyText: "Saving...",
				}),
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
