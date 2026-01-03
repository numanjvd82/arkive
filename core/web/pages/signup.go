package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/validation"
)

type SignupPageProps struct {
	Ctx            PageContext
	Errors         validation.Errors
	BrandName      string
	Email          string
	GoogleClientID string
}

func SignupPage(props SignupPageProps) web.Page {
	js := []string{}
	if props.GoogleClientID != "" {
		js = append(js, "/static/auth_google.js")
	}
	return web.Page{
		Title: "Arkive · Sign Up",
		CSS:   []string{"/web/pages/signup.css"},
		JS:    js,
		Body:  signUpBody(props),
	}
}

func signUpBody(props SignupPageProps) g.Node {
	generalError := validation.FieldError(props.Errors, validation.GeneralKey)

	return h.Div(
		h.Class("auth-page"),
		h.Div(
			h.Class("auth-container"),
			components.Card(components.CardProps{
				Title:    "Sign up",
				Subtitle: "Create your Arkive workspace.",
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
							Label:       "Brand name",
							Name:        "brand_name",
							Type:        components.InputTypeText,
							Placeholder: "Studio Arkive",
							Value:       props.BrandName,
							Required:    true,
							HelperText:  validation.FieldError(props.Errors, "brand_name"),
							HasError:    validation.FieldError(props.Errors, "brand_name") != "",
						}),
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
							Placeholder: "Create a password",
							Required:    true,
							HelperText:  validation.FieldError(props.Errors, "password"),
							HasError:    validation.FieldError(props.Errors, "password") != "",
						}),
						components.InputField(components.InputProps{
							Label:       "Confirm password",
							Name:        "confirm_password",
							Type:        components.InputTypePassword,
							Placeholder: "Confirm your password",
							Required:    true,
							HelperText:  validation.FieldError(props.Errors, "confirm_password"),
							HasError:    validation.FieldError(props.Errors, "confirm_password") != "",
						}),
						components.Button(components.ButtonProps{
							Text:    "Create account",
							Type:    "submit",
							Variant: "primary",
							Class:   "auth-submit",
						}),
					),
					h.P(
						h.Class("auth-alt"),
						g.Text("Already have an account? "),
						h.A(h.Class("text-link"), h.Href("/login"), g.Text("Log in")),
					),
					googleAuthSection(props.GoogleClientID),
				},
			}),
		),
	)
}
