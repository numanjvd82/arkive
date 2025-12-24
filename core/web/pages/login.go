package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

func LoginPage() web.Page {
	return web.Page{
		Title: "Arkive · Login",
		CSS:   []string{"/web/pages/login.css"},
		Body:  loginBody(),
	}
}

func loginBody() g.Node {
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
						components.InputField(components.InputProps{
							Label:       "Email",
							Name:        "email",
							Type:        components.InputTypeEmail,
							Placeholder: "you@example.com",
							Required:    true,
						}),
						components.InputField(components.InputProps{
							Label:       "Password",
							Name:        "password",
							Type:        components.InputTypePassword,
							Placeholder: "Enter your password",
							Required:    true,
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
