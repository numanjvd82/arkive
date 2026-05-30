package pages

import (
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/validation"
)

type LoginPageProps struct {
	Ctx     PageContext
	Errors  validation.Errors
	Message string
	Email   string
}

func LoginPage(props LoginPageProps) web.Page {
	return web.Page{
		Title:   "Arkive · Admin Login",
		Robots:  RobotsNoIndex,
		CSS:     []string{"/web/pages/login.css"},
		ModuleJS: []string{"/static/login.js"},
		Body:    loginBody(props),
		HideNav: true,
	}
}

func loginBody(props LoginPageProps) g.Node {
	generalError := validation.FieldError(props.Errors, validation.GeneralKey)
	message := strings.TrimSpace(props.Message)

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
					h.H1(g.Text("Arkive Core")),
					h.P(
						h.Class("auth-subtitle"),
						g.Text("Sign in to your self-hosted encrypted file server."),
					),
				),
				components.Card(components.CardProps{
					Class: "auth-card auth-card-login",
					Body: []g.Node{
						h.Form(
							h.Class("auth-form"),
							g.Attr("method", "POST"),
							g.Attr("data-login-form", "true"),
							g.If(
								message != "",
								h.P(
									h.Class("auth-message"),
									g.Text(message),
								),
							),
							g.If(
								generalError != "",
								h.P(
									h.Class("form-error"),
									g.Text(generalError),
								),
							),
							components.InputField(components.InputProps{
								Label:       "Email",
								Name:        "email",
								Type:        components.InputTypeEmail,
								Placeholder: "user@local.node",
								Value:       props.Email,
								Required:    true,
								HelperText:  validation.FieldError(props.Errors, "email"),
								HasError:    validation.FieldError(props.Errors, "email") != "",
								InputClass:  "form-input-auth mono",
							}),
							components.InputField(components.InputProps{
								Label:       "Password",
								Name:        "password",
								Type:        components.InputTypePassword,
								Placeholder: "Enter your password",
								Required:    true,
								HelperText:  validation.FieldError(props.Errors, "password"),
								HasError:    validation.FieldError(props.Errors, "password") != "",
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
							h.P(
								h.Class("auth-message"),
								h.A(
									h.Href("/forgot-password"),
									g.Text("Forgot password?"),
								),
							),
						),
					},
				}),
				h.Footer(
					h.Class("auth-status"),
					h.Span(h.Class("auth-status-dot")),
					h.Span(g.Text("System standing by.")),
				),
			),
		),
	)
}
