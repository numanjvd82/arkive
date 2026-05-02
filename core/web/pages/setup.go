package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/validation"
)

type SetupPageProps struct {
	Ctx       PageContext
	Errors    validation.Errors
	BrandName string
	Email     string
}

func SetupPage(props SetupPageProps) web.Page {
	return web.Page{
		Title:   "Arkive · Setup",
		Robots:  RobotsNoIndex,
		CSS:     []string{"/web/pages/setup.css"},
		Body:    setupBody(props),
		HideNav: true,
	}
}

func setupBody(props SetupPageProps) g.Node {
	generalError := validation.FieldError(props.Errors, validation.GeneralKey)

	return h.Div(
		h.Class("auth-page"),
		h.Div(
			h.Class("auth-container"),
			components.Card(components.CardProps{
				Title:    "Set up Arkive",
				Subtitle: "Create the admin account for this instance.",
				Class:    "auth-card",
				Body: []g.Node{
					h.P(
						h.Class("auth-helper"),
						g.Text("This account owns the self-hosted instance."),
					),
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
							Label:       "Instance name",
							Name:        "brand_name",
							Type:        components.InputTypeText,
							Placeholder: "Arkive",
							Value:       props.BrandName,
							Required:    true,
							HelperText:  validation.FieldError(props.Errors, "brand_name"),
							HasError:    validation.FieldError(props.Errors, "brand_name") != "",
						}),
						components.InputField(components.InputProps{
							Label:       "Admin email",
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
							Text:    "Create admin account",
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
