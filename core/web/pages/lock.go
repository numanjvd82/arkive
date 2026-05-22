package pages

import (
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type LockPageProps struct {
	Ctx  PageContext
	Next string
}

func LockPage(props LockPageProps) web.Page {
	return web.Page{
		Title:              "Arkive · Lock",
		Robots:             RobotsNoIndex,
		CSS:                []string{"/web/pages/login.css"},
		ModuleJS:           []string{"/static/lock.js"},
		Body:               lockBody(props),
		HideNav:            true,
		RequireVaultUnlock: false,
	}
}

func lockBody(props LockPageProps) g.Node {
	email := ""
	if props.Ctx.User != nil {
		email = strings.TrimSpace(props.Ctx.User.Email)
	}

	return h.Div(
		h.Class("auth-page"),
		h.Div(
			h.Class("auth-container"),
			h.Div(
				h.Class("auth-shell"),
				h.Header(
					h.Class("auth-header"),
					h.Span(
						h.Class("auth-mark"),
						lucide.Lock(
							h.Class("auth-lucide auth-lucide-lock"),
							g.Attr("aria-hidden", "true"),
						),
					),
					h.H1(g.Text("Arkive Vault Locked")),
					h.P(
						h.Class("auth-subtitle"),
						g.Text("Your login session is still active. Enter your password to unlock the vault again."),
					),
					g.If(email != "",
						h.P(
							h.Class("auth-message"),
							g.Text("Signed in as "+email),
						),
					),
				),
				components.Card(components.CardProps{
					Class: "auth-card auth-card-login",
					Body: []g.Node{
						h.Form(
							h.Class("auth-form"),
							g.Attr("data-lock-form", "true"),
							g.Attr("data-lock-next", strings.TrimSpace(props.Next)),
							components.InputField(components.InputProps{
								Label:       "Password",
								Name:        "password",
								Type:        components.InputTypePassword,
								Placeholder: "Enter your password",
								Required:    true,
								InputClass:  "form-input-auth mono",
								LabelSuffix: h.Span(
									h.Class("form-label-suffix"),
									lucide.Shield(
										h.Class("auth-lucide auth-lucide-shield"),
										g.Attr("aria-hidden", "true"),
									),
								),
							}),
							components.Button(components.ButtonProps{
								Text:     "Unlock Vault",
								Type:     "submit",
								Variant:  "primary",
								Class:    "auth-submit auth-submit-login",
								Icon:     "key",
								BusyText: "Unlocking...",
							}),
						),
					},
				}),
				h.Footer(
					h.Class("auth-status"),
					h.Span(h.Class("auth-status-dot")),
					h.Span(g.Text("Vault access paused.")),
				),
			),
		),
	)
}
