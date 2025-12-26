package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/validation"
)

type LoginPageProps struct {
	Ctx    PageContext
	Errors validation.Errors
	Email  string
}

func LoginPage(props LoginPageProps) web.Page {
	return web.Page{
		Title: "Arkive · Login",
		CSS:   []string{"/web/pages/login.css"},
		Body:  loginBody(props),
	}
}

func loginBody(props LoginPageProps) g.Node {
	generalError := validation.FieldError(props.Errors, validation.GeneralKey)

	return h.Div(
		h.Class("auth-page"),
		h.Div(
			h.Class("auth-container"),
			components.Card(components.CardProps{
				Title:    "Login",
				Subtitle: "Welcome back to Arkive.",
				Class:    "auth-card",
				Body: []g.Node{
					h.Form(
						h.Class("auth-form"),
						g.Attr("method", "POST"),
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
							Placeholder: "you@example.com",
							Value:       props.Email,
							Required:    true,
							HelperText:  validation.FieldError(props.Errors, "email"),
							HasError:    validation.FieldError(props.Errors, "email") != "",
						}),
						components.InputField(components.InputProps{
							Label:       "Password",
							Name:        "password",
							Type:        components.InputTypePassword,
							Placeholder: "Enter your password",
							Required:    true,
							HelperText:  validation.FieldError(props.Errors, "password"),
							HasError:    validation.FieldError(props.Errors, "password") != "",
						}),
						components.Button(components.ButtonProps{
							Text:    "Login",
							Type:    "submit",
							Variant: "primary",
							Class:   "auth-submit",
						}),
					),
				},
			}),
		),
	)
}
