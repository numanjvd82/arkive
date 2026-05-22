package pages

import (
	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
)

type PublicSharePasswordProps struct {
	Token   string
	File    models.File
	Message string
}

func PublicSharePassword(props PublicSharePasswordProps) web.Page {
	return web.Page{
		Title:    "Arkive · Shared file",
		Robots:   RobotsNoIndex,
		CSS:      []string{"/web/pages/share.css"},
		ModuleJS: []string{"/static/share_password.js"},
		HideNav:  true,
		Body: g.Group([]g.Node{
			components.InlineStyle(components.InputCSS),
			h.Main(
				h.Class("share-page share-gate-page"),
				h.Div(
					h.Class("share-gate-brand"),
					h.Span(h.Class("share-gate-brand-text"), g.Text("Arkive Core")),
				),
				h.Div(
					h.Class("share-card share-gate-card"),
					h.Div(h.Class("share-gate-texture")),
					h.Div(
						h.Class("share-gate-header"),
						h.Div(
							h.Class("share-gate-file-icon-wrap"),
							lucide.Image(
								h.Class("share-gate-file-icon"),
								g.Attr("aria-hidden", "true"),
							),
						),
						h.H1(
							h.Class("share-gate-file-name"),
							g.Text(publicSharePasswordName()),
						),
						h.Div(
							h.Class("share-gate-pill"),
							lucide.Shield(
								h.Class("share-gate-pill-icon"),
								g.Attr("aria-hidden", "true"),
							),
							h.Span(g.Text("End-to-End Encrypted")),
						),
					),
					h.Form(
						h.Class("share-form share-gate-form"),
						h.Method("POST"),
						h.Action("/s/"+props.Token),
						h.P(
							h.Class("share-gate-label-row"),
							h.Span(h.Class("share-gate-label"), g.Text("Shared Key / Password")),
							h.Span(h.Class("share-gate-required"), g.Text("Required")),
						),
						h.Div(
							h.Class("share-gate-password-wrap"),
							h.Div(
								h.Class("password-wrapper share-gate-password-shell"),
								h.Input(
									h.Class("form-input password-input share-gate-password-input"),
									g.Attr("id", "share-password"),
									g.Attr("name", "password"),
									g.Attr("type", "password"),
									g.Attr("placeholder", "Enter the shared password"),
									g.Attr("required", "required"),
								),
								h.Button(
									h.Class("password-toggle share-gate-password-toggle"),
									g.Attr("type", "button"),
									g.Attr("aria-label", "Show password"),
									g.Attr("aria-pressed", "false"),
									g.Attr("data-target", "share-password"),
									g.Attr("data-visible", "false"),
									h.Span(h.Class("icon-eye"),
										lucide.KeyRound(
											h.Class("share-gate-password-icon"),
											g.Attr("aria-hidden", "true"),
										),
									),
									h.Span(
										h.Class("icon-eye-off"),
										lucide.KeyRound(
											h.Class("share-gate-password-icon"),
											g.Attr("aria-hidden", "true"),
										),
									),
								),
							),
						),
						h.Button(
							h.Class("button primary share-submit share-gate-submit"),
							h.Type("submit"),
							g.Attr("data-busy-text", "Unlocking..."),
							h.Span(g.Text("Unlock Secure Share")),
							lucide.ArrowRight(
								h.Class("share-gate-submit-icon"),
								g.Attr("aria-hidden", "true"),
							),
						),
						g.If(props.Message != "", h.P(
							h.Class("form-error share-gate-error"),
							g.Text(props.Message),
						)),
						h.Div(
							h.Class("share-gate-footnote"),
							lucide.Info(
								h.Class("share-gate-footnote-icon"),
								g.Attr("aria-hidden", "true"),
							),
							h.P(
								g.Text("Your password is verified by Arkive Core before access is granted. File decryption still happens in your browser."),
							),
						),
					),
					h.Div(
						h.Class("share-gate-meta"),
						h.Div(
							h.Class("share-gate-meta-item"),
							h.Span(h.Class("share-gate-meta-label"), g.Text("Size")),
							h.Span(h.Class("share-gate-meta-value"), g.Text(publicSharePasswordSize(props.File))),
						),
						h.Div(
							h.Class("share-gate-meta-item share-gate-meta-item-right"),
							h.Span(h.Class("share-gate-meta-label"), g.Text("Origin Node")),
							h.Span(h.Class("share-gate-meta-link"), g.Text("Arkive Core")),
						),
					),
				),
				h.Div(
					h.Class("share-gate-footer"),
					h.P(
						h.Class("share-gate-footer-copy"),
						g.Text("Powered by "),
						h.Span(h.Class("share-gate-footer-strong"), g.Text("Arkive Core Protocol")),
					),
					lucide.BadgeCheck(
						h.Class("share-gate-footer-icon"),
						g.Attr("aria-hidden", "true"),
					),
				),
			),
		}),
	}
}

func publicSharePasswordName() string {
	return "Encrypted file"
}

func publicSharePasswordSize(file models.File) string {
	if file.PlaintextSize <= 0 {
		return "Unknown"
	}
	return format.Bytes(file.PlaintextSize)
}
