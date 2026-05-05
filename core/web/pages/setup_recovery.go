package pages

import (
	"fmt"
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type SetupRecoveryPageProps struct {
	Ctx          PageContext
	Words        []string
	Error        string
	Acknowledged bool
}

func SetupRecoveryPage(props SetupRecoveryPageProps) web.Page {
	return web.Page{
		Title:   "Arkive · Vault Recovery Key",
		Robots:  RobotsNoIndex,
		CSS:     []string{"/web/pages/setup_recovery.css"},
		JS:      []string{"/static/setup_recovery.js"},
		Body:    setupRecoveryBody(props),
		HideNav: true,
	}
}

func setupRecoveryBody(props SetupRecoveryPageProps) g.Node {
	phrase := strings.Join(props.Words, " ")

	return h.Div(
		h.Class("recovery-page"),
		h.Main(
			h.Class("recovery-main"),
			h.Div(h.Class("recovery-grain")),
			h.Div(
				h.Class("recovery-shell"),
				h.Section(
					h.Class("recovery-panel"),
					h.Div(
						h.Class("recovery-header"),
						h.Span(
							h.Class("recovery-header-alert"),
							lucide.TriangleAlert(
								h.Class("recovery-alert-icon"),
								g.Attr("aria-hidden", "true"),
							),
						),
						h.Div(
							h.Class("recovery-header-copy"),
							h.H1(g.Text("Vault Recovery Key")),
							h.P(
								g.Text("This is the "),
								h.Strong(h.Class("recovery-danger-emphasis"), g.Text("ONLY")),
								g.Text(" way to recover your data if you lose your password. Arkive Core is a zero-knowledge system; we cannot reset your access."),
							),
						),
					),
					h.Form(
						h.Class("recovery-form"),
						g.Attr("method", "POST"),
						g.Attr("action", "/setup/recovery-key"),
						h.Section(
							h.Class("recovery-key-shell"),
							h.Div(
								h.Class("recovery-key-actions"),
								components.CopyButton(components.CopyButtonProps{
									Text:           "",
									Icon:           "copy",
									Value:          phrase,
									Variant:        "secondary",
									Class:          "recovery-copy-button",
									AriaLabel:      "Copy recovery key",
									SuccessTitle:   "Copied",
									SuccessMessage: "Recovery key copied.",
								}),
								h.Button(
									h.Type("button"),
									h.Class("button secondary recovery-utility-button"),
									g.Attr("data-recovery-print", "true"),
									g.Attr("aria-label", "Print recovery key"),
									lucide.Printer(
										h.Class("button-lucide"),
										g.Attr("aria-hidden", "true"),
									),
								),
							),
							h.Div(
								h.Class("recovery-key-grid"),
								g.Group(recoveryWordNodes(props.Words)),
							),
						),
						h.Div(
							h.Class("recovery-divider"),
							h.Div(h.Class("recovery-divider-bar")),
						),
						g.If(
							strings.TrimSpace(props.Error) != "",
							h.P(
								h.Class("form-error recovery-form-error"),
								g.Text(props.Error),
							),
						),
						h.Label(
							h.Class("recovery-confirm"),
							h.For("confirm-backup"),
							h.Input(
								h.Type("checkbox"),
								h.ID("confirm-backup"),
								h.Name("confirm_backup"),
								g.Attr("value", "true"),
								g.If(props.Acknowledged, g.Attr("checked", "checked")),
							),
							h.Span(
								h.Class("recovery-confirm-copy"),
								g.Text("I have written down or otherwise securely backed up this 24-word recovery phrase. I understand that Arkive Core cannot recover this key for me and losing it will result in "),
								h.Strong(g.Text("permanent data loss")),
								g.Text("."),
							),
						),
						h.Div(
							h.Class("recovery-actions"),
							h.Button(
								h.Type("submit"),
								h.Class("button primary recovery-submit"),
								g.Attr("data-recovery-submit", "true"),
								g.Attr("disabled", "disabled"),
								h.Span(g.Text("Confirm & Secure Vault")),
								lucide.ShieldCheck(
									h.Class("button-lucide"),
									g.Attr("aria-hidden", "true"),
								),
							),
							h.Button(
								h.Type("button"),
								h.Class("button secondary recovery-download-button"),
								g.Attr("data-recovery-download", "true"),
								lucide.Download(
									h.Class("button-lucide"),
									g.Attr("aria-hidden", "true"),
								),
								h.Span(g.Text("Download Recovery PDF")),
							),
						),
					),
				),
			),
		),
	)
}

func recoveryWordNodes(words []string) []g.Node {
	nodes := make([]g.Node, 0, len(words))
	for i, word := range words {
		nodes = append(nodes, h.Div(
			h.Class("recovery-word"),
			h.Span(
				h.Class("recovery-word-index"),
				g.Text(fmt.Sprintf("%02d", i+1)),
			),
			h.Span(
				h.Class("recovery-word-value"),
				g.Text(strings.TrimSpace(word)),
			),
		))
	}
	return nodes
}
