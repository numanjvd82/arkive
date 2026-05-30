package pages

import (
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type ForgotPasswordPageProps struct {
	Ctx PageContext
}

type ResetPasswordPageProps struct {
	Ctx PageContext
}

func ForgotPasswordPage(props ForgotPasswordPageProps) web.Page {
	return web.Page{
		Title:    "Arkive · Forgot Password",
		Robots:   RobotsNoIndex,
		CSS:      []string{"/web/pages/login.css"},
		ModuleJS: []string{"/static/features/password_reset.js"},
		Body:     forgotPasswordBody(props),
		HideNav:  true,
	}
}

func ResetPasswordPage(props ResetPasswordPageProps) web.Page {
	return web.Page{
		Title:    "Arkive · Reset Password",
		Robots:   RobotsNoIndex,
		CSS:      []string{"/web/pages/login.css"},
		ModuleJS: []string{"/static/features/password_reset.js"},
		Body:     resetPasswordBody(props),
		HideNav:  true,
	}
}

func forgotPasswordBody(props ForgotPasswordPageProps) g.Node {
	return authResetShell(
		"Request password reset",
		"Create password reset token. Temporary until Arkive mail delivery returns.",
		h.Form(
			h.Class("auth-form"),
			g.Attr("data-forgot-password-form", "true"),
			components.InputField(components.InputProps{
				Label:       "Email",
				Name:        "email",
				Type:        components.InputTypeEmail,
				Placeholder: "user@local.node",
				Required:    true,
				InputClass:  "form-input-auth mono",
			}),
			h.P(h.Class("auth-message"), g.Attr("data-forgot-password-result", "true"), g.Attr("hidden", "hidden")),
			components.Button(components.ButtonProps{
				Text:     "Create Reset Link",
				Type:     "submit",
				Variant:  "primary",
				Class:    "auth-submit auth-submit-login",
				Icon:     "key",
				BusyText: "Creating...",
			}),
		),
	)
}

func resetPasswordBody(props ResetPasswordPageProps) g.Node {
	return authResetShell(
		"Reset vault password",
		"Recovery key unwraps existing master key. New password re-wraps same vault.",
		h.Form(
			h.Class("auth-form"),
			g.Attr("data-reset-password-form", "true"),
			components.InputField(components.InputProps{
				Label:       "Recovery key",
				Name:        "recovery_key",
				Type:        components.InputTypeText,
				Placeholder: "ARK-RK1-...",
				Required:    true,
				InputClass:  "form-input-auth mono",
			}),
			components.InputField(components.InputProps{
				Label:       "New password",
				Name:        "new_password",
				Type:        components.InputTypePassword,
				Placeholder: "Enter new password",
				Required:    true,
				InputClass:  "form-input-auth mono",
			}),
			components.InputField(components.InputProps{
				Label:       "Confirm password",
				Name:        "confirm_password",
				Type:        components.InputTypePassword,
				Placeholder: "Confirm new password",
				Required:    true,
				InputClass:  "form-input-auth mono",
			}),
			h.P(h.Class("auth-message"), g.Attr("data-reset-password-result", "true"), g.Attr("hidden", "hidden")),
			components.Button(components.ButtonProps{
				Text:     "Reset Password",
				Type:     "submit",
				Variant:  "primary",
				Class:    "auth-submit auth-submit-login",
				Icon:     "shield",
				BusyText: "Resetting...",
			}),
		),
	)
}

func authResetShell(title, subtitle string, form g.Node) g.Node {
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
					h.H1(g.Text(strings.TrimSpace(title))),
					h.P(
						h.Class("auth-subtitle"),
						g.Text(strings.TrimSpace(subtitle)),
					),
				),
				components.Card(components.CardProps{
					Class: "auth-card auth-card-login",
					Body:  []g.Node{form},
				}),
			),
		),
	)
}
