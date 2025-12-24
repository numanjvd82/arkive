package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

func SignupPage() web.Page {
	return web.Page{
		Title: "Arkive · Sign Up",
		CSS:   []string{"/web/pages/signup.css"},
		Body:  signUpBody(),
	}
}

func signUpBody() g.Node {
	return h.Div(
		h.Class("auth-page"),
		h.Div(
			h.Class("auth-container"),
			h.Form(
				h.Class("auth-form"),
				g.Attr("method", "POST"),
				h.H2(g.Text("Login")),
				components.InputField(components.InputProps{
					Label:       "Email",
					Name:        "email",
					Type:        "email",
					Placeholder: "you@example.com",
					Required:    true,
				}),
				components.InputField(components.InputProps{
					Label:       "Password",
					Name:        "password",
					Type:        "password",
					Placeholder: "Create a password",
					Required:    true,
				}),
				h.Button(
					g.Attr("type", "submit"),
					g.Text("Login"),
				),
			),
		),
	)
}
